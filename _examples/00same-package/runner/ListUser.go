// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runner

import (
	"m/00same-package/design"
)

func ListUser(component Component) ([]*design.User, error) {
	var db *design.DB
	{
		db = component.DB()
	}
	return design.ListUser(db)
}
