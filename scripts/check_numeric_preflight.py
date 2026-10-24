"""Check numeric preflight through the running proxy without telemetry reads/writes."""
import json
import urllib.error
import urllib.request


def main():
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    url = "http://127.0.0.1:8088/api/telemetry-packages/numeric-plan"
    question = {"signal": "logs", "environment": "numeric-preflight",
                "fromMs": 1000, "toMs": 61000, "limit": 100, "offset": 0}
    binding = {"field": {"kind": "attribute", "key": "db.response.returned_rows"},
               "type": "uint64", "unit": "{row}", "parse": "canonical-decimal"}

    def call(payload):
        request = urllib.request.Request(url, data=json.dumps(payload).encode(),
                                         headers={"Content-Type": "application/json"})
        with opener.open(request, timeout=10) as response:
            assert response.status == 200
            return json.load(response)

    plan = call({"question": question, "binding": binding})
    assert plan["executable"] is True
    assert plan["binding"] == binding
    assert plan["scope"]["signal"] == "logs"
    assert plan["scope"]["environment"] == question["environment"]
    assert plan["meanDenominator"] == "valid"
    assert plan["coverage"] == ["valid", "missing", "empty", "invalid"]
    assert [d["code"] for d in plan["diagnostics"]] == []
    decimal = {**binding, "type": "decimal64", "scale": 2}
    decimal_plan = call({"question": question, "binding": decimal})
    assert decimal_plan["executable"] is True
    assert decimal_plan["binding"] == decimal
    for scale in (0, -1, 10):
        rejected = call({"question": question, "binding": {**decimal, "scale": scale}})
        assert rejected["executable"] is False
        assert any(d["path"] == "binding.scale" for d in rejected["diagnostics"])
    bad = {**binding, "type": "float64", "parse": "or-zero", "unit": ""}
    unsupported = call({"question": question, "binding": bad})
    assert unsupported["executable"] is False
    assert [d["code"] for d in unsupported["diagnostics"]] == [
        "unsupported_type", "unsupported_parse", "invalid_unit"]
    try:
        call({"question": question, "binding": binding, "sql": "SELECT 1"})
    except urllib.error.HTTPError as error:
        assert error.code == 400
    else:
        raise AssertionError("unknown request fields accepted")
    print(json.dumps({"validPlan": plan, "unsupportedPlan": unsupported}, indent=2))
    print("PASS: integer/decimal preflight, scale diagnostics and strict proxy decoding")


if __name__ == "__main__":
    main()
