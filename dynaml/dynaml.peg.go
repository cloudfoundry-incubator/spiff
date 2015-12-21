package dynaml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

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
	ruleAuto
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
	"Auto",
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
	Add(rule pegRule, begin, end, next uint32, depth int)
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
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
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
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
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
		token.next = uint32(i)
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
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
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
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
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

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

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
	rules  [26]func() bool
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
	for i, c := range []rune(buffer) {
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

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

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

	add := func(rule pegRule, begin uint32) {
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
		/* 8 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Auto / Reference)> */
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
					if !_rules[ruleAuto]() {
						goto l44
					}
					goto l35
				l44:
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
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if buffer[position] != rune('(') {
					goto l45
				}
				position++
				if !_rules[ruleExpression]() {
					goto l45
				}
				if buffer[position] != rune(')') {
					goto l45
				}
				position++
				depth--
				add(ruleGrouped, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 10 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if !_rules[ruleName]() {
					goto l47
				}
				if buffer[position] != rune('(') {
					goto l47
				}
				position++
				if !_rules[ruleArguments]() {
					goto l47
				}
				if buffer[position] != rune(')') {
					goto l47
				}
				position++
				depth--
				add(ruleCall, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 11 Arguments <- <(Expression (Comma ws Expression)*)> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l49
				}
			l51:
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if !_rules[ruleComma]() {
						goto l52
					}
					if !_rules[rulews]() {
						goto l52
					}
					if !_rules[ruleExpression]() {
						goto l52
					}
					goto l51
				l52:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
				}
				depth--
				add(ruleArguments, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 12 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position53, tokenIndex53, depth53 := position, tokenIndex, depth
			{
				position54 := position
				depth++
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l58
					}
					position++
					goto l57
				l58:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l59
					}
					position++
					goto l57
				l59:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l60
					}
					position++
					goto l57
				l60:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if buffer[position] != rune('_') {
						goto l53
					}
					position++
				}
			l57:
			l55:
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l62
						}
						position++
						goto l61
					l62:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l63
						}
						position++
						goto l61
					l63:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l64
						}
						position++
						goto l61
					l64:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if buffer[position] != rune('_') {
							goto l56
						}
						position++
					}
				l61:
					goto l55
				l56:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
				}
				depth--
				add(ruleName, position54)
			}
			return true
		l53:
			position, tokenIndex, depth = position53, tokenIndex53, depth53
			return false
		},
		/* 13 Comma <- <','> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				if buffer[position] != rune(',') {
					goto l65
				}
				position++
				depth--
				add(ruleComma, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 14 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				{
					position69, tokenIndex69, depth69 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l69
					}
					position++
					goto l70
				l69:
					position, tokenIndex, depth = position69, tokenIndex69, depth69
				}
			l70:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l74
					}
					position++
					goto l73
				l74:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
					if buffer[position] != rune('_') {
						goto l67
					}
					position++
				}
			l73:
			l71:
				{
					position72, tokenIndex72, depth72 := position, tokenIndex, depth
					{
						position75, tokenIndex75, depth75 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l76
						}
						position++
						goto l75
					l76:
						position, tokenIndex, depth = position75, tokenIndex75, depth75
						if buffer[position] != rune('_') {
							goto l72
						}
						position++
					}
				l75:
					goto l71
				l72:
					position, tokenIndex, depth = position72, tokenIndex72, depth72
				}
				depth--
				add(ruleInteger, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 15 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if buffer[position] != rune('"') {
					goto l77
				}
				position++
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l82
						}
						position++
						if buffer[position] != rune('"') {
							goto l82
						}
						position++
						goto l81
					l82:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						{
							position83, tokenIndex83, depth83 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l83
							}
							position++
							goto l80
						l83:
							position, tokenIndex, depth = position83, tokenIndex83, depth83
						}
						if !matchDot() {
							goto l80
						}
					}
				l81:
					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				if buffer[position] != rune('"') {
					goto l77
				}
				position++
				depth--
				add(ruleString, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 16 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l87
					}
					position++
					if buffer[position] != rune('r') {
						goto l87
					}
					position++
					if buffer[position] != rune('u') {
						goto l87
					}
					position++
					if buffer[position] != rune('e') {
						goto l87
					}
					position++
					goto l86
				l87:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('f') {
						goto l84
					}
					position++
					if buffer[position] != rune('a') {
						goto l84
					}
					position++
					if buffer[position] != rune('l') {
						goto l84
					}
					position++
					if buffer[position] != rune('s') {
						goto l84
					}
					position++
					if buffer[position] != rune('e') {
						goto l84
					}
					position++
				}
			l86:
				depth--
				add(ruleBoolean, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 17 Nil <- <('n' 'i' 'l')> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				if buffer[position] != rune('n') {
					goto l88
				}
				position++
				if buffer[position] != rune('i') {
					goto l88
				}
				position++
				if buffer[position] != rune('l') {
					goto l88
				}
				position++
				depth--
				add(ruleNil, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 18 List <- <('[' Contents? ']')> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				if buffer[position] != rune('[') {
					goto l90
				}
				position++
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l92
					}
					goto l93
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
			l93:
				if buffer[position] != rune(']') {
					goto l90
				}
				position++
				depth--
				add(ruleList, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 19 Contents <- <(Expression (Comma ws Expression)*)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l94
				}
			l96:
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if !_rules[ruleComma]() {
						goto l97
					}
					if !_rules[rulews]() {
						goto l97
					}
					if !_rules[ruleExpression]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
				}
				depth--
				add(ruleContents, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 20 Merge <- <('m' 'e' 'r' 'g' 'e')> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if buffer[position] != rune('m') {
					goto l98
				}
				position++
				if buffer[position] != rune('e') {
					goto l98
				}
				position++
				if buffer[position] != rune('r') {
					goto l98
				}
				position++
				if buffer[position] != rune('g') {
					goto l98
				}
				position++
				if buffer[position] != rune('e') {
					goto l98
				}
				position++
				depth--
				add(ruleMerge, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 21 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('a') {
					goto l100
				}
				position++
				if buffer[position] != rune('u') {
					goto l100
				}
				position++
				if buffer[position] != rune('t') {
					goto l100
				}
				position++
				if buffer[position] != rune('o') {
					goto l100
				}
				position++
				depth--
				add(ruleAuto, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 22 Reference <- <('.'? ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (('.' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l104
					}
					position++
					goto l105
				l104:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
				}
			l105:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l107
					}
					position++
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l108
					}
					position++
					goto l106
				l108:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l109
					}
					position++
					goto l106
				l109:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if buffer[position] != rune('_') {
						goto l102
					}
					position++
				}
			l106:
			l110:
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					{
						position112, tokenIndex112, depth112 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l113
						}
						position++
						goto l112
					l113:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l114
						}
						position++
						goto l112
					l114:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l115
						}
						position++
						goto l112
					l115:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
						if buffer[position] != rune('_') {
							goto l116
						}
						position++
						goto l112
					l116:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
						if buffer[position] != rune('-') {
							goto l111
						}
						position++
					}
				l112:
					goto l110
				l111:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
				}
			l117:
				{
					position118, tokenIndex118, depth118 := position, tokenIndex, depth
					{
						position119, tokenIndex119, depth119 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l120
						}
						position++
						{
							position121, tokenIndex121, depth121 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l122
							}
							position++
							goto l121
						l122:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l123
							}
							position++
							goto l121
						l123:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l124
							}
							position++
							goto l121
						l124:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							if buffer[position] != rune('_') {
								goto l120
							}
							position++
						}
					l121:
					l125:
						{
							position126, tokenIndex126, depth126 := position, tokenIndex, depth
							{
								position127, tokenIndex127, depth127 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l128
								}
								position++
								goto l127
							l128:
								position, tokenIndex, depth = position127, tokenIndex127, depth127
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l129
								}
								position++
								goto l127
							l129:
								position, tokenIndex, depth = position127, tokenIndex127, depth127
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l130
								}
								position++
								goto l127
							l130:
								position, tokenIndex, depth = position127, tokenIndex127, depth127
								if buffer[position] != rune('_') {
									goto l131
								}
								position++
								goto l127
							l131:
								position, tokenIndex, depth = position127, tokenIndex127, depth127
								if buffer[position] != rune('-') {
									goto l126
								}
								position++
							}
						l127:
							goto l125
						l126:
							position, tokenIndex, depth = position126, tokenIndex126, depth126
						}
						goto l119
					l120:
						position, tokenIndex, depth = position119, tokenIndex119, depth119
						if buffer[position] != rune('.') {
							goto l118
						}
						position++
						if buffer[position] != rune('[') {
							goto l118
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l118
						}
						position++
					l132:
						{
							position133, tokenIndex133, depth133 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l133
							}
							position++
							goto l132
						l133:
							position, tokenIndex, depth = position133, tokenIndex133, depth133
						}
						if buffer[position] != rune(']') {
							goto l118
						}
						position++
					}
				l119:
					goto l117
				l118:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
				}
				depth--
				add(ruleReference, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 23 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position135 := position
				depth++
			l136:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					{
						position138, tokenIndex138, depth138 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l139
						}
						position++
						goto l138
					l139:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if buffer[position] != rune('\t') {
							goto l140
						}
						position++
						goto l138
					l140:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if buffer[position] != rune('\n') {
							goto l141
						}
						position++
						goto l138
					l141:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if buffer[position] != rune('\r') {
							goto l137
						}
						position++
					}
				l138:
					goto l136
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
				depth--
				add(rulews, position135)
			}
			return true
		},
		/* 24 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l147
					}
					position++
					goto l146
				l147:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if buffer[position] != rune('\t') {
						goto l148
					}
					position++
					goto l146
				l148:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if buffer[position] != rune('\n') {
						goto l149
					}
					position++
					goto l146
				l149:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if buffer[position] != rune('\r') {
						goto l142
					}
					position++
				}
			l146:
			l144:
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					{
						position150, tokenIndex150, depth150 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l151
						}
						position++
						goto l150
					l151:
						position, tokenIndex, depth = position150, tokenIndex150, depth150
						if buffer[position] != rune('\t') {
							goto l152
						}
						position++
						goto l150
					l152:
						position, tokenIndex, depth = position150, tokenIndex150, depth150
						if buffer[position] != rune('\n') {
							goto l153
						}
						position++
						goto l150
					l153:
						position, tokenIndex, depth = position150, tokenIndex150, depth150
						if buffer[position] != rune('\r') {
							goto l145
						}
						position++
					}
				l150:
					goto l144
				l145:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
				}
				depth--
				add(rulereq_ws, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
	}
	p.rules = _rules
}
