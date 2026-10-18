// Package validation provides validation for data inside widgets.
package validation

import (
	"errors"
	"regexp"

	"github.com/Uzer007-admin/granio/v2"
)

// NewRegexp creates a new validator that uses regular expression parsing.
// The validator will return nil if valid, otherwise returns an error with a reason text.
//
// Since: 1.4
func NewRegexp(regexpstr, reason string) granio.StringValidator {
	expression, err := regexp.Compile(regexpstr)
	if err != nil {
		granio.LogError("Regexp did not compile", err)
		return nil
	}

	err = errors.New(reason)
	return func(text string) error {
		if expression != nil && !expression.MatchString(text) {
			return err
		}

		return nil // Nothing to validate with, same as having no validator.
	}
}
