package dynaml

type BooleanExpr struct {
	Value bool
}

func (e BooleanExpr) Evaluate(Context) Node {
	return e.Value
}
