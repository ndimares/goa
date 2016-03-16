// Package goa standardizes on structured error responses: a request that fails because of
// invalid input or unexpected condition produces a response that contains one or more structured
// error(s). Each error object has three keys: a id (number), a title and a message. The title
// for a given id is always the same, the intent is to provide a human friendly categorization.
// The message is specific to the error occurrence and provides additional details that often
// include contextual information (name of parameters etc.).
//
// The basic data structure backing errors is TypedError which simply contains the id and message.
// Multiple errors (not just TypedError instances) can be encapsulated in a MultiError. Both
// TypedError and MultiError implement the error interface, the Error methods return valid JSON
// that can be written directly to a response body.
//
// The code generated by goagen calls the helper functions exposed in this file when it encounters
// invalid data (wrong type, validation errors etc.) such as InvalidParamTypeError,
// InvalidAttributeTypeError etc. These methods take and return an error which is a MultiError that
// gets built over time. The final MultiError object then gets serialized into the response and sent
// back to the client. The response status code is inferred from the type wrapping the error object:
// a BadRequestError produces a 400 status code while any other error produce a 500. This behavior
// can be overridden by setting a custom ErrorHandler in the application.
package goa

import (
	"fmt"
	"strings"
)

var (
	// ErrInvalidParamType is the class of errors produced by the generated code when
	// a request parameter type does not match the design.
	ErrInvalidParamType = NewErrorClass("invalid_parameter_type", 400)

	// ErrMissingParam is the error produced by the generated code when a
	// required request parameter is missing.
	ErrMissingParam = NewErrorClass("missing_parameter", 400)

	// ErrInvalidAttributeType is the error produced by the generated
	// code when a data structure attribute type does not match the design
	// definition.
	ErrInvalidAttributeType = NewErrorClass("invalid_attribute", 400)

	// ErrMissingAttribute is the error produced by the generated
	// code when a data structure attribute required by the design
	// definition is missing.
	ErrMissingAttribute = NewErrorClass("missing_attribute", 400)

	// ErrInvalidEnumValue is the error produced by the generated code when
	// a values does not match one of the values listed in the attribute
	// definition as being valid (i.e. not part of the enum).
	ErrInvalidEnumValue = NewErrorClass("invalid_value", 400)

	// ErrMissingHeader is the error produced by the generated code when a
	// required header is missing.
	ErrMissingHeader = NewErrorClass("missing_header", 400)

	// ErrInvalidFormat is the error produced by the generated code when
	// a value does not match the format specified in the attribute
	// definition.
	ErrInvalidFormat = NewErrorClass("invalid_format", 400)

	// ErrInvalidPattern is the error produced by the generated code when
	// a value does not match the regular expression specified in the
	// attribute definition.
	ErrInvalidPattern = NewErrorClass("invalid_pattern", 400)

	// ErrInvalidRange is the error produced by the generated code when
	// a value is less than the minimum specified in the design definition
	// or more than the maximum.
	ErrInvalidRange = NewErrorClass("invalid_range", 400)

	// ErrInvalidLength is the error produced by the generated code when
	// a value is a slice with less elements than the minimum length
	// specified in the design definition or more elements than the
	// maximum length.
	ErrInvalidLength = NewErrorClass("invalid_length", 400)

	// ErrInvalidEncoding is the error produced when a request body fails
	// to be decoded.
	ErrInvalidEncoding = NewErrorClass("invalid_encoding", 400)

	// ErrInternal is the class of error used for non HTTPError.
	ErrInternal = NewErrorClass("internal", 500)
)

type (
	// HTTPError describes an error that can be returned in a response.
	HTTPError struct {
		// Code identifies the class of errors for client programs.
		Code string `json:"code" xml:"code"`
		// Status is the HTTP status code used by responses that cary the error.
		Status int `json:"status" xml:"status"`
		// Detail describes the specific error occurrence.
		Detail string `json:"detail" xml:"detail"`
		// MetaValues contains additional key/value pairs useful to clients.
		MetaValues map[string]interface{} `json:"meta,omitempty" xml:"meta,omitempty"`
	}

	// ErrorClass is an error generating function.
	// It accepts a format and values and produces errors with the resulting string.
	// If the format is a string or a Stringer then the string value is used.
	// If the format is an error then the string returned by Error() is used.
	// Otherwise the string produced using fmt.Sprintf("%v") is used.
	ErrorClass func(fm interface{}, v ...interface{}) *HTTPError

	// MultiError is an error composed of potentially multiple errors.
	MultiError []error
)

// NewErrorClass creates a new error class.
// It is the responsability of the client to guarantee uniqueness of code.
func NewErrorClass(code string, status int) ErrorClass {
	return func(fm interface{}, v ...interface{}) *HTTPError {
		var f string
		switch actual := fm.(type) {
		case string:
			f = actual
		case error:
			f = actual.Error()
		case fmt.Stringer:
			f = actual.String()
		default:
			f = fmt.Sprintf("%v", actual)
		}
		return &HTTPError{Code: code, Status: status, Detail: fmt.Sprintf(f, v...)}
	}
}

