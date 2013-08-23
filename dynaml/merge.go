package dynaml

type MergeExpr struct {
	Path []string
}

func (e MergeExpr) Evaluate(Context) Node {
	return nil
}
