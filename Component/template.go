package Component

import (
	"encoding/json"
	"github.com/zllangct/RockGO/3RD/errors"
)

// ObjectTemplate is a simple, flat, serializable object structure that directly converts to and from ObjectsInChildren.
type ObjectTemplate struct {
	Name       string
	Components []ComponentTemplate
	Objects    []ObjectTemplate
}

// ComponentTemplate is a serializable representation of a component
type ComponentTemplate struct {
	Type string
	Data interface{}
}

// FromJson loads an object template from a json block.
func ObjectTemplateFromJson(raw string) (*ObjectTemplate, error) {
	var data ObjectTemplate
	err := json.Unmarshal([]byte(raw), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// AsJson saves an object template as a json block.
func ObjectTemplateAsJson(template *ObjectTemplate) ([]byte, error) {
	raw, err := json.Marshal(template)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// AsObject converts an arbitrary object into a json format map[string]interface{}.
// This is a helper for serializable components.
func SerializeState(state interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(state)
	if err != nil {
		return nil, errors.Fail(ErrBadValue{}, err, "Failed to re-encode data")
	}

	var placeholder interface{}
	err = json.Unmarshal(bytes, &placeholder)
	if err != nil {
		return nil, errors.Fail(ErrBadValue{}, err, "Failed to decode data")
	}

	return placeholder.(map[string]interface{}), nil
}

// AsObject converts a map[string]interface{} as a typed object.
// This is a helper for serializable components.
func DeserializeState(target interface{}, raw interface{}) (err error) {
	defer (func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	})()
	if raw == nil {
		return errors.Fail(ErrNullValue{}, nil, "No data (null)")
	}
	if target == nil {
		return errors.Fail(ErrNullValue{}, nil, "No target (null)")
	}

	value := raw.(map[string]interface{})

	bytes, err := json.Marshal(value)
	if err != nil {
		return errors.Fail(ErrBadValue{}, err, "Failed to re-encode data")
	}

	err = json.Unmarshal(bytes, target)
	if err != nil {
		return errors.Fail(ErrBadValue{}, err, "Failed to decode data")
	}

	return nil
}
