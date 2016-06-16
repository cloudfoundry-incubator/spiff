package dynaml

type NilExpr struct{}

func (e NilExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	return nil, DefaultInfo(), true
}

func (e NilExpr) String() string {
	return "nil"
}
