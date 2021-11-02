package namelib

import "strings"

func ToExported(name string) string {
	if name == "" {
		return name
	}
	return strings.ToUpper(name[0:1]) + name[1:]
}
