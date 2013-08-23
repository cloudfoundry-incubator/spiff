package dynaml

type StringExpr struct {
	Value string
}

func (e StringExpr) Evaluate(Context) Node {
	return e.Value
}
