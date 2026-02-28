package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrPermissionDenied = errors.New("permission denied")
	ErrCancelled        = errors.New("operation cancelled")
	ErrTimeout          = errors.New("operation timeout")
	ErrNotSupported     = errors.New("not supported")
	ErrInternal         = errors.New("internal error")
)

type Error struct {
	Op   string
	Err  error
	Kind ErrorKind
	Key  string
}

type ErrorKind int

const (
	KindUnknown ErrorKind = iota
	KindNotFound
	KindAlreadyExists
	KindInvalidInput
	KindPermission
	KindNetwork
	KindIO
	KindExec
	KindCancelled
	KindTimeout
	KindConflict
)

func (e *Error) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(kind ErrorKind, msg string) error {
	return &Error{
		Err:  errors.New(msg),
		Kind: kind,
	}
}

func Newf(kind ErrorKind, format string, args ...interface{}) error {
	return &Error{
		Err:  fmt.Errorf(format, args...),
		Kind: kind,
	}
}

func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:  op,
		Err: err,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:  fmt.Sprintf(format, args...),
		Err: err,
	}
}

func WrapWithKind(err error, op string, kind ErrorKind) error {
	if err == nil {
		return nil
	}
	return &Error{
		Op:   op,
		Err:  err,
		Kind: kind,
	}
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func GetKind(err error) ErrorKind {
	if err == nil {
		return KindUnknown
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindUnknown
}

func IsKind(err error, kind ErrorKind) bool {
	return GetKind(err) == kind
}
