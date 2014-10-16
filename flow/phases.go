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

// VisitFlow will iterate through phases, waiting until each phase reaches
// stability before calling visitor with the resulting nodes.
func (self *Phases) VisitFlow(visitor Visitor, source yaml.Node, stubs ...yaml.Node) error {
	result := source

	for !self.Done() {
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

		if err := visitor(result); err != nil {
			return err
		}

		self.Next()
	}

	// This performs a final check to ensure that all data was resolved.
	unresolved := findUnresolvedNodes(result)
	if len(unresolved) > 0 {
		return UnresolvedNodes{unresolved}
	}

	return nil
}
