package dynaml

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(context Context) Node {
	a := e.A.Evaluate(context)
	if a != nil {
		return a
	}

	return e.B.Evaluate(context)
}
