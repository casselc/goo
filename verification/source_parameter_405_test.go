package verification

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/internal/sourcehttp"
	"hegel.dev/go/hegel"
)

// This independent inverse accepts only the canonical Escaped spelling, after
// exactly one URL decode. It does not apply question admission or normalize UTF-8.
func decodeCanonicalParameter405(value string) ([]byte, error) {
	out := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < 32 || c == 127 {
			return nil, fmt.Errorf("raw control in canonical wire")
		}
		if c != '\\' {
			out = append(out, c)
			continue
		}
		i++
		if i >= len(value) {
			return nil, fmt.Errorf("torn escape")
		}
		switch value[i] {
		case '\\':
			out = append(out, '\\')
		case 'x':
			if i+2 >= len(value) {
				return nil, fmt.Errorf("torn hexadecimal escape")
			}
			pair := value[i+1 : i+3]
			if strings.ToLower(pair) != pair {
				return nil, fmt.Errorf("noncanonical hex")
			}
			decoded, err := hex.DecodeString(pair)
			if err != nil {
				return nil, err
			}
			if decoded[0] >= 32 && decoded[0] != 127 {
				return nil, fmt.Errorf("unnecessary hexadecimal escape")
			}
			out = append(out, decoded[0])
			i += 2
		default:
			return nil, fmt.Errorf("noncanonical escape")
		}
	}
	return out, nil
}

func TestHegelParameterEncoding405(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		length := hegel.Draw(tc, hegel.Integers(0, 512))
		generator := rand.New(rand.NewPCG(uint64(seed), 405))
		raw := make([]byte, length)
		for i := range raw {
			raw[i] = byte(generator.IntN(256))
		}
		original := bytes.Clone(raw)
		wire := sourcehttp.EncodeParameter(string(raw))
		query := url.Values{"param_literal": []string{wire}}.Encode()
		parsed, err := url.ParseQuery(query)
		if err != nil {
			tc.Fatal(err)
		}
		decoded, err := decodeCanonicalParameter405(parsed.Get("param_literal"))
		if err != nil || !bytes.Equal(decoded, original) || !bytes.Equal(raw, original) {
			tc.Fatalf("wire lost literal bytes: %v", err)
		}
		changed := append(bytes.Clone(original), '\\')
		if sourcehttp.EncodeParameter(string(changed)) == wire {
			tc.Fatal("distinct literal collapsed")
		}
	}, options(128)...)
}

// Freeze the pre-405 metric writer for semantic comparison, never as a fallback.
func legacyMetricParameter405(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\t", "\\t", "\r", "\\r", "\x00", "\\0").Replace(value)
}

type parameterProbe405 struct {
	Name, LiteralHex, OldWireHex, SharedWireHex string
	OldStatus, SharedStatus                     int
	OldDecodedHex, SharedDecodedHex             string
	OldError, SharedError                       string `json:",omitempty"`
	WireChanged, OldMatches, SharedMatches      bool
}

func probeParameter405(ctx context.Context, client *http.Client, endpoint, user, password, encoded string) (int, string, error) {
	target, err := url.Parse(endpoint)
	if err != nil {
		return 0, "", err
	}
	parameters := target.Query()
	for key, value := range map[string]string{"readonly": "1", "max_execution_time": "10", "max_result_rows": "1", "result_overflow_mode": "throw", "max_bytes_to_read": "268435456", "read_overflow_mode": "throw", "param_literal": encoded} {
		parameters.Set(key, value)
	}
	target.RawQuery = parameters.Encode()
	request, err := http.NewRequestWithContext(ctx, "POST", target.String(), strings.NewReader("SELECT hex({literal:String}) AS literal_hex FORMAT JSONEachRow"))
	if err != nil {
		return 0, "", err
	}
	request.SetBasicAuth(user, password)
	response, err := client.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return response.StatusCode, "", err
	}
	if response.StatusCode != http.StatusOK {
		return response.StatusCode, "", fmt.Errorf("String decoder HTTP%d", response.StatusCode)
	}
	var result map[string]json.RawMessage
	if err = json.Unmarshal(body, &result); err != nil {
		return response.StatusCode, "", err
	}
	var decoded string
	value, present := result["literal_hex"]
	if !present || len(result) != 1 || bytes.Equal(value, []byte("null")) || json.Unmarshal(value, &decoded) != nil {
		return response.StatusCode, "", fmt.Errorf("String decoder omitted exact hexadecimal value")
	}
	return response.StatusCode, decoded, nil
}

