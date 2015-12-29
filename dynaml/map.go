package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"
)

type MapContext struct {
	Binding
	Name string
	Node yaml.Node
}

func (c MapContext) FindReference(path []string) (yaml.Node, bool) {
	if len(path)==1 && path[0]==c.Name {
		return c.Node, true
	}
	return c.Binding.FindReference(path)
}

type MapExpr struct {
	A Expression
	Name string
	B Expression
}

func (e MapExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved:=true
	
	debug.Debug("evaluate mapping\n")
	value, info, ok := ResolveExpressionOrPushEvaluation(&e.A,&resolved,nil,binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, ok
	}
	
    list, ok := value.([]yaml.Node)
	if !ok {
		return nil, info, false
	}
	
	debug.Debug("map: using expression %+v\n", e.B)
	result:=[]yaml.Node{}
	for i,n := range list {
		debug.Debug("map:  mapping for %d: %+v\n",i,n)
		ctx := MapContext{binding, e.Name, n}
		mapped,info,ok := e.B.Evaluate(ctx)
		if !ok {
			debug.Debug("map:  %d %+v: failed\n",i,n)
			return nil, info, false
		}
		
		_, ok = mapped.Value().(Expression)
		if ok {
			debug.Debug("map:  %d unresolved  -> KEEP\n")
			return node(e), info, true
		}
		debug.Debug("map:  %d --> %+v\n",i,mapped)
		result=append(result,mapped)
	}
	
	debug.Debug("map: --> %+v\n", result)
	return node(result), info, true
}

func (e MapExpr) String() string {
	return fmt.Sprintf("map[%s|%s|->%s]", e.A, e.Name,  e.B)
}