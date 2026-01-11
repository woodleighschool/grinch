package errx

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const jsonTagSplitLimit = 2

var (
	sha256Pattern    = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	cdhashPattern    = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	teamIDPattern    = regexp.MustCompile(`^[A-Z0-9]{10}$`)
	signingIDPattern = regexp.MustCompile(`^(?:[A-Z0-9]{10}|platform):[a-zA-Z0-9.-]+$`)
)

func getValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", jsonTagSplitLimit)[0]
		if name == "-" {
			return fld.Name
		}
		if name != "" {
			return name
		}
		return fld.Name
	})

	_ = v.RegisterValidation("sha256", validateSHA256)
	_ = v.RegisterValidation("cdhash", validateCDHash)
	_ = v.RegisterValidation("teamid", validateTeamID)
	_ = v.RegisterValidation("signingid", validateSigningID)

	return v
}

func validateSHA256(fl validator.FieldLevel) bool {
	return sha256Pattern.MatchString(fl.Field().String())
}

func validateCDHash(fl validator.FieldLevel) bool {
	return cdhashPattern.MatchString(fl.Field().String())
}

func validateTeamID(fl validator.FieldLevel) bool {
	return teamIDPattern.MatchString(fl.Field().String())
}

func validateSigningID(fl validator.FieldLevel) bool {
	return signingIDPattern.MatchString(fl.Field().String())
}

// ValidateStruct validates a struct using go-playground/validator tags.
func ValidateStruct(s any) error {
	v := getValidator()

	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return Internal("validation failed", err)
	}

	fields := make(map[string]string, len(validationErrs))
	for _, fe := range validationErrs {
		field := toSnakeCase(fe.Field())
		if _, exists := fields[field]; !exists {
			fields[field] = validationMessage(fe)
		}
	}

	return &Error{
		Code:    CodeInvalid,
		Message: "Validation failed",
		Fields:  fields,
	}
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too small (min: " + fe.Param() + ")"
	case "max":
		return "Value is too large (max: " + fe.Param() + ")"
	case "gte":
		return "Must be at least " + fe.Param()
	case "lte":
		return "Must be at most " + fe.Param()
	case "gt":
		return "Must be greater than " + fe.Param()
	case "lt":
		return "Must be less than " + fe.Param()
	case "oneof":
		return "Must be one of: " + fe.Param()
	case "uuid":
		return "Must be a valid UUID"
	case "email":
		return "Must be a valid email address"
	case "url":
		return "Must be a valid URL"
	case "sha256":
		return "Must be a SHA256 hash (64 hex chars)"
	case "cdhash":
		return "Must be a CDHash (40 hex chars)"
	case "teamid":
		return "Must be a Team ID (10 alphanumeric chars)"
	case "signingid":
		return "Must be TEAMID:bundle.id or platform:id"
	default:
		return "Invalid value"
	}
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
