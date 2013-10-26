package yaml

import (
	"errors"
	"fmt"

	"launchpad.net/goyaml"
)

type Node interface{}

type NonStringKeyError struct {
	Key interface{}
}

func (e NonStringKeyError) Error() string {
	return fmt.Sprintf("map key must be a string: %#v", e.Key)
}

func Parse(source []byte) (Node, error) {
	var parsed interface{}

	err := goyaml.Unmarshal(source, &parsed)
	if err != nil {
		return nil, err
	}

	return sanitize(parsed)
}

func sanitize(root interface{}) (Node, error) {
	switch root.(type) {
	case map[interface{}]interface{}:
		sanitized := map[string]Node{}

		for key, val := range root.(map[interface{}]interface{}) {
			str, ok := key.(string)
			if !ok {
				return nil, NonStringKeyError{key}
			}

			sub, err := sanitize(val)
			if err != nil {
				return nil, err
			}

			sanitized[str] = sub
		}

		return Node(sanitized), nil

	case []interface{}:
		sanitized := []Node{}

		for _, val := range root.([]interface{}) {
			sub, err := sanitize(val)
			if err != nil {
				return nil, err
			}

			sanitized = append(sanitized, sub)
		}

		return Node(sanitized), nil

	case string:
		return Node(root.(string)), nil

	case []byte:
		return Node(string(root.([]byte))), nil

	case int:
		return Node(root.(int)), nil

	case float64:
		return Node(root.(float64)), nil

	case bool:
		return Node(root.(bool)), nil

	case nil:
		return Node(nil), nil
	}

	return nil, errors.New(fmt.Sprintf("unknown type during sanitization: %#v\n", root))
}
