#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import sys
from datetime import date, timedelta
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import HTTPRedirectHandler, Request, build_opener

BASE = "https://external-api-service.pagamerican.app/api/v1"
ROUTES = ("purchases", "orders", "refunds", "chargebacks", "checkouts", "abandoned-carts")

class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, request, fp, code, msg, headers, newurl):
        return None


def inspect(route, day, key, open_url=None):
    if route not in ROUTES:
        raise ValueError("rota inválida")
    start = date.fromisoformat(day)
    if not key:
        raise ValueError("PAGAMERICAN_API_KEY ausente")
    if any(ord(ch) < 32 or ord(ch) == 127 for ch in key):
        raise ValueError("credencial inválida")
    url = f"{BASE}/{route}?" + urlencode({"from": start.isoformat(), "to": (start + timedelta(days=1)).isoformat(), "page": 1, "limit": 100})
    request = Request(url, headers={"Authorization": f"Bearer {key}", "Accept": "application/json"}, method="GET")
    if open_url is None:
        open_url = build_opener(NoRedirect).open
    with open_url(request, timeout=15) as response:
        raw = response.read(1024 * 1024 + 1)
    if len(raw) > 1024 * 1024:
        raise ValueError("resposta grande demais")
    payload = json.loads(raw)
    if not isinstance(payload, dict):
        raise ValueError("schema inesperado")
    meta = payload.get("meta")
    if not isinstance(meta, dict):
        raise ValueError("metadados indisponíveis; confira contrato atual da API")
    return {field: meta[field] for field in ("total", "count", "truncated", "page", "totalPages") if field in meta and isinstance(meta[field], (int, bool))}


def self_check():
    from io import BytesIO

    class Response(BytesIO):
        def __enter__(self):
            return self

        def __exit__(self, *_):
            self.close()

    def fake(request, timeout):
        assert request.get_method() == "GET" and timeout == 15
        assert request.full_url.startswith(BASE + "/purchases?")
        assert "limit=100" in request.full_url and "to=2026-10-02" in request.full_url
        return Response(b'{"meta":{"count":1,"truncated":false},"data":[{"email":"private@example.com"}]}')

    result = inspect("purchases", "2026-10-01", "fake", fake)
    assert result == {"count": 1, "truncated": False}
    assert NoRedirect().redirect_request(None, None, 302, None, None, "https://other.invalid/") is None
    def oversized(request, timeout):
        return Response(b" " * (1024 * 1024 + 1))
    try:
        inspect("purchases", "2026-10-01", "fake", oversized)
        raise AssertionError("resposta grande aceita")
    except ValueError:
        pass
    def invalid(request, timeout):
        return Response(b"[]")
    try:
        inspect("purchases", "2026-10-01", "fake", invalid)
        raise AssertionError("schema inválido aceito")
    except ValueError:
        pass
    try:
        inspect("purchases", "2026-10-01", "fake-secret\r\nInjected", fake)
        raise AssertionError("credencial com controle aceita")
    except ValueError as error:
        assert "fake-secret" not in str(error)
    process = subprocess.run(
        [sys.executable, __file__, "purchases", "2026-10-01"],
        env={**os.environ, "PAGAMERICAN_API_KEY": "fake-secret\r\nInjected"},
        capture_output=True, text=True, timeout=5,
    )
    assert process.returncode != 0 and "fake-secret" not in process.stdout + process.stderr
    print("self-check OK")


def main():
    parser = argparse.ArgumentParser(description="Consulta metadados PagAmerican sem imprimir registros")
    parser.add_argument("route", nargs="?", choices=ROUTES)
    parser.add_argument("day", nargs="?")
    parser.add_argument("--self-check", action="store_true")
    args = parser.parse_args()
    if args.self_check:
        self_check()
        return
    if not args.route or not args.day:
        parser.error("informe rota e dia YYYY-MM-DD")
    try:
        result = inspect(args.route, args.day, os.environ.get("PAGAMERICAN_API_KEY"))
        print(json.dumps(result, ensure_ascii=False))
    except (ValueError, HTTPError, URLError) as error:
        print(f"consulta falhou: {type(error).__name__}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
