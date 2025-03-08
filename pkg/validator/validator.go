package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator предоставляет функции для валидации структур
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// NewValidator создает новый экземпляр Validator
func NewValidator() *Validator {
	validate := validator.New()

	// Регистрируем функцию для получения имен полей из json тегов
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Настраиваем переводчик для ошибок
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)

	return &Validator{
		validate: validate,
		trans:    trans,
	}
}

// Validate выполняет валидацию структуры
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		// Переводим ошибки валидации в понятный формат
		errs := err.(validator.ValidationErrors)
		var errMessages []string
		for _, e := range errs {
			errMessages = append(errMessages, e.Translate(v.trans))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errMessages, "; "))
	}
	return nil
}

// RegisterCustomValidation регистрирует пользовательскую функцию валидации
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}
