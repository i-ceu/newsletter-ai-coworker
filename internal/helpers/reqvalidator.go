package helpers

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	tr_en "github.com/go-playground/validator/v10/translations/en"
)

func ValidateReq(req interface{}) error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	_ = tr_en.RegisterDefaultTranslations(validate, trans)

	err := validate.Struct(req)
	if err != nil {
		// Translate the error
		translations := err.(validator.ValidationErrors).Translate(trans)
		for _, msg := range translations {
			return errors.New(msg)
		}
	}

	return nil
}
