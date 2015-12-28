package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type MultiplicationExpr struct {
	A Expression
	B Expression
}

func (e MultiplicationExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved:=true
	
	aint, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.A,&resolved,nil,binding)
	if !ok {
		return nil, info, false
	}

	bint, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.B,&resolved,&info,binding)
	if !ok {
		return nil, info, false
	}

    if !resolved {
		return node(e), info, true
	}
	return node(aint * bint), info, true
}

func (e MultiplicationExpr) String() string {
	return fmt.Sprintf("%s * %s", e.A, e.B)
}
