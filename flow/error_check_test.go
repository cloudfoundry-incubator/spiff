package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reporting issues for unresolved nodes", func() {
	
	It("reports unknown nodes", func() {
		source := parseYAML(`
---
node: (( ref ))
`)
		Expect(source).To(FlowToErr(
			`	(( ref ))	in test	node	()	'ref' not found`,
		))
	})
	
	It("reports addition errors", func() {
		source := parseYAML(`
---
a: true
node: (( a + 1 ))
`)
		Expect(source).To(FlowToErr(
			`	(( a + 1 ))	in test	node	()	first argument of PLUS must be IP address or integer`,
		))
	})
	
	It("reports subtraction errors", func() {
		source := parseYAML(`
---
a: true
node: (( a - 1 ))
`)
		Expect(source).To(FlowToErr(
			`	(( a - 1 ))	in test	node	()	first argument of MINUS must be IP address or integer`,
		))
	})
	
	It("reports division by zero", func() {
		source := parseYAML(`
---
a: 1
node: (( a / 0 ))
`)
		Expect(source).To(FlowToErr(
			`	(( a / 0 ))	in test	node	()	division by zero`,
		))
	})
	
	It("requires integer for second arith operand", func() {
		source := parseYAML(`
---
a: 1
node: (( a / true ))
`)
		Expect(source).To(FlowToErr(
			`	(( a / true ))	in test	node	()	integer operand required`,
		))
	})
	
	It("reports merge failure", func() {
		source := parseYAML(`
---
node: (( merge ))
`)
		Expect(source).To(FlowToErr(
			`	(( merge ))	in test	node	(node)	'node' not found in any stub`,
		))
	})
	
	It("reports merge redirect failure", func() {
		source := parseYAML(`
---
node: (( merge other.node))
`)
		Expect(source).To(FlowToErr(
			`	(( merge other.node ))	in test	node	(other.node)	'other.node' not found in any stub`,
		))
	})
	
	It("reports join failure", func() {
		source := parseYAML(`
---
list:
  - a: true
node: (( join( ",", list.[0] ) ))
`)
		Expect(source).To(FlowToErr(
`	(( join(",", list.[0]) ))	in test	node	()	argument 1 to join must be simple value or list`,
		))
	})
	
	It("reports join failure", func() {
		source := parseYAML(`
---
list:
  - a: true
node: (( join( [], "a" ) ))
`)
		Expect(source).To(FlowToErr(
`	(( join([], "a") ))	in test	node	()	first argument for join must be a string`,
		))
	})
	
	It("reports join failure", func() {
		source := parseYAML(`
---
list:
  - a: true
node: (( join( ",", list ) ))
`)
		Expect(source).To(FlowToErr(
`	(( join(",", list) ))	in test	node	()	elements of list(arg 1) to join must be simple values`,
		))
	})
	
	It("reports ip_min", func() {
		source := parseYAML(`
---
node: (( min_ip( "10" ) ))
`)
		Expect(source).To(FlowToErr(
`	(( min_ip("10") ))	in test	node	()	CIDR argument required`,
		))
	})
	
	It("reports ip_min", func() {
		source := parseYAML(`
---
a:
- a
node: (( "." a ))
`)
		Expect(source).To(FlowToErr(
`	(( "." a ))	in test	node	()	simple value can only be concatenated with simple values`,
		))
	})
	
	It("reports unparseable", func() {
		source := parseYAML(`
---
node: (( a "." ) ))
`)
		Expect(source).To(FlowToErr(
`	(( a "." ) ))	in test	node	()	unparseable expression`,
		))
	})
})
