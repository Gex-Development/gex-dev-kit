#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import sys
from datetime import date
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import HTTPRedirectHandler, Request, build_opener

BASE = "https://admin3.buygoods.com/public/v2"

class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, request, fp, code, msg, headers, newurl):
        return None


def inspect(route, start, end, account, token, open_url=None):
    if route not in ("orders", "transactions"):
        raise ValueError("rota inválida")
    if date.fromisoformat(start) > date.fromisoformat(end):
        raise ValueError("início posterior ao fim")
    if not account or not token:
        raise ValueError("BUYGOODS_ACCOUNT_ID e BUYGOODS_API_TOKEN obrigatórios")
    if any(ord(ch) < 32 or ord(ch) == 127 for value in (account, token) for ch in value):
        raise ValueError("credencial inválida")
    url = f"{BASE}/{route}.php?" + urlencode({"account_id": account, "start_date": start, "end_date": end, "limit": 1, "offset": 0})
    request = Request(url, headers={"Authorization": f"Bearer {token}", "Accept": "application/json", "User-Agent": "Mozilla/5.0 Chrome/131.0.0.0 Safari/537.36"}, method="GET")
    if open_url is None:
        open_url = build_opener(NoRedirect).open
    with open_url(request, timeout=15) as response:
        raw = response.read(1024 * 1024 + 1)
    if len(raw) > 1024 * 1024:
        raise ValueError("resposta grande demais")
    payload = json.loads(raw)
    if not isinstance(payload, dict):
        raise ValueError("schema inesperado")
    if payload.get("success") is not True or not isinstance(payload.get("pagination"), dict):
        raise ValueError("API sem paginação válida; confira contrato atual")
    total = payload["pagination"].get("total")
    if not isinstance(total, int) or isinstance(total, bool) or total < 0:
        raise ValueError("total indisponível; confira contrato atual")
    return {"total": total}


def self_check():
    from io import BytesIO

    def fake(request, timeout):
        assert request.get_method() == "GET" and timeout == 15
        assert request.full_url.startswith(BASE + "/orders.php?")
        assert "limit=1" in request.full_url and "offset=0" in request.full_url
        return BytesIO(b'{"success":true,"pagination":{"total":7},"orders":[{"email":"private@example.com"}]}')

    assert inspect("orders", "2026-10-01", "2026-10-02", "synthetic", "fake", fake) == {"total": 7}
    assert NoRedirect().redirect_request(None, None, 302, None, None, "https://other.invalid/") is None
    def oversized(request, timeout):
        return BytesIO(b" " * (1024 * 1024 + 1))
    try:
        inspect("orders", "2026-10-01", "2026-10-02", "synthetic", "fake", oversized)
        raise AssertionError("resposta grande aceita")
    except ValueError:
        pass
    def invalid(request, timeout):
        return BytesIO(b"[]")
    try:
        inspect("orders", "2026-10-01", "2026-10-02", "synthetic", "fake", invalid)
        raise AssertionError("schema inválido aceito")
    except ValueError:
        pass
    try:
        inspect("orders", "2026-10-01", "2026-10-02", "synthetic", "fake-secret\r\nInjected", fake)
        raise AssertionError("credencial com controle aceita")
    except ValueError as error:
        assert "fake-secret" not in str(error)
    process = subprocess.run(
        [sys.executable, __file__, "orders", "2026-10-01", "2026-10-02"],
        env={**os.environ, "BUYGOODS_ACCOUNT_ID": "synthetic", "BUYGOODS_API_TOKEN": "fake-secret\r\nInjected"},
        capture_output=True, text=True, timeout=5,
    )
    assert process.returncode != 0 and "fake-secret" not in process.stdout + process.stderr
    print("self-check OK")


def main():
    parser = argparse.ArgumentParser(description="Metadados BuyGoods sem imprimir linhas")
    parser.add_argument("route", nargs="?", choices=("orders", "transactions"))
    parser.add_argument("start", nargs="?")
    parser.add_argument("end", nargs="?")
    parser.add_argument("--self-check", action="store_true")
    args = parser.parse_args()
    if args.self_check:
        self_check()
        return
    if not args.route or not args.start or not args.end:
        parser.error("informe rota e datas YYYY-MM-DD")
    try:
        result = inspect(args.route, args.start, args.end, os.environ.get("BUYGOODS_ACCOUNT_ID"), os.environ.get("BUYGOODS_API_TOKEN"))
        print(json.dumps(result))
    except (ValueError, HTTPError, URLError) as error:
        print(f"consulta falhou: {type(error).__name__}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
