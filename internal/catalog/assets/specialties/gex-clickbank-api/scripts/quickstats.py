#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import sys
from decimal import Decimal, InvalidOperation
from datetime import date
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import HTTPRedirectHandler, Request, build_opener

BASE = "https://api.clickbank.com/rest/1.3/quickstats/count"

class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, request, fp, code, msg, headers, newurl):
        return None


def aggregates(payload, requested_account):
    account = payload.get("accountData") if isinstance(payload, dict) else None
    if isinstance(account, list):
        if len(account) != 1:
            raise ValueError("resposta não limitada a uma conta")
        account = account[0]
    if isinstance(account, dict) and account.get("nickName") and account["nickName"] != requested_account:
        raise ValueError("API retornou outra conta")
    stats = account.get("quickStats") if isinstance(account, dict) else None
    if isinstance(stats, list):
        if len(stats) != 1:
            raise ValueError("resposta agregada inesperada")
        stats = stats[0]
    if not isinstance(stats, dict) or any(field not in stats for field in ("sale", "refund", "chargeback")):
        raise ValueError("schema Quickstats inesperado")
    try:
        amounts = {field: Decimal(str(stats[field])) for field in ("sale", "refund", "chargeback")}
    except (InvalidOperation, TypeError) as error:
        raise ValueError("valores Quickstats inválidos") from error
    if not all(amount.is_finite() for amount in amounts.values()):
        raise ValueError("valores Quickstats inválidos")
    return {field: str(value) for field, value in amounts.items()}


def inspect(start, end, account, key, open_url=None):
    if date.fromisoformat(start) > date.fromisoformat(end):
        raise ValueError("início posterior ao fim")
    if not account or not key:
        raise ValueError("CLICKBANK_ACCOUNT e CLICKBANK_API_KEY obrigatórios")
    if any(ord(ch) < 32 or ord(ch) == 127 for value in (account, key) for ch in value):
        raise ValueError("credencial inválida")
    url = BASE + "?" + urlencode({"startDate": start, "endDate": end, "account": account})
    request = Request(url, headers={"Authorization": key, "Accept": "application/json"}, method="GET")
    if open_url is None:
        open_url = build_opener(NoRedirect).open
    with open_url(request, timeout=15) as response:
        raw = response.read(1024 * 1024 + 1)
    if len(raw) > 1024 * 1024:
        raise ValueError("resposta grande demais")
    return aggregates(json.loads(raw), account)


def self_check():
    from io import BytesIO

    def fake(request, timeout):
        assert request.get_method() == "GET" and timeout == 15
        assert request.full_url.startswith(BASE + "?")
        assert "account=synthetic" in request.full_url
        return BytesIO(b'{"accountData":{"nickName":"synthetic","quickStats":{"sale":"10.00","refund":"2.00","chargeback":"0.00","email":"private@example.com"}}}')

    assert inspect("2026-10-01", "2026-10-02", "synthetic", "fake", fake) == {"sale": "10.00", "refund": "2.00", "chargeback": "0.00"}
    assert NoRedirect().redirect_request(None, None, 302, None, None, "https://other.invalid/") is None
    def oversized(request, timeout):
        return BytesIO(b" " * (1024 * 1024 + 1))
    try:
        inspect("2026-10-01", "2026-10-02", "synthetic", "fake", oversized)
        raise AssertionError("resposta grande aceita")
    except ValueError:
        pass
    try:
        aggregates({"accountData": {"quickStats": {"sale": "private@example.com", "refund": "0", "chargeback": "0"}}}, "synthetic")
        raise AssertionError("valor não numérico aceito")
    except ValueError:
        pass
    try:
        aggregates([], "synthetic")
        raise AssertionError("schema inválido aceito")
    except ValueError:
        pass
    try:
        inspect("2026-10-01", "2026-10-02", "synthetic", "fake-secret\r\nInjected", fake)
        raise AssertionError("credencial com controle aceita")
    except ValueError as error:
        assert "fake-secret" not in str(error)
    process = subprocess.run(
        [sys.executable, __file__, "2026-10-01", "2026-10-02"],
        env={**os.environ, "CLICKBANK_ACCOUNT": "synthetic", "CLICKBANK_API_KEY": "fake-secret\r\nInjected"},
        capture_output=True, text=True, timeout=5,
    )
    assert process.returncode != 0 and "fake-secret" not in process.stdout + process.stderr
    print("self-check OK")


def main():
    parser = argparse.ArgumentParser(description="Totais Quickstats sem imprimir dados de clientes")
    parser.add_argument("start", nargs="?")
    parser.add_argument("end", nargs="?")
    parser.add_argument("--self-check", action="store_true")
    args = parser.parse_args()
    if args.self_check:
        self_check()
        return
    if not args.start or not args.end:
        parser.error("informe início e fim YYYY-MM-DD")
    try:
        print(json.dumps(inspect(args.start, args.end, os.environ.get("CLICKBANK_ACCOUNT"), os.environ.get("CLICKBANK_API_KEY"))))
    except (ValueError, HTTPError, URLError) as error:
        print(f"consulta falhou: {type(error).__name__}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
