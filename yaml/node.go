package yaml

import (
	"reflect"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

type Node interface {
	candiedyaml.Marshaler

	Value() interface{}
	SourceName() string
	EquivalentToNode(Node) bool
}

type AnnotatedNode struct {
	value      interface{}
	sourceName string
}

func NewNode(value interface{}, sourcePath string) Node {
	return AnnotatedNode{massageType(value), sourcePath}
}

func massageType(value interface{}) interface{} {
	switch value.(type) {
	case int, int8, int16, int32:
		value = reflect.ValueOf(value).Int()
	}
	return value
}

func (n AnnotatedNode) Value() interface{} {
	return n.value
}

func (n AnnotatedNode) SourceName() string {
	return n.sourceName
}

func (n AnnotatedNode) MarshalYAML() (string, interface{}, error) {
	return "", n.Value(), nil
}

func (n AnnotatedNode) EquivalentToNode(o Node) bool {
	if o == nil {
		return false
	}

	at := reflect.TypeOf(n.Value())
	bt := reflect.TypeOf(o.Value())

	if at != bt {
		return false
	}

	switch nv := n.Value().(type) {
	case map[string]Node:
		ov := o.Value().(map[string]Node)

		if len(nv) != len(ov) {
			return false
		}

		for key, nval := range nv {
			oval, found := ov[key]
			if !found {
				return false
			}

			if !nval.EquivalentToNode(oval) {
				return false
			}
		}

		return true

	case []Node:
		ov := o.Value().([]Node)

		if len(nv) != len(ov) {
			return false
		}

		for i, nval := range nv {
			oval := ov[i]

			if !nval.EquivalentToNode(oval) {
				return false
			}
		}

		return true
	}

	return reflect.DeepEqual(n.Value(), o.Value())
}
