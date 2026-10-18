package validation

import (
	"github.com/Uzer007-admin/granio/v2"
)

// NewAllStrings creates a validator that requires all of the passed string validators to pass.
// In short, it combines multiple string validators into one.
//
// Since: 2.2
func NewAllStrings(validators ...granio.StringValidator) granio.StringValidator {
	return func(text string) error {
		for _, validator := range validators {
			if err := validator(text); err != nil {
				return err
			}
		}
		return nil
	}
}
