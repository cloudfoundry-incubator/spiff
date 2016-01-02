package dynaml

import (
	"fmt"
	"net"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type SubtractionExpr struct {
	A Expression
	B Expression
}

func (e SubtractionExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved:=true
	
	a, info, ok := ResolveExpressionOrPushEvaluation(&e.A,&resolved,nil,binding)
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
	
	aint, ok:= a.(int64)
	if ok {
	  return node(aint - bint), info, true
	}
	
	str, ok:= a.(string)
	if ok {
		ip:=net.ParseIP(str)
		if ip != nil {
			return node(IPAdd(ip,-bint).String()), info, true
		}
		info.Issue="string argument for MINUS must be an IP address"
	} else {
		info.Issue="first argument of MINUS must be IP address or integer"
	}
	return nil, info, false
}

func (e SubtractionExpr) String() string {
	return fmt.Sprintf("%s - %s", e.A, e.B)
}
