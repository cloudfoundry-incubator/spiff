package yaml

import (
	"errors"
	"fmt"

	"launchpad.net/goyaml"

	"github.com/vito/spiff/dynaml"
)

type NonStringKeyError struct {
	Key interface{}
}

func (e NonStringKeyError) Error() string {
	return fmt.Sprintf("map key must be a string: %#v", e.Key)
}

func Parse(source []byte) (dynaml.Node, error) {
	var parsed interface{}

	err := goyaml.Unmarshal(source, &parsed)
	if err != nil {
		return nil, err
	}

	return sanitize(parsed)
}

func sanitize(root interface{}) (dynaml.Node, error) {
	switch root.(type) {
	case map[interface{}]interface{}:
		sanitized := map[string]dynaml.Node{}

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

		return dynaml.Node(sanitized), nil

	case []interface{}:
		sanitized := []dynaml.Node{}

		for _, val := range root.([]interface{}) {
			sub, err := sanitize(val)
			if err != nil {
				return nil, err
			}

			sanitized = append(sanitized, sub)
		}

		return dynaml.Node(sanitized), nil

	case string:
		return dynaml.Node(root.(string)), nil

	case []byte:
		return dynaml.Node(string(root.([]byte))), nil

	case int:
		return dynaml.Node(root.(int)), nil

	case bool:
		return dynaml.Node(root.(bool)), nil

	default:
		return nil, errors.New(fmt.Sprintf("unknown type during sanitization: %#v\n", root))
	}
}
