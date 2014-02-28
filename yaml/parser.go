package yaml

import (
	"errors"
	"fmt"

	"launchpad.net/goyaml"
)

type NonStringKeyError struct {
	Key interface{}
}

func (e NonStringKeyError) Error() string {
	return fmt.Sprintf("map key must be a string: %#v", e.Key)
}

func Parse(sourceName string, source []byte) (Node, error) {
	var parsed interface{}

	err := goyaml.Unmarshal(source, &parsed)
	if err != nil {
		return nil, err
	}

	return sanitize(sourceName, parsed)
}

func sanitize(sourceName string, root interface{}) (Node, error) {
	switch rootVal := root.(type) {
	case map[interface{}]interface{}:
		sanitized := map[string]Node{}

		for key, val := range rootVal {
			str, ok := key.(string)
			if !ok {
				return nil, NonStringKeyError{key}
			}

			sub, err := sanitize(sourceName, val)
			if err != nil {
				return nil, err
			}

			sanitized[str] = sub
		}

		return AnnotatedNode{sanitized, sourceName}, nil

	case []interface{}:
		sanitized := []Node{}

		for _, val := range rootVal {
			sub, err := sanitize(sourceName, val)
			if err != nil {
				return nil, err
			}

			sanitized = append(sanitized, sub)
		}

		return AnnotatedNode{sanitized, sourceName}, nil

	case string, []byte, int, float64, bool, nil:
		return AnnotatedNode{rootVal, sourceName}, nil
	}

	return nil, errors.New(fmt.Sprintf("unknown type during sanitization: %#v\n", root))
}
