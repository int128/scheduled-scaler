package errors

import (
	"golang.org/x/xerrors"
)

// Temporary represents an error that will be fixed by retrying.
type Temporary interface {
	error
	IsTemporary() bool
}

func IsTemporary(err error) bool {
	var e Temporary
	if xerrors.As(err, &e) {
		return e.IsTemporary()
	}
	return false
}

// NotFound represents a resource does not exist.
type NotFound interface {
	error
	IsNotFound() bool
}

func IsNotFound(err error) bool {
	var e NotFound
	if xerrors.As(err, &e) {
		return e.IsNotFound()
	}
	return false
}
