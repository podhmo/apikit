// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.


package component

import (
	"m/01separated-package/design"
)

type Component interface {
	DB() *design.DB
	Messenger() (*design.Messenger, error)
}