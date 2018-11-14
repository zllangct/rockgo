package Component

// ErrNullValue is raised when a null value is passed in as a pointer
type ErrNullValue struct{}

// ErrUnknownComponent is raised when trying to deserialize an unknown component type
type ErrUnknownComponent struct{}

// ErrBadValue is raised when an invalid value is used, eg. for serialization
type ErrBadValue struct{}

// ErrNoMatch is raised when failed trying to find an object
type ErrNoMatch struct{}

// ErrBadObject is raised when trying to use an object for an invalid purpose (eg. as a parent for itself).
type ErrBadObject struct{}

// ErrNotSupported is raised when trying to perform an invalid operation that is not supported.
type ErrNotSupported struct{}