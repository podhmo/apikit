// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runtime

import (
	"m/design"
)

func init() {
	HandleResult = CreateHandleResultFunction(design.HTTPStatusOf)
}
