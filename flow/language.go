package flow

import (
	"reflect"

	"github.com/shutej/spiff/dynaml"
	"github.com/shutej/spiff/yaml"
)

type Language struct {
	dynaml.Builtins
	order []string
	done  dynaml.StringSet
}

func NewLanguage(order ...string) *Language {
	return &Language{
		order: order,
		done:  dynaml.StringSet{},
	}
}

func (self *Language) Done() bool {
	return len(self.order) == self.done.Len()
}

func (self *Language) HasAll(phases dynaml.StringSet) bool {
	return phases.Difference(self.done).Len() == 0
}

func (self *Language) Current() string {
	return self.order[self.done.Len()]
}

func (self *Language) Next() {
	if !self.Done() {
		self.done.Add(self.Current())
	}
}

type Visitor func(yaml.Node) error

func (self *Language) doPhase(source yaml.Node, stubs []yaml.Node) yaml.Node {
	result := source

	for {
		environment := Environment{
			Language: self,
			Stubs:    stubs,
		}
		next := flow(result, environment, true)

		if reflect.DeepEqual(result, next) {
			break
		}

		result = next
	}

	return result
}

func checkUnresolved(result yaml.Node) error {
	unresolved := findUnresolvedNodes(result)
	if len(unresolved) > 0 {
		return UnresolvedNodes{unresolved}
	}

	return nil
}

// VisitFlow is like flow except that it runs in phases.
func (self *Language) VisitFlow(visitor Visitor, source yaml.Node, stubs ...yaml.Node) error {
	result := source

	for !self.Done() {
		result = self.doPhase(result, stubs)

		if err := visitor(result); err != nil {
			return err
		}

		self.Next()
	}

	// This performs a final check to ensure that all data was resolved.
	return checkUnresolved(result)
}

// VisitCascade is like Cascade except that it runs in phases.
func (self *Language) VisitCascade(visitor Visitor, template yaml.Node, templates ...yaml.Node) error {
	for !self.Done() {
		for i := len(templates) - 1; i >= 0; i-- {
			templates[i] = self.doPhase(templates[i], templates[i+1:])
		}

		template = self.doPhase(template, templates)

		if err := visitor(template); err != nil {
			return err
		}

		self.Next()
	}

	// This performs a final check to ensure that all data was resolved.
	return checkUnresolved(template)
}
