package dynaml

type ReferenceExpr struct {
	Path []string
}

func (e ReferenceExpr) Evaluate(context Context) Node {
	return context.FindReference(e.Path)
}
