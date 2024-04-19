package appvalidate

import (
	"strings"
)

func b64StdEncodingToURLEncoding(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "+", "-"), "/", "_"), "=", "\"")
}
