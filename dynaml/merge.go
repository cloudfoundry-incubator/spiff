package dynaml

type MergeExpr struct {
	Path []string
}

func (e MergeExpr) Evaluate(context Context) Node {
	return context.FindInStubs(e.Path)
}