// InvalidParamTypeError creates a HTTPError with class ID ErrInvalidParamType
func InvalidParamTypeError(name string, val interface{}, expected string) error {
	return ErrInvalidParamType("invalid value %#v for parameter %#v, must be a %s", val, name, expected)
}

// MissingParamError creates a HTTPError with class ID ErrMissingParam
func MissingParamError(name string) error {
	return ErrMissingParam("missing required parameter %#v", name)
}

// InvalidAttributeTypeError creates a HTTPError with class ID ErrInvalidAttributeType
func InvalidAttributeTypeError(ctx string, val interface{}, expected string) error {
	return ErrInvalidAttributeType("type of %s must be %s but got value %#v", ctx, expected, val)
}

// MissingAttributeError creates a HTTPError with class ID ErrMissingAttribute
func MissingAttributeError(ctx, name string) error {
	return ErrMissingAttribute("attribute %#v of %s is missing and required", name, ctx)
}

// MissingHeaderError creates a HTTPError with class ID ErrMissingHeader
func MissingHeaderError(name string) error {
	return ErrMissingHeader("missing required HTTP header %#v", name)
}

// InvalidEnumValueError creates a HTTPError with class ID ErrInvalidEnumValue
func InvalidEnumValueError(ctx string, val interface{}, allowed []interface{}) error {
	elems := make([]string, len(allowed))
	for i, a := range allowed {
		elems[i] = fmt.Sprintf("%#v", a)
	}
	return ErrInvalidEnumValue("value of %s must be one of %s but got value %#v", ctx, strings.Join(elems, ", "), val)
}

// InvalidFormatError creates a HTTPError with class ID ErrInvalidFormat
func InvalidFormatError(ctx, target string, format Format, formatError error) error {
	return ErrInvalidFormat("%s must be formatted as a %s but got value %#v, %s", ctx, format, target, formatError.Error())
}

// InvalidPatternError creates a HTTPError with class ID ErrInvalidPattern
func InvalidPatternError(ctx, target string, pattern string) error {
	return ErrInvalidPattern("%s must match the regexp %#v but got value %#v", ctx, pattern, target)
}

// InvalidRangeError creates a HTTPError with class ID ErrInvalidRange
func InvalidRangeError(ctx string, target interface{}, value int, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return ErrInvalidRange("%s must be %s than %d but got value %#v", ctx, comp, value, target)
}

// InvalidLengthError creates a HTTPError with class ID ErrInvalidLength
func InvalidLengthError(ctx string, target interface{}, ln, value int, min bool) error {
	comp := "greater or equal"
	if !min {
		comp = "lesser or equal"
	}
	return ErrInvalidLength("length of %s must be %s than %d but got value %#v (len=%d)", ctx, comp, value, target, ln)
}

// Error returns the error occurrence details.
func (e *HTTPError) Error() string {
	return e.Detail
}

// Meta adds to the error metadata.
func (e *HTTPError) Meta(keyvals ...interface{}) *HTTPError {
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v interface{} = "MISSING"
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		e.MetaValues[fmt.Sprintf("%v", k)] = v
	}
	return e
}

// Error returns the multiple error messages.
func (m MultiError) Error() string {
	errs := make([]string, len(m))
	for i, err := range m {
		errs[i] = err.Error()
	}
	return strings.Join(errs, ", ")
}

// Status computes a status from all the HTTP errors.
// The algorithms returns 500 if any error in the multi error is not a HTTPError or has status 500.
// If all errors are http errors and they all have the same status that status is returned.
// Otherwise Status returns 400.
func (m MultiError) Status() int {
	if len(m) == 0 {
		return 500 // bug
	}
	var status int
	if he, ok := m[0].(*HTTPError); ok {
		status = he.Status
	} else {
		return 500
	}
	if len(m) == 1 {
		return status
	}
	for _, e := range m[1:] {
		if he, ok := e.(*HTTPError); ok {
			if he.Status == 500 {
				return 500
			}
			if he.Status != status {
				status = 400
			}
		} else {
			return 500
		}
	}
	return status
}

// StackErrors coerces the first argument into a MultiError then appends the second argument and
// returns the resulting MultiError.
func StackErrors(err error, err2 error) error {
	if err == nil {
		if err2 == nil {
			return MultiError{}
		}
		if _, ok := err2.(MultiError); ok {
			return err2
		}
		return MultiError{err2}
	}
	merr, ok := err.(MultiError)
	if err2 == nil {
		if ok {
			return merr
		}
		return MultiError{err}
	}
	merr2, ok2 := err2.(MultiError)
	if ok {
		if ok2 {
			return append(merr, merr2...)
		}
		return append(merr, err2)
	}
	merr = MultiError{err}
	if ok2 {
		return append(merr, merr2...)
	}
	return append(merr, err2)
}
