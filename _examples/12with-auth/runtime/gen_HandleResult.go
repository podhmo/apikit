// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runtime

import (
	"m/12with-auth/design"
)

func init() {
	HandleResult = CreateHandleResultFunction(design.HTTPStatusOf)
}
