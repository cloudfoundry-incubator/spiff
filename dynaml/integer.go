package dynaml

type IntegerExpr struct {
	Value int
}

func (e IntegerExpr) Evaluate(Context) Node {
	return e.Value
}
