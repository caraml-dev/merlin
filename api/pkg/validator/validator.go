package validator

import (
	"regexp"

	"github.com/go-playground/locales/en_US"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator"
)

var (
	english = en_US.New()
	trans   = ut.New(english, english)
	EN, _   = trans.GetTranslator("en")

	subdomainRffc1123 = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]+[a-zA-Z0-9]$`)
)

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()
	err := validate.RegisterValidation("subdomain_rfc1123", isRFC1123SubDomain)
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("required", EN,
		func(ut ut.Translator) error {
			return ut.Add("required", "{0} is required", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			fld := fe.StructField()
			t, err := ut.T(fe.Tag(), fld)
			if err != nil {
				return fe.(error).Error()
			}
			return t
		})
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("min", EN,
		func(ut ut.Translator) error {
			return ut.Add("min", "{0} should be more than {1} characters", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			fld := fe.StructField()
			param := fe.Param()
			t, err := ut.T(fe.Tag(), fld, param)
			if err != nil {
				return fe.(error).Error()
			}
			return t
		})
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("max", EN,
		func(ut ut.Translator) error {
			return ut.Add("max", "{0} should be less than {1} characters", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			fld := fe.StructField()
			param := fe.Param()
			t, err := ut.T(fe.Tag(), fld, param)
			if err != nil {
				return fe.(error).Error()
			}
			return t
		})
	if err != nil {
		return nil, err
	}

	err = validate.RegisterTranslation("subdomain_rfc1123", EN,
		func(ut ut.Translator) error {
			return ut.Add("subdomain_rfc1123", "{0} should be a valid RFC1123 sub-domain", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			fld := fe.StructField()
			t, err := ut.T(fe.Tag(), fld)
			if err != nil {
				return fe.(error).Error()
			}
			return t
		})

	if err != nil {
		return nil, err
	}

	return validate, nil
}

func isRFC1123SubDomain(fl validator.FieldLevel) bool {
	return subdomainRffc1123.MatchString(fl.Field().String())
}