// The companion 393 acceptance test already executes the real record, metric
// attribute and exact-stream adapters with physical membership and budget oracles.
// Run that unchanged over the same fresh bank. This probe isolates whether the
// former metric spelling and the shared spelling differ semantically, rather
// than declaring every wire difference a production corruption bug.
func TestActualParameterEncodingDifferential405(t *testing.T) {
	path := os.Getenv("GOO_SOURCE_PARAMETER_405_BANK")
	if path == "" {
		t.Skip("fresh settled ordinary OTLP bank required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var bank literalBank393
	if err = json.Unmarshal(raw, &bank); err != nil {
		t.Fatal(err)
	}
	if err = admitLiteralBank393(bank); err != nil {
		t.Fatal(err)
	}
	endpoint := os.Getenv("GOO_SOURCE_PARAMETER_405_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("GOO_TEST_CLICKHOUSE_URL")
	}
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8123"
	}
	user, password := os.Getenv("GOO_TEST_CLICKHOUSE_USER"), os.Getenv("GOO_TEST_CLICKHOUSE_PASSWORD")
	if user == "" {
		user = "lab"
	}
	if password == "" {
		password = "lab-local"
	}
	client := &http.Client{Timeout: 15 * time.Second}
	cases := []parameterProbe405{}
	changed, oldFailures := 0, 0
	defer func() {
		report := map[string]any{"version": "goo.parameterEncoding405/1", "bank": path, "cases": cases, "wireDifferences": changed, "oldSemanticMismatches": oldFailures, "stringDecoderRequests": len(cases) * 2}
		encoded, _ := json.MarshalIndent(report, "", "  ")
		if err := artifact("parameter-encoding405-differential.json", encoded); err != nil {
			t.Error(err)
		}
	}()
	for _, vector := range bank.Vectors {
		old, shared := legacyMetricParameter405(vector.Value), sourcehttp.EncodeParameter(vector.Value)
		c := parameterProbe405{Name: vector.Name, LiteralHex: vector.UTF8Hex, OldWireHex: strings.ToUpper(hex.EncodeToString([]byte(old))), SharedWireHex: strings.ToUpper(hex.EncodeToString([]byte(shared))), WireChanged: old != shared}
		if c.WireChanged {
			changed++
		}
		c.OldStatus, c.OldDecodedHex, err = probeParameter405(t.Context(), client, endpoint, user, password, old)
		if err != nil {
			c.OldError = err.Error()
		}
		c.OldMatches = err == nil && c.OldDecodedHex == vector.UTF8Hex
		if !c.OldMatches {
			oldFailures++
		}
		c.SharedStatus, c.SharedDecodedHex, err = probeParameter405(t.Context(), client, endpoint, user, password, shared)
		if err != nil {
			c.SharedError = err.Error()
		}
		c.SharedMatches = err == nil && c.SharedDecodedHex == vector.UTF8Hex
		cases = append(cases, c)
		if !c.SharedMatches {
			t.Errorf("%s shared parameter did not roundtrip independent UTF-8 hex oracle (HTTP %d)", vector.Name, c.SharedStatus)
		}
	}
	if changed == 0 {
		t.Fatal("differential fixture did not distinguish the previous and shared writers")
	}
	t.Logf("%d old/shared String-decoder pairs / %d wire differences / %d old semantic mismatches; shared values must all match independent hex", len(cases), changed, oldFailures)
}

// Keep malformed public-literal refusals with their owning typed compilers;
// internal byte encoding itself intentionally accepts arbitrary bytes.
func TestParameterCanonicalDecoderControls405(t *testing.T) {
	for _, wire := range []string{"\\", "\\x", "\\x0", "\\xzz", "\\x0A", "\\x41", "\\n", "raw\tcontrol"} {
		if _, err := decodeCanonicalParameter405(wire); err == nil {
			t.Fatalf("bad canonical wire admitted: %s", strconv.Quote(wire))
		}
	}
}
