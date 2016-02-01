package dynaml

import (
	"fmt"
	"net"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type AdditionExpr struct {
	A Expression
	B Expression
}

func (e AdditionExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	resolved := true

	a, info, ok := ResolveExpressionOrPushEvaluation(&e.A, &resolved, nil, binding)
	if !ok {
		return nil, info, false
	}

	bint, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.B, &resolved, &info, binding)
	if !ok {
		return nil, info, false
	}

	if !resolved {
		return e, info, true
	}

	aint, ok := a.(int64)
	if ok {
		return aint + bint, info, true
	}

	str, ok := a.(string)
	if ok {
		ip := net.ParseIP(str)
		if ip != nil {
			return IPAdd(ip, bint).String(), info, true
		}
		info.Issue = yaml.NewIssue("string argument for PLUS must be an IP address")
	} else {
		info.Issue = yaml.NewIssue("first argument of PLUS must be IP address or integer")
	}
	return nil, info, false
}

func (e AdditionExpr) String() string {
	return fmt.Sprintf("%s + %s", e.A, e.B)
}

func IPAdd(ip net.IP, offset int64) net.IP {
	for j := len(ip) - 1; j >= 0; j-- {
		tmp := offset + int64(ip[j])
		ip[j] = byte(tmp)
		if tmp < 0 {
			tmp = tmp - 256
		}
		offset = tmp / 256
		if offset == 0 {
			break
		}
	}
	return ip
}
