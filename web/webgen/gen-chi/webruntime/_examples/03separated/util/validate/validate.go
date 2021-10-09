package validate

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func Validate(ob interface{}) error {
	// TODO: merge error
	if err := validate.Struct(ob); err != nil {
		return err
	}
	if v, ok := ob.(interface{ Validate() error }); ok {
		return v.Validate() // TODO: 422
	}
	return nil
}
