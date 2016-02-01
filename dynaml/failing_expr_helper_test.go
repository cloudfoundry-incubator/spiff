package dynaml

type FailingExpr struct{}

func (FailingExpr) Evaluate(Binding) (interface{}, EvaluationInfo, bool) {
	return nil, DefaultInfo(), false
}
