package optionalmetric

import (
	"encoding/json"
	"testing"
)

func TestDecodeExists379(t *testing.T) {
	for _, s := range []string{`{"result":0}`, `{"result":1}`} {
		present, err := DecodeExists([]json.RawMessage{json.RawMessage(s)})
		if err != nil || present != (s == `{"result":1}`) {
			t.Fatal(s, present, err)
		}
	}
	for _, s := range []string{`{}`, `null`, `{"result":-0}`, `{"result":1.0}`, `{"result":1e0}`, `{"result":"1"}`, `{"result":true}`, `{"result":null}`, `{"result":2}`, `{"result":1,"other":0}`, `{"result":1,"result":0}`, `{"result":0,"result":1}`, `{"result":1} {"result":0}`} {
		if _, err := DecodeExists([]json.RawMessage{json.RawMessage(s)}); err == nil {
			t.Fatal("accepted", s)
		}
	}
	for _, rows := range [][]json.RawMessage{nil, {json.RawMessage(`{"result":1}`), json.RawMessage(`{"result":0}`)}} {
		if _, err := DecodeExists(rows); err == nil {
			t.Fatal("accepted row count")
		}
	}
}
