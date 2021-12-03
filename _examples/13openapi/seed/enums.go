package seed

import "github.com/podhmo/apikit/plugins/enum"

var Enums struct {
	SortOrder enum.EnumSet
}

func init() {
	Enums.SortOrder = enum.EnumSet{
		Name: "SortOrder",
		Enums: []enum.Enum{
			{Name: "desc", Value: "desc", Description: "descending order"},
			{Name: "asc", Value: "asc", Description: "ascending order"},
		},
	}
}
