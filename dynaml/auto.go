package dynaml

type AutoExpr struct {
	Path []string
}

func (e AutoExpr) Evaluate(Context) Node {
	return nil
}
