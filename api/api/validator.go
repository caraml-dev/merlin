// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"regexp"

	"github.com/go-playground/locales/en_US"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator"
)

var (
	english = en_US.New()
	trans   = ut.New(english, english)
	en, _   = trans.GetTranslator("en")

	subdomainRffc1123 = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]+[a-zA-Z0-9]$`)
)

func NewValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("subdomain_rfc1123", isRFC1123SubDomain)

	validate.RegisterTranslation("required", en, func(ut ut.Translator) error {
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
	validate.RegisterTranslation("min", en, func(ut ut.Translator) error {
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

	validate.RegisterTranslation("max", en, func(ut ut.Translator) error {
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
	validate.RegisterTranslation("subdomain_rfc1123", en, func(ut ut.Translator) error {
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

	return validate
}

func isRFC1123SubDomain(fl validator.FieldLevel) bool {
	return subdomainRffc1123.MatchString(fl.Field().String())
}
