package internal

import (
	"errors"

	"github.com/Azure/go-autorest/autorest"
)

func GetAutorestError(err error) (autorest.DetailedError, bool) {
	var restErr autorest.DetailedError
	if errors.As(err, &restErr) {
		return restErr, true
	}

	return autorest.DetailedError{}, false
}
