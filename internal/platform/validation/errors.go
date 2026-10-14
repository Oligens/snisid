package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/snisid/platform/internal/platform/errors"
)

// TranslateError converts go-playground ValidationErrors into our structured platform error.
func TranslateError(err error, op string) error {
	if err == nil {
		return nil
	}

	var errMessages []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			msg := fmt.Sprintf("field '%s' failed on the '%s' tag", e.Field(), e.Tag())
			if e.Param() != "" {
				msg = fmt.Sprintf("%s (required %s)", msg, e.Param())
			}
			errMessages = append(errMessages, msg)
		}

		fullMessage := "Validation failed: " + strings.Join(errMessages, "; ")
		return errors.New(errors.InvalidArgument, fullMessage, op, err)
	}

	// Unhandled binding error (e.g. malformed JSON syntax)
	return errors.New(errors.InvalidArgument, "invalid request format", op, err)
}
