package web

import "github.com/go-playground/validator/v10"

type Validator struct {
	validator *validator.Validate
}

func (v Validator) New() Validator {
	validate := validator.New()
	v = Validator{
		validator: validate,
	}
	return v
}

func (cv *Validator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
