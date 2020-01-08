package errors

import (
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

type kubernetesAPIError struct {
	error
}

func (err *kubernetesAPIError) IsTemporary() bool {
	return true
}

func (err *kubernetesAPIError) IsNotFound() bool {
	return kerrors.IsNotFound(err)
}

// Wrap converts the error to an error which implements the following interfaces:
//
//	- domain/errors.NotFound
//	- domain/errors.Temporary
//
func Wrap(err error) error {
	return &kubernetesAPIError{error: err}
}
