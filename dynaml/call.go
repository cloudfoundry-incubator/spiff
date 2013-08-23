package dynaml

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) Evaluate(Context) Node {
	return nil
}
