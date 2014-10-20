package dynaml

import (
	"strings"

	"github.com/shutej/spiff/yaml"
)

type FakeBinding struct {
	Builtins        Builtins
	ProvidedPhases  StringSet
	FoundFromRoot   map[string]yaml.Node
	FoundReferences map[string]yaml.Node
	FoundInStubs    map[string]yaml.Node
}

func (c FakeBinding) ProvidesPhases(phases StringSet) bool {
	return phases.Difference(c.ProvidedPhases).Len() == 0
}

func (c FakeBinding) Builtin(name string) (Builtin, bool) {
	intId := func(i int64) int64 {
		return i
	}
	stringId := func(s string) string {
		return s
	}
	if c.Builtins == nil {
		c.Builtins = map[string]Builtin{}
		c.Builtins.AddBuiltin("phase0_int_id", intId, []string{})
		c.Builtins.AddBuiltin("phase0_string_id", stringId, []string{})
		c.Builtins.AddBuiltin("phase1_int_id", intId, []string{"phase1"})
		c.Builtins.AddBuiltin("phase1_string_id", stringId, []string{"phase1"})
		c.Builtins.AddBuiltin("phase2_int_id", intId, []string{"phase2"})
		c.Builtins.AddBuiltin("phase2_string_id", stringId, []string{"phase2"})
	}
	retval, ok := c.Builtins[name]
	return retval, ok
}

func (c FakeBinding) FindFromRoot(path []string) (yaml.Node, bool) {
	val, found := c.FoundFromRoot[strings.Join(path, ".")]
	return val, found
}

func (c FakeBinding) FindReference(path []string) (yaml.Node, bool) {
	val, found := c.FoundReferences[strings.Join(path, ".")]
	return val, found
}

func (c FakeBinding) FindInStubs(path []string) (yaml.Node, bool) {
	val, found := c.FoundInStubs[strings.Join(path, ".")]
	return val, found
}
