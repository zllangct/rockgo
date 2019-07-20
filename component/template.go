package Component

import (
	"encoding/json"
	"errors"
	"runtime/debug"
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
		return nil, ErrBadValue
	}

	var placeholder interface{}
	err = json.Unmarshal(bytes, &placeholder)
	if err != nil {
		return nil, ErrBadValue
	}

	return placeholder.(map[string]interface{}), nil
}

// AsObject converts a map[string]interface{} as a typed object.
// This is a helper for serializable components.
func DeserializeState(target interface{}, raw interface{}) (err error) {
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str = r.(error).Error()
			case string:
				str = r.(string)
			}
			err = errors.New(str + string(debug.Stack()))

		}
	})()
	if raw == nil {
		return ErrNullValue
	}
	if target == nil {
		return ErrNullValue
	}

	value := raw.(map[string]interface{})

	bytes, err := json.Marshal(value)
	if err != nil {
		return ErrBadValue
	}

	err = json.Unmarshal(bytes, target)
	if err != nil {
		return ErrBadValue
	}

	return nil
}
