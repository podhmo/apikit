package namelib

import "strings"

type Normalized string

var coerceMap map[string]Normalized = map[string]Normalized{}

func RegisterNormalized(k string, v string) {
	coerceMap[k] = Normalized(v)
}
func ToNormalized(s string) Normalized {
	v, ok := coerceMap[s]
	if ok {
		return v
	}
	v = toNormalized(s)
	coerceMap[s] = v
	return v
}

func toNormalized(s string) Normalized {
	return Normalized(strings.ToLower(s))
}
