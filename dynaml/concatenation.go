package dynaml

import (
	"fmt"
	"strconv"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type ConcatenationExpr struct {
	A Expression
	B Expression
}

func (e ConcatenationExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	
	//fmt.Printf("CONCAT %+v,%+v\n",e.A,e.B)
	a, infoa, ok := e.A.Evaluate(binding)
	if !ok {
		//fmt.Printf("  eval a failed\n")
		return nil, infoa, false
	}

	b, infob, ok := e.B.Evaluate(binding)
	info := infoa.Join(infob)
	if !ok {
		//fmt.Printf("  eval b failed\n")
		return nil, info, false
	}
	
	//fmt.Printf("CONCAT %+v,%+v\n",a,b)
	
	val, ok := concatenateStringAndInt(a, b)
	if ok {
		return node(val), info, true
	}

	alist, aok := a.Value().([]yaml.Node)
	if !aok {
		return nil, info, false
	}
	
	bval := b.Value()
	switch bval.(type) {
		case []yaml.Node:
			return node(append(alist, bval.([]yaml.Node)...)), info, true
		case string:
			return node(append(alist, b)), info, true
		case int64:
			return node(append(alist, b)), info, true
		case map[string]yaml.Node:
			return node(append(alist, b)), info, true
		default:
			return nil, info, false
			
	}
}

func (e ConcatenationExpr) String() string {
	return fmt.Sprintf("%s %s", e.A, e.B)
}

func concatenateStringAndInt(a yaml.Node, b yaml.Node) (string, bool) {
	var aString string

	switch a.Value().(type) {
		case string:
			aString = a.Value().(string)
		case int64:
			aString = strconv.FormatInt(a.Value().(int64), 10)
		default:
			return "", false
	}
	
	switch b.Value().(type) {
		case string:
			return  aString + b.Value().(string), true
		case int64:
			return aString + strconv.FormatInt(b.Value().(int64), 10), true
		default:
			return "", false
	}
}
