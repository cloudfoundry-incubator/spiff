package dynaml

type Node interface{}

type Context interface {
	FindReference([]string) Node
	FindInStubs([]string) Node
}

type Expression interface {
	Evaluate(Context) Node
}
