package sourcehttp

import "strings"

// EncodeParameter encodes a literal value for ClickHouse's Escaped parameter
// format. Callers subsequently URL-encode it as param_<name>; URL encoding alone
// does not protect backslashes or controls from ClickHouse's parameter parser.
// This byte transformation does not validate questions, identifiers or Unicode.
func EncodeParameter(value string) string {
	const hex = "0123456789abcdef"
	var out strings.Builder
	out.Grow(len(value))
	for i := range len(value) {
		c := value[i]
		switch {
		case c == '\\':
			out.WriteString(`\\`)
		case c < 32 || c == 127:
			out.WriteString(`\x`)
			out.WriteByte(hex[c>>4])
			out.WriteByte(hex[c&15])
		default:
			out.WriteByte(c)
		}
	}
	return out.String()
}
