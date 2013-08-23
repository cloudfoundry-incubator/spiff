package dynaml

type FakeContext struct{}

func (FakeContext) FindReference([]string) Node { return nil }
func (FakeContext) FindInStubs([]string) Node   { return nil }
