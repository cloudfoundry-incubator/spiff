package dynaml

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"

)

type ConcatenationExpr struct {
	A Expression
	B Expression
}

func (e ConcatenationExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true
	
	debug.Debug("CONCAT %+v,%+v\n",e.A,e.B)

	a, infoa, ok := ResolveExpressionOrPushEvaluation(&e.A,&resolved,nil,binding)
	if !ok {
		debug.Debug("  eval a failed\n")
		return nil, infoa, false
	}

	b, info, ok := ResolveExpressionOrPushEvaluation(&e.B,&resolved,&infoa,binding)
	if !ok {
		debug.Debug("  eval b failed\n")
		return nil, info, false
	}
	
	if !resolved {
		debug.Debug("  still unresolved operands\n")
		return node(e), info, true
	}
	
	debug.Debug("CONCAT resolved %+v,%+v\n",a,b)
	
	val, ok := concatenateStringAndInt(a, b)
	if ok {
		debug.Debug("CONCAT --> string %+v\n",val)
		return node(val), info, true
	}

	alist, aok := a.([]yaml.Node)
	if !aok {
		switch a.(type) {
			case map[string]yaml.Node:
				info.Issue="first argument must be list or simple value"
			default:
				info.Issue="simple value can only be concatenated with simple values"
		}
		return nil, info, false
	}
	
	switch b.(type) {
		case []yaml.Node:
			return node(append(alist, b.([]yaml.Node)...)), info, true
		default:
			return node(append(alist, node(b))), info, true
	}
}

func (e ConcatenationExpr) String() string {
	return fmt.Sprintf("%s %s", e.A, e.B)
}

func concatenateStringAndInt(a interface{}, b interface{}) (string, bool) {
	var aString string

	switch v:=a.(type) {
		case string:
			aString = v
		case int64:
			aString = strconv.FormatInt(v, 10)
		case bool:
			aString = strconv.FormatBool(v)
		default:
			return "", false
	}
	
	switch v:=b.(type) {
		case string:
			return  aString + v, true
		case int64:
			return aString + strconv.FormatInt(v, 10), true
		case bool:
			return aString + strconv.FormatBool(v), true
		default:
			return "", false
	}
}
