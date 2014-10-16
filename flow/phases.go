package flow

import (
	"reflect"

	"github.com/shutej/spiff/yaml"
)

type Phases struct {
	order []string
	done  map[string]bool
}

func NewPhases(order ...string) *Phases {
	return &Phases{
		order: order,
		done:  map[string]bool{},
	}
}

func (self *Phases) Done() bool {
	return len(self.order) == len(self.done)
}

func (self *Phases) Query(phase string) bool {
	done, ok := self.done[phase]
	return ok && done
}

func (self *Phases) Current() string {
	return self.order[len(self.done)]
}

func (self *Phases) Next() {
	if !self.Done() {
		self.done[self.Current()] = true
	}
}

type Visitor func(yaml.Node) error

func (self *Phases) doPhase(source yaml.Node, stubs []yaml.Node) yaml.Node {
	result := source

	for {
		environment := Environment{
			Phases: self,
			Stubs:  stubs,
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
func (self *Phases) VisitFlow(visitor Visitor, source yaml.Node, stubs ...yaml.Node) error {
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
func (self *Phases) VisitCascade(visitor Visitor, template yaml.Node, templates ...yaml.Node) error {
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
