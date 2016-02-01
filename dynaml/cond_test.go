package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("conditional operator", func() {
	Context("on boolean condition", func() {
		It("returns second if false", func() {
			expr := CondExpr{
				C: BooleanExpr{false},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(3, FakeBinding{}))
		})

		It("returns first if true", func() {
			expr := CondExpr{
				C: BooleanExpr{true},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})
	})

	Context("on integer condition", func() {
		It("returns first if not 0", func() {
			expr := CondExpr{
				C: IntegerExpr{1},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})

		It("returns second if 0", func() {
			expr := CondExpr{
				C: IntegerExpr{0},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(3, FakeBinding{}))
		})
	})

	Context("on list condition", func() {
		It("returns first if length != 0", func() {
			expr := CondExpr{
				C: ListExpr{[]Expression{IntegerExpr{1}}},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})

		It("returns second if length == 0", func() {
			expr := CondExpr{
				C: ListExpr{[]Expression{}},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(3, FakeBinding{}))
		})
	})

	Context("on string condition", func() {
		It("returns first if length != 0", func() {
			expr := CondExpr{
				C: StringExpr{"alice"},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})

		It("returns second if length == 0", func() {
			expr := CondExpr{
				C: StringExpr{""},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(3, FakeBinding{}))
		})
	})

	Context("on map condition", func() {
		It("returns first if length != 0", func() {
			mapNode := node(map[string]yaml.Node{
				"bar": node("alice", nil),
			}, nil)
			expr := CondExpr{
				C: ReferenceExpr{[]string{"foo"}},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{FoundReferences: map[string]yaml.Node{"foo": mapNode}}))
		})

		It("returns second if length == 0", func() {
			mapNode := node(map[string]yaml.Node{}, nil)
			expr := CondExpr{
				C: ReferenceExpr{[]string{"foo"}},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(EvaluateAs(3, FakeBinding{FoundReferences: map[string]yaml.Node{"foo": mapNode}}))
		})
	})

	Context("on failing condition", func() {
		It("fails", func() {
			expr := CondExpr{
				C: FailingExpr{},
				T: IntegerExpr{2},
				F: IntegerExpr{3},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})
