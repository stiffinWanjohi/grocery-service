package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("phone", validatePhone); err != nil {
		panic(
			fmt.Sprintf(
				"failed to register phone validator: %v",
				err,
			),
		)
	}
	if err := validate.RegisterValidation("password", validatePassword); err != nil {
		panic(
			fmt.Sprintf(
				"failed to register password validator: %v",
				err,
			),
		)
	}
}

func Struct(s interface{}) error {
	return validate.Struct(s)
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	regex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	return regex.MatchString(phone)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= 8
}

func FormatError(err error) map[string]string {
	fields := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			fields[field] = formatErrorMsg(e)
		}
	}
	return fields
}

func formatErrorMsg(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "phone":
		return "Invalid phone number format"
	case "password":
		return "Password must be at least 8 characters"
	default:
		return fmt.Sprintf(
			"Invalid value for %s",
			e.Field(),
		)
	}
}
