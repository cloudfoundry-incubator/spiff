package dynaml

type Builtin struct {
	Function       interface{}
	RequiredPhases StringSet
}

type Builtins map[string]Builtin

func (self Builtins) AddBuiltin(name string, function interface{}, requiredPhases []string) {
	ss := StringSet{}
	ss.UpdateSlice(requiredPhases)
	self[name] = Builtin{
		Function:       function,
		RequiredPhases: ss,
	}
}
