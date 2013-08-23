package dynaml

type ListExpr struct {
	Contents []Expression
}

func (e ListExpr) Evaluate(context Context) Node {
	nodes := []Node{}

	for _, c := range e.Contents {
		nodes = append(nodes, c.Evaluate(context))
	}

	return nodes
}
