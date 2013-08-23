package dynaml

type ReferenceExpr struct {
	Path []string
}

func (e ReferenceExpr) Evaluate(Context) Node {
	return nil
}
