// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runner

import (
	"m/01separated-package/component"
	"m/01separated-package/design"
)

func ListUser(component component.Component) ([]*design.User, error) {
	var db *design.DB
	{
		db = component.DB()
	}
	return design.ListUser(db)
}