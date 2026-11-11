package sourcehttp

import "testing"

func TestEncodeParameterLiteralBytes405(t *testing.T) {
	for _, c := range []struct{ name, value, want string }{
		{"empty", "", ""},
		{"printable and URI punctuation", "space +%&=#?'\"", "space +%&=#?'\""},
		{"backslash", `a\b\n\x09`, `a\\b\\n\\x09`},
		{"actual controls", "\t\n\r\x00\b\f", `\x09\x0a\x0d\x00\x08\x0c`},
		{"literal versus actual newline", "\\n\n", `\\n\x0a`},
		{"Unicode and combining text", "雪😀e\u0301\u2028\u2029", "雪😀e\u0301\u2028\u2029"},
		{"opaque bytes are not normalized", string([]byte{0xff, 0xfe}), string([]byte{0xff, 0xfe})},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := EncodeParameter(c.value); got != c.want {
				t.Fatalf("encoded bytes %x, want %x", got, c.want)
			}
		})
	}
}

func TestEncodeParameterEveryASCIIControl405(t *testing.T) {
	value := make([]byte, 33)
	for i := range 32 {
		value[i] = byte(i)
	}
	value[32] = 127
	const want = `\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f\x7f`
	if got := EncodeParameter(string(value)); got != want {
		t.Fatalf("%q != %q", got, want)
	}
}
