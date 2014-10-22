package dynaml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDynaml
	ruleExpression
	ruleLevel2
	ruleOr
	ruleLevel1
	ruleConcatenation
	ruleAddition
	ruleSubtraction
	ruleLevel0
	ruleGrouped
	ruleCall
	ruleArguments
	ruleName
	ruleComma
	ruleInteger
	ruleString
	ruleBoolean
	ruleNil
	ruleList
	ruleContents
	ruleMerge
	ruleReference
	rulews
	rulereq_ws

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Dynaml",
	"Expression",
	"Level2",
	"Or",
	"Level1",
	"Concatenation",
	"Addition",
	"Subtraction",
	"Level0",
	"Grouped",
	"Call",
	"Arguments",
	"Name",
	"Comma",
	"Integer",
	"String",
	"Boolean",
	"Nil",
	"List",
	"Contents",
	"Merge",
	"Reference",
	"ws",
	"req_ws",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(buffer[node.begin:node.end]))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegRule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token16{pegRule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token32{pegRule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type DynamlGrammar struct {
	Buffer string
	buffer []rune
	rules  [25]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *DynamlGrammar
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *DynamlGrammar) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *DynamlGrammar) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *DynamlGrammar) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Dynaml <- <(ws Expression ws !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[rulews]() {
					goto l0
				}
				if !_rules[ruleExpression]() {
					goto l0
				}
				if !_rules[rulews]() {
					goto l0
				}
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
				depth--
				add(ruleDynaml, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Expression <- <Level2> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l3
				}
				depth--
				add(ruleExpression, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 Level2 <- <(Or / Level1)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !_rules[ruleOr]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !_rules[ruleLevel1]() {
						goto l5
					}
				}
			l7:
				depth--
				add(ruleLevel2, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 3 Or <- <(Level1 req_ws ('|' '|') req_ws Expression)> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l9
				}
				if !_rules[rulereq_ws]() {
					goto l9
				}
				if buffer[position] != rune('|') {
					goto l9
				}
				position++
				if buffer[position] != rune('|') {
					goto l9
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l9
				}
				if !_rules[ruleExpression]() {
					goto l9
				}
				depth--
				add(ruleOr, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 4 Level1 <- <(Concatenation / Addition / Subtraction / Level0)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !_rules[ruleAddition]() {
						goto l15
					}
					goto l13
				l15:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !_rules[ruleSubtraction]() {
						goto l16
					}
					goto l13
				l16:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !_rules[ruleLevel0]() {
						goto l11
					}
				}
			l13:
				depth--
				add(ruleLevel1, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 5 Concatenation <- <(Level0 (' ' / '\t' / '\n' / '\r')+ Level1)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l17
				}
				{
					position21, tokenIndex21, depth21 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l22
					}
					position++
					goto l21
				l22:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
					if buffer[position] != rune('\t') {
						goto l23
					}
					position++
					goto l21
				l23:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
					if buffer[position] != rune('\n') {
						goto l24
					}
					position++
					goto l21
				l24:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
					if buffer[position] != rune('\r') {
						goto l17
					}
					position++
				}
			l21:
			l19:
				{
					position20, tokenIndex20, depth20 := position, tokenIndex, depth
					{
						position25, tokenIndex25, depth25 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l26
						}
						position++
						goto l25
					l26:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if buffer[position] != rune('\t') {
							goto l27
						}
						position++
						goto l25
					l27:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if buffer[position] != rune('\n') {
							goto l28
						}
						position++
						goto l25
					l28:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if buffer[position] != rune('\r') {
							goto l20
						}
						position++
					}
				l25:
					goto l19
				l20:
					position, tokenIndex, depth = position20, tokenIndex20, depth20
				}
				if !_rules[ruleLevel1]() {
					goto l17
				}
				depth--
				add(ruleConcatenation, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 6 Addition <- <(Level0 req_ws '+' req_ws Level1)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l29
				}
				if !_rules[rulereq_ws]() {
					goto l29
				}
				if buffer[position] != rune('+') {
					goto l29
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l29
				}
				if !_rules[ruleLevel1]() {
					goto l29
				}
				depth--
				add(ruleAddition, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 7 Subtraction <- <(Level0 req_ws '-' req_ws Level1)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l31
				}
				if !_rules[rulereq_ws]() {
					goto l31
				}
				if buffer[position] != rune('-') {
					goto l31
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l31
				}
				if !_rules[ruleLevel1]() {
					goto l31
				}
				depth--
				add(ruleSubtraction, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 8 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Reference)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[ruleGrouped]() {
						goto l36
					}
					goto l35
				l36:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleCall]() {
						goto l37
					}
					goto l35
				l37:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleBoolean]() {
						goto l38
					}
					goto l35
				l38:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleNil]() {
						goto l39
					}
					goto l35
				l39:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleString]() {
						goto l40
					}
					goto l35
				l40:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleInteger]() {
						goto l41
					}
					goto l35
				l41:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleList]() {
						goto l42
					}
					goto l35
				l42:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleMerge]() {
						goto l43
					}
					goto l35
				l43:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleReference]() {
						goto l33
					}
				}
			l35:
				depth--
				add(ruleLevel0, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 9 Grouped <- <('(' Expression ')')> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if buffer[position] != rune('(') {
					goto l44
				}
				position++
				if !_rules[ruleExpression]() {
					goto l44
				}
				if buffer[position] != rune(')') {
					goto l44
				}
				position++
				depth--
				add(ruleGrouped, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 10 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[ruleName]() {
					goto l46
				}
				if buffer[position] != rune('(') {
					goto l46
				}
				position++
				if !_rules[ruleArguments]() {
					goto l46
				}
				if buffer[position] != rune(')') {
					goto l46
				}
				position++
				depth--
				add(ruleCall, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 11 Arguments <- <(Expression (Comma ws Expression)*)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l48
				}
			l50:
				{
					position51, tokenIndex51, depth51 := position, tokenIndex, depth
					if !_rules[ruleComma]() {
						goto l51
					}
					if !_rules[rulews]() {
						goto l51
					}
					if !_rules[ruleExpression]() {
						goto l51
					}
					goto l50
				l51:
					position, tokenIndex, depth = position51, tokenIndex51, depth51
				}
				depth--
				add(ruleArguments, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 12 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l57
					}
					position++
					goto l56
				l57:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l58
					}
					position++
					goto l56
				l58:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l59
					}
					position++
					goto l56
				l59:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if buffer[position] != rune('_') {
						goto l52
					}
					position++
				}
			l56:
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l61
						}
						position++
						goto l60
					l61:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l62
						}
						position++
						goto l60
					l62:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l63
						}
						position++
						goto l60
					l63:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if buffer[position] != rune('_') {
							goto l55
						}
						position++
					}
				l60:
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				depth--
				add(ruleName, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 13 Comma <- <','> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if buffer[position] != rune(',') {
					goto l64
				}
				position++
				depth--
				add(ruleComma, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 14 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				{
					position68, tokenIndex68, depth68 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l68
					}
					position++
					goto l69
				l68:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
				}
			l69:
				{
					position72, tokenIndex72, depth72 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l73
					}
					position++
					goto l72
				l73:
					position, tokenIndex, depth = position72, tokenIndex72, depth72
					if buffer[position] != rune('_') {
						goto l66
					}
					position++
				}
			l72:
			l70:
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					{
						position74, tokenIndex74, depth74 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l75
						}
						position++
						goto l74
					l75:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						if buffer[position] != rune('_') {
							goto l71
						}
						position++
					}
				l74:
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				depth--
				add(ruleInteger, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 15 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != rune('"') {
					goto l76
				}
				position++
			l78:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					{
						position80, tokenIndex80, depth80 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l81
						}
						position++
						if buffer[position] != rune('"') {
							goto l81
						}
						position++
						goto l80
					l81:
						position, tokenIndex, depth = position80, tokenIndex80, depth80
						{
							position82, tokenIndex82, depth82 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l82
							}
							position++
							goto l79
						l82:
							position, tokenIndex, depth = position82, tokenIndex82, depth82
						}
						if !matchDot() {
							goto l79
						}
					}
				l80:
					goto l78
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
				if buffer[position] != rune('"') {
					goto l76
				}
				position++
				depth--
				add(ruleString, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 16 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l86
					}
					position++
					if buffer[position] != rune('r') {
						goto l86
					}
					position++
					if buffer[position] != rune('u') {
						goto l86
					}
					position++
					if buffer[position] != rune('e') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('f') {
						goto l83
					}
					position++
					if buffer[position] != rune('a') {
						goto l83
					}
					position++
					if buffer[position] != rune('l') {
						goto l83
					}
					position++
					if buffer[position] != rune('s') {
						goto l83
					}
					position++
					if buffer[position] != rune('e') {
						goto l83
					}
					position++
				}
			l85:
				depth--
				add(ruleBoolean, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 17 Nil <- <('n' 'i' 'l')> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != rune('n') {
					goto l87
				}
				position++
				if buffer[position] != rune('i') {
					goto l87
				}
				position++
				if buffer[position] != rune('l') {
					goto l87
				}
				position++
				depth--
				add(ruleNil, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 18 List <- <('[' Contents? ']')> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('[') {
					goto l89
				}
				position++
				{
					position91, tokenIndex91, depth91 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l91
					}
					goto l92
				l91:
					position, tokenIndex, depth = position91, tokenIndex91, depth91
				}
			l92:
				if buffer[position] != rune(']') {
					goto l89
				}
				position++
				depth--
				add(ruleList, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 19 Contents <- <(Expression (Comma ws Expression)*)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l93
				}
			l95:
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[ruleComma]() {
						goto l96
					}
					if !_rules[rulews]() {
						goto l96
					}
					if !_rules[ruleExpression]() {
						goto l96
					}
					goto l95
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
				depth--
				add(ruleContents, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 20 Merge <- <('m' 'e' 'r' 'g' 'e')> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if buffer[position] != rune('m') {
					goto l97
				}
				position++
				if buffer[position] != rune('e') {
					goto l97
				}
				position++
				if buffer[position] != rune('r') {
					goto l97
				}
				position++
				if buffer[position] != rune('g') {
					goto l97
				}
				position++
				if buffer[position] != rune('e') {
					goto l97
				}
				position++
				depth--
				add(ruleMerge, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 21 Reference <- <('.'? ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (('.' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l101
					}
					position++
					goto l102
				l101:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
				}
			l102:
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l104
					}
					position++
					goto l103
				l104:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l105
					}
					position++
					goto l103
				l105:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l106
					}
					position++
					goto l103
				l106:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
					if buffer[position] != rune('_') {
						goto l99
					}
					position++
				}
			l103:
			l107:
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					{
						position109, tokenIndex109, depth109 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l110
						}
						position++
						goto l109
					l110:
						position, tokenIndex, depth = position109, tokenIndex109, depth109
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l111
						}
						position++
						goto l109
					l111:
						position, tokenIndex, depth = position109, tokenIndex109, depth109
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l112
						}
						position++
						goto l109
					l112:
						position, tokenIndex, depth = position109, tokenIndex109, depth109
						if buffer[position] != rune('_') {
							goto l113
						}
						position++
						goto l109
					l113:
						position, tokenIndex, depth = position109, tokenIndex109, depth109
						if buffer[position] != rune('-') {
							goto l108
						}
						position++
					}
				l109:
					goto l107
				l108:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
				}
			l114:
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					{
						position116, tokenIndex116, depth116 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l117
						}
						position++
						{
							position118, tokenIndex118, depth118 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l119
							}
							position++
							goto l118
						l119:
							position, tokenIndex, depth = position118, tokenIndex118, depth118
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l120
							}
							position++
							goto l118
						l120:
							position, tokenIndex, depth = position118, tokenIndex118, depth118
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l121
							}
							position++
							goto l118
						l121:
							position, tokenIndex, depth = position118, tokenIndex118, depth118
							if buffer[position] != rune('_') {
								goto l117
							}
							position++
						}
					l118:
					l122:
						{
							position123, tokenIndex123, depth123 := position, tokenIndex, depth
							{
								position124, tokenIndex124, depth124 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l125
								}
								position++
								goto l124
							l125:
								position, tokenIndex, depth = position124, tokenIndex124, depth124
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l126
								}
								position++
								goto l124
							l126:
								position, tokenIndex, depth = position124, tokenIndex124, depth124
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l127
								}
								position++
								goto l124
							l127:
								position, tokenIndex, depth = position124, tokenIndex124, depth124
								if buffer[position] != rune('_') {
									goto l128
								}
								position++
								goto l124
							l128:
								position, tokenIndex, depth = position124, tokenIndex124, depth124
								if buffer[position] != rune('-') {
									goto l123
								}
								position++
							}
						l124:
							goto l122
						l123:
							position, tokenIndex, depth = position123, tokenIndex123, depth123
						}
						goto l116
					l117:
						position, tokenIndex, depth = position116, tokenIndex116, depth116
						if buffer[position] != rune('.') {
							goto l115
						}
						position++
						if buffer[position] != rune('[') {
							goto l115
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l115
						}
						position++
					l129:
						{
							position130, tokenIndex130, depth130 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l130
							}
							position++
							goto l129
						l130:
							position, tokenIndex, depth = position130, tokenIndex130, depth130
						}
						if buffer[position] != rune(']') {
							goto l115
						}
						position++
					}
				l116:
					goto l114
				l115:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
				}
				depth--
				add(ruleReference, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 22 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position132 := position
				depth++
			l133:
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					{
						position135, tokenIndex135, depth135 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l136
						}
						position++
						goto l135
					l136:
						position, tokenIndex, depth = position135, tokenIndex135, depth135
						if buffer[position] != rune('\t') {
							goto l137
						}
						position++
						goto l135
					l137:
						position, tokenIndex, depth = position135, tokenIndex135, depth135
						if buffer[position] != rune('\n') {
							goto l138
						}
						position++
						goto l135
					l138:
						position, tokenIndex, depth = position135, tokenIndex135, depth135
						if buffer[position] != rune('\r') {
							goto l134
						}
						position++
					}
				l135:
					goto l133
				l134:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
				}
				depth--
				add(rulews, position132)
			}
			return true
		},
		/* 23 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l144
					}
					position++
					goto l143
				l144:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if buffer[position] != rune('\t') {
						goto l145
					}
					position++
					goto l143
				l145:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if buffer[position] != rune('\n') {
						goto l146
					}
					position++
					goto l143
				l146:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if buffer[position] != rune('\r') {
						goto l139
					}
					position++
				}
			l143:
			l141:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					{
						position147, tokenIndex147, depth147 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l148
						}
						position++
						goto l147
					l148:
						position, tokenIndex, depth = position147, tokenIndex147, depth147
						if buffer[position] != rune('\t') {
							goto l149
						}
						position++
						goto l147
					l149:
						position, tokenIndex, depth = position147, tokenIndex147, depth147
						if buffer[position] != rune('\n') {
							goto l150
						}
						position++
						goto l147
					l150:
						position, tokenIndex, depth = position147, tokenIndex147, depth147
						if buffer[position] != rune('\r') {
							goto l142
						}
						position++
					}
				l147:
					goto l141
				l142:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
				}
				depth--
				add(rulereq_ws, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
	}
	p.rules = _rules
}
