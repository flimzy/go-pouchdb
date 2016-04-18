// +build js

package pouchdb

import (
	"github.com/gopherjs/gopherjs/js"
)

// PouchError records an error returned by the PouchDB library
type PouchError struct {
	e       *js.Error
	Status  int    `js:"status"`
	Message string `js:"message"`
	Name    string `js:"name"`
	IsError bool   `js:"isError"`
	Reason  string `js:"reason"`
}

// Error satisfies the error interface for the PouchError type
func (e *PouchError) Error() string {
	return e.Message + ": " + e.Reason
}

// ErrorStatus returns the status of a PouchError, or 0 for other errors
func ErrorStatus(err error) int {
	switch pe := err.(type) {
	case *PouchError:
		return pe.Status
	}
	return 0
}

// ErrorMessage returns the message portion of a PouchError, or "" for other errors
func ErrorMessage(err error) string {
	switch pe := err.(type) {
	case *PouchError:
		return pe.Message
	}
	return ""
}

// ErrorName returns the name portion of a PouchError, or "" for other errors
func ErrorName(err error) string {
	switch pe := err.(type) {
	case *PouchError:
		return pe.Name
	}
	return ""
}

// ErrorName returns the reason portion of a PouchError, or "" for other errors
func ErrorReason(err error) string {
	switch pe := err.(type) {
	case *PouchError:
		return pe.Reason
	}
	return ""
}

// IsNotExist returns true if the passed error represents a PouchError with
// a status of 404 (not found)
func IsNotExist(err error) bool {
	switch pe := err.(type) {
	case *PouchError:
		return pe.Status == 404
	}
	return false
}

// IsConflict returns true if the passed error is a PouchError with a status
// of 409 (conflict)
func IsConflict(err error) bool {
	switch pe := err.(type) {
	case *PouchError:
		return pe.Status == 409
	}
	return false
}

// IsPouchError returns true if the passed error is a PouchError, false
// if it is any other type of error.
func IsPouchError(err error) bool {
	switch err.(type) {
	case *PouchError:
		return true
	}
	return false
}

// NewPouchError creates a new PouchError from a js.Error object returned from the PouchDB library
func NewPouchError(err *js.Error) error {
	if err == nil {
		return nil
	}
	return &PouchError{e: err}
}

// Underlying returns the underlying js.Error object, as returned from the PouchDB library
func (e *PouchError) Underlying() *js.Error {
	return e.e
}
