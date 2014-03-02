package compare

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Diffing YAML", func() {
	Describe("trivially equal", func() {
		It("reports no differences", func() {
			Expect(
				Compare(parseYAMLFrom("1", "a"), parseYAMLFrom("1", "b")),
			).To(BeEmpty())
		})
	})

	Describe("maps", func() {
		Context("when there is a toplevel difference in value", func() {
			a := parseYAML(`
---
foo: 1
`)

			b := parseYAML(`
---
foo: 2
`)

			It("reports one difference with the key as the path", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					{
						A:    parseYAML("1"),
						B:    parseYAML("2"),
						Path: []string{"foo"},
					},
				}))
			})
		})

		Context("when there is a nested difference in value", func() {
			a := parseYAML(`
---
foo:
  bar: 1
`)

			b := parseYAML(`
---
foo:
  bar: 2
`)

			It("reports one difference for the nested difference, not the wholistic one", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					{
						A:    parseYAML("1"),
						B:    parseYAML("2"),
						Path: []string{"foo", "bar"},
					},
				}))
			})
		})

		Context("when there is a nested difference in type", func() {
			a := parseYAML(`
---
foo:
  bar: 1
`)

			b := parseYAML(`
---
foo: 2
`)

			It("reports one difference with the different nodes", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					{
						A:    parseYAML("bar: 1"),
						B:    parseYAML("2"),
						Path: []string{"foo"},
					},
				}))
			})
		})

		Context("when there is a value missing from B", func() {
			a := parseYAML(`
---
foo: 1
`)

			b := parseYAML(`
--- {}
`)

			It("reports one difference", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					{
						A:    parseYAML("1"),
						B:    nil,
						Path: []string{"foo"},
					},
				}))
			})
		})

		Context("when there is a value missing from A", func() {
			a := parseYAML(`
--- {}
`)

			b := parseYAML(`
---
foo: 1
`)

			It("reports one difference", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					Diff{
						A:    nil,
						B:    parseYAML("1"),
						Path: []string{"foo"},
					},
				}))
			})
		})

		Context("when there is a value missing from both A and B", func() {
			a := parseYAML(`
---
foo: 1
`)

			b := parseYAML(`
---
bar: 2
`)
			It("reports both differences", func() {
				diff := Compare(a, b)

				Expect(diff).To(ContainElement(
					Diff{
						A:    parseYAML("1"),
						B:    nil,
						Path: []string{"foo"},
					},
				))

				Expect(diff).To(ContainElement(
					Diff{
						A:    nil,
						B:    parseYAML("2"),
						Path: []string{"bar"},
					},
				))
			})
		})

		Context("when B is a list of named maps", func() {
			a := parseYAML(`
---
foo:
  fizz: 1
`)

			Context("with no differences", func() {
				b := parseYAML(`
---
- name: foo
  fizz: 1
`)

				It("reports no differences", func() {
					Expect(Compare(a, b)).To(BeEmpty())
				})
			})

			Context("with a different value", func() {
				b := parseYAML(`
---
- name: foo
  fizz: 2
`)

				It("reports no differences", func() {
					Expect(Compare(a, b)).To(Equal([]Diff{
						{
							A:    parseYAML("1"),
							B:    parseYAML("2"),
							Path: []string{"foo", "fizz"},
						},
					}))
				})
			})
		})
	})

	Describe("lists", func() {
		Context("when there is a difference in value", func() {
			a := parseYAML(`
---
- 1
`)

			b := parseYAML(`
---
- 2
`)

			It("reports one difference with the index in the path", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					{
						A:    parseYAML("1"),
						B:    parseYAML("2"),
						Path: []string{"[0]"},
					},
				}))
			})
		})

		Context("when comparing to a non-list", func() {
			a := parseYAML(`
---
- - hello
  - world
`)

			b := parseYAML(`
---
- 42
`)

			It("reports one difference with the index in the path", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					{
						A:    parseYAML("[hello, world]"),
						B:    parseYAML("42"),
						Path: []string{"[0]"},
					},
				}))
			})
		})

		Context("when there are named jobs differing only in order", func() {
			a := parseYAML(`
---
jobs:
- name: a
  value: foo
- name: b
  value: bar
`)

			b := parseYAML(`
---
jobs:
- name: b
  value: bar
- name: a
  value: foo
`)

			It("reports their indices as different", func() {
				diff := Compare(a, b)

				Expect(diff).To(HaveLen(2))

				Expect(diff).To(ContainElement(
					Diff{
						A:    parseYAML("0"),
						B:    parseYAML("1"),
						Path: []string{"jobs", "a", "index"},
					},
				))

				Expect(diff).To(ContainElement(
					Diff{
						A:    parseYAML("1"),
						B:    parseYAML("0"),
						Path: []string{"jobs", "b", "index"},
					},
				))
			})
		})

		Context("when there are named jobs only in B", func() {
			a := parseYAML(`
---
jobs:
- name: a
  value: foo
`)

			b := parseYAML(`
---
jobs:
- name: a
  value: foo
- name: b
  value: bar
`)

			It("reports it as different", func() {
				diff := Compare(a, b)

				Expect(diff).To(Equal([]Diff{
					Diff{
						A:    nil,
						B:    parseYAML("name: b\nvalue: bar\nindex: 1\n"),
						Path: []string{"jobs", "b"},
					},
				}))
			})
		})

		Context("when there are named resource pools differing only in order", func() {
			a := parseYAML(`
---
resource_pools:
- name: a
  value: foo
- name: b
  value: bar
`)

			b := parseYAML(`
---
resource_pools:
- name: b
  value: bar
- name: a
  value: foo
`)

			It("reports no differences", func() {
				Expect(Compare(a, b)).To(BeEmpty())
			})
		})

		Context("when there are values missing from B", func() {
			a := parseYAML(`
---
foo:
- baz:
  - 1
  - 2
  - 3
  - 4
`)

			b := parseYAML(`
---
foo:
- baz:
  - 1
`)

			It("reports each difference", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					Diff{
						A:    parseYAML("2"),
						B:    nil,
						Path: []string{"foo", "[0]", "baz", "[1]"},
					},
					Diff{
						A:    parseYAML("3"),
						B:    nil,
						Path: []string{"foo", "[0]", "baz", "[2]"},
					},
					Diff{
						A:    parseYAML("4"),
						B:    nil,
						Path: []string{"foo", "[0]", "baz", "[3]"},
					},
				}))
			})
		})

		Context("when there is a value missing from A", func() {
			a := parseYAML(`
---
- 1
`)

			b := parseYAML(`
---
- 1
- 2
`)

			It("reports one difference", func() {
				Expect(Compare(a, b)).To(Equal([]Diff{
					Diff{
						A:    nil,
						B:    parseYAML("2"),
						Path: []string{"[1]"},
					},
				}))
			})
		})
	})
})
