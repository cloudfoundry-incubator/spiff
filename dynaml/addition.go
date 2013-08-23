package dynaml

type AdditionExpr struct {
	A Expression
	B Expression
}

func (e AdditionExpr) Evaluate(context Context) Node {
	a := e.A.Evaluate(context)
	b := e.B.Evaluate(context)

	aint, ok := a.(int)
	if !ok {
		return nil
	}

	bint, ok := b.(int)
	if !ok {
		return nil
	}

	return aint + bint
}
