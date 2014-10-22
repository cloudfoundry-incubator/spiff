package dynaml

import "reflect"

type Builtin struct {
	Function       reflect.Value
	RequiredPhases StringSet
}

type Builtins map[string]Builtin

func (self Builtins) AddBuiltin(name string, function interface{}, requiredPhases []string) {
	ss := StringSet{}
	ss.UpdateSlice(requiredPhases)
	self[name] = Builtin{
		Function:       reflect.ValueOf(function),
		RequiredPhases: ss,
	}
}
