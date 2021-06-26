package validate

import (
	"reflect"
	"strings"

	"github.com/google/uuid"

	fallback "github.com/go-playground/locales/en"
	universal "github.com/go-playground/universal-translator"
	v9validator "gopkg.in/go-playground/validator.v9"
	translation "gopkg.in/go-playground/validator.v9/translations/en"
)

type Validator interface {
	ValidateStruct(input interface{}) map[string]string
}

type v struct {
	translator universal.Translator
	validator  *v9validator.Validate
}

func NewValidator() (Validator, error) {
	f := fallback.New()
	u := universal.New(f, f)
	t, _ := u.GetTranslator("en")
	validator, err := validator(t)

	if err != nil {
		return v{}, err
	}

	return v{
		translator: t,
		validator:  validator,
	}, nil
}

func (v v) ValidateStruct(input interface{}) map[string]string {
	errs := v.validator.Struct(input)
	if errs != nil {
		var list = map[string]string{}

		for _, err := range errs.(v9validator.ValidationErrors) {
			list[strings.SplitN(err.Namespace(), ".", 2)[1]] = err.Translate(v.translator)
		}

		return list
	}

	return nil
}

func validator(trans universal.Translator) (*v9validator.Validate, error) {
	val := v9validator.New()
	if err := translation.RegisterDefaultTranslations(val, trans); err != nil {
		return val, err
	}

	val.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return val, nil
}

func IsUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
