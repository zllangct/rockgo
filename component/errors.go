package Component

import "errors"

var (
	// ErrNullValue is raised when a null value is passed in as a pointer
	ErrNullValue = errors.New("a null value is passed in as a pointer")

	// ErrUnknownComponent is raised when trying to deserialize an unknown component
	ErrUnknownComponent = errors.New("trying to deserialize an unknown component")

	// ErrBadValue is raised when an invalid value is used, eg. for serialization
	ErrBadValue = errors.New("an invalid value is used, eg. for serialization")

	// ErrNoMatch is raised when failed trying to find an object
	ErrNoMatch = errors.New("failed trying to find an object")

	// ErrBadObject is raised when trying to use an object for an invalid purpose (eg. as a parent for itself).
	ErrBadObject = errors.New("trying to use an object for an invalid purpose")

	// ErrNotSupported is raised when trying to perform an invalid operation that is not supported.
	ErrNotSupported = errors.New("rying to perform an invalid operation that is not supported")

	// ErrNotSupported is raised when add a existed component.
	ErrUniqueComponent = errors.New("add a existed component")

	// ErrNotSupported is raised when add a component witch require component is missing.
	ErrMissingComponent = errors.New("add a component witch require component is missing")

	// ErrNotSupported is raised when attach a missing component group.
	ErrMissingGroup = errors.New("attach a missing component group")

	ErrNoThisChild = errors.New("this object has not the child")
)
