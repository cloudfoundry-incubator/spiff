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
	ruleLevel4
	ruleOr
	ruleLevel3
	ruleConcatenation
	ruleLevel2
	ruleAddition
	ruleSubtraction
	ruleLevel1
	ruleMultiplication
	ruleDivision
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
	ruleRefMerge
	ruleSimpleMerge
	ruleAuto
	ruleReference
	ruleKey
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
	"Level4",
	"Or",
	"Level3",
	"Concatenation",
	"Level2",
	"Addition",
	"Subtraction",
	"Level1",
	"Multiplication",
	"Division",
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
	"RefMerge",
	"SimpleMerge",
	"Auto",
	"Reference",
	"Key",
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
	rules  [33]func() bool
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
		/* 0 Dynaml <- <(Expression !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[ruleExpression]() {
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
		/* 1 Expression <- <(ws Level4 ws)> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !_rules[rulews]() {
					goto l3
				}
				if !_rules[ruleLevel4]() {
					goto l3
				}
				if !_rules[rulews]() {
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
		/* 2 Level4 <- <(Level3 (req_ws Or)*)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l5
				}
			l7:
				{
					position8, tokenIndex8, depth8 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l8
					}
					if !_rules[ruleOr]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex, depth = position8, tokenIndex8, depth8
				}
				depth--
				add(ruleLevel4, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 3 Or <- <('|' '|' req_ws Level3)> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
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
				if !_rules[ruleLevel3]() {
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
		/* 4 Level3 <- <(Level2 Concatenation*)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l11
				}
			l13:
				{
					position14, tokenIndex14, depth14 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position14, tokenIndex14, depth14
				}
				depth--
				add(ruleLevel3, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 5 Concatenation <- <(req_ws Level2)> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l15
				}
				if !_rules[ruleLevel2]() {
					goto l15
				}
				depth--
				add(ruleConcatenation, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 6 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l17
				}
			l19:
				{
					position20, tokenIndex20, depth20 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l20
					}
					{
						position21, tokenIndex21, depth21 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l22
						}
						goto l21
					l22:
						position, tokenIndex, depth = position21, tokenIndex21, depth21
						if !_rules[ruleSubtraction]() {
							goto l20
						}
					}
				l21:
					goto l19
				l20:
					position, tokenIndex, depth = position20, tokenIndex20, depth20
				}
				depth--
				add(ruleLevel2, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 7 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if buffer[position] != rune('+') {
					goto l23
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l23
				}
				if !_rules[ruleLevel1]() {
					goto l23
				}
				depth--
				add(ruleAddition, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 8 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if buffer[position] != rune('-') {
					goto l25
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l25
				}
				if !_rules[ruleLevel1]() {
					goto l25
				}
				depth--
				add(ruleSubtraction, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 9 Level1 <- <(Level0 (req_ws (Multiplication / Division))*)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l27
				}
			l29:
				{
					position30, tokenIndex30, depth30 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l30
					}
					{
						position31, tokenIndex31, depth31 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l32
						}
						goto l31
					l32:
						position, tokenIndex, depth = position31, tokenIndex31, depth31
						if !_rules[ruleDivision]() {
							goto l30
						}
					}
				l31:
					goto l29
				l30:
					position, tokenIndex, depth = position30, tokenIndex30, depth30
				}
				depth--
				add(ruleLevel1, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 10 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				if buffer[position] != rune('*') {
					goto l33
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l33
				}
				if !_rules[ruleLevel0]() {
					goto l33
				}
				depth--
				add(ruleMultiplication, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 11 Division <- <('/' req_ws Level0)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				if buffer[position] != rune('/') {
					goto l35
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l35
				}
				if !_rules[ruleLevel0]() {
					goto l35
				}
				depth--
				add(ruleDivision, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 12 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Auto / Reference)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				{
					position39, tokenIndex39, depth39 := position, tokenIndex, depth
					if !_rules[ruleGrouped]() {
						goto l40
					}
					goto l39
				l40:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleCall]() {
						goto l41
					}
					goto l39
				l41:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleBoolean]() {
						goto l42
					}
					goto l39
				l42:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleNil]() {
						goto l43
					}
					goto l39
				l43:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleString]() {
						goto l44
					}
					goto l39
				l44:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleInteger]() {
						goto l45
					}
					goto l39
				l45:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleList]() {
						goto l46
					}
					goto l39
				l46:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleMerge]() {
						goto l47
					}
					goto l39
				l47:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleAuto]() {
						goto l48
					}
					goto l39
				l48:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
					if !_rules[ruleReference]() {
						goto l37
					}
				}
			l39:
				depth--
				add(ruleLevel0, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 13 Grouped <- <('(' Expression ')')> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				if buffer[position] != rune('(') {
					goto l49
				}
				position++
				if !_rules[ruleExpression]() {
					goto l49
				}
				if buffer[position] != rune(')') {
					goto l49
				}
				position++
				depth--
				add(ruleGrouped, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 14 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				if !_rules[ruleName]() {
					goto l51
				}
				if buffer[position] != rune('(') {
					goto l51
				}
				position++
				if !_rules[ruleArguments]() {
					goto l51
				}
				if buffer[position] != rune(')') {
					goto l51
				}
				position++
				depth--
				add(ruleCall, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 15 Arguments <- <(Expression (Comma Expression)*)> */
		func() bool {
			position53, tokenIndex53, depth53 := position, tokenIndex, depth
			{
				position54 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l53
				}
			l55:
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if !_rules[ruleComma]() {
						goto l56
					}
					if !_rules[ruleExpression]() {
						goto l56
					}
					goto l55
				l56:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
				}
				depth--
				add(ruleArguments, position54)
			}
			return true
		l53:
			position, tokenIndex, depth = position53, tokenIndex53, depth53
			return false
		},
		/* 16 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
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
						goto l57
					}
					position++
				}
			l61:
			l59:
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					{
						position65, tokenIndex65, depth65 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l66
						}
						position++
						goto l65
					l66:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l67
						}
						position++
						goto l65
					l67:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l68
						}
						position++
						goto l65
					l68:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
						if buffer[position] != rune('_') {
							goto l60
						}
						position++
					}
				l65:
					goto l59
				l60:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
				}
				depth--
				add(ruleName, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 17 Comma <- <','> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				if buffer[position] != rune(',') {
					goto l69
				}
				position++
				depth--
				add(ruleComma, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 18 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l73
					}
					position++
					goto l74
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
			l74:
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l78
					}
					position++
					goto l77
				l78:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if buffer[position] != rune('_') {
						goto l71
					}
					position++
				}
			l77:
			l75:
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					{
						position79, tokenIndex79, depth79 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l80
						}
						position++
						goto l79
					l80:
						position, tokenIndex, depth = position79, tokenIndex79, depth79
						if buffer[position] != rune('_') {
							goto l76
						}
						position++
					}
				l79:
					goto l75
				l76:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
				}
				depth--
				add(ruleInteger, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 19 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if buffer[position] != rune('"') {
					goto l81
				}
				position++
			l83:
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					{
						position85, tokenIndex85, depth85 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l86
						}
						position++
						if buffer[position] != rune('"') {
							goto l86
						}
						position++
						goto l85
					l86:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
						{
							position87, tokenIndex87, depth87 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l87
							}
							position++
							goto l84
						l87:
							position, tokenIndex, depth = position87, tokenIndex87, depth87
						}
						if !matchDot() {
							goto l84
						}
					}
				l85:
					goto l83
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
				if buffer[position] != rune('"') {
					goto l81
				}
				position++
				depth--
				add(ruleString, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 20 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l91
					}
					position++
					if buffer[position] != rune('r') {
						goto l91
					}
					position++
					if buffer[position] != rune('u') {
						goto l91
					}
					position++
					if buffer[position] != rune('e') {
						goto l91
					}
					position++
					goto l90
				l91:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if buffer[position] != rune('f') {
						goto l88
					}
					position++
					if buffer[position] != rune('a') {
						goto l88
					}
					position++
					if buffer[position] != rune('l') {
						goto l88
					}
					position++
					if buffer[position] != rune('s') {
						goto l88
					}
					position++
					if buffer[position] != rune('e') {
						goto l88
					}
					position++
				}
			l90:
				depth--
				add(ruleBoolean, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 21 Nil <- <('n' 'i' 'l')> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if buffer[position] != rune('n') {
					goto l92
				}
				position++
				if buffer[position] != rune('i') {
					goto l92
				}
				position++
				if buffer[position] != rune('l') {
					goto l92
				}
				position++
				depth--
				add(ruleNil, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 22 List <- <('[' Contents? ']')> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				if buffer[position] != rune('[') {
					goto l94
				}
				position++
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l96
					}
					goto l97
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
			l97:
				if buffer[position] != rune(']') {
					goto l94
				}
				position++
				depth--
				add(ruleList, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 23 Contents <- <(Expression (Comma Expression)*)> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l98
				}
			l100:
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if !_rules[ruleComma]() {
						goto l101
					}
					if !_rules[ruleExpression]() {
						goto l101
					}
					goto l100
				l101:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
				}
				depth--
				add(ruleContents, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 24 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l105
					}
					goto l104
				l105:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
					if !_rules[ruleSimpleMerge]() {
						goto l102
					}
				}
			l104:
				depth--
				add(ruleMerge, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 25 RefMerge <- <('m' 'e' 'r' 'g' 'e' req_ws Reference)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				if buffer[position] != rune('m') {
					goto l106
				}
				position++
				if buffer[position] != rune('e') {
					goto l106
				}
				position++
				if buffer[position] != rune('r') {
					goto l106
				}
				position++
				if buffer[position] != rune('g') {
					goto l106
				}
				position++
				if buffer[position] != rune('e') {
					goto l106
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l106
				}
				if !_rules[ruleReference]() {
					goto l106
				}
				depth--
				add(ruleRefMerge, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 26 SimpleMerge <- <('m' 'e' 'r' 'g' 'e')> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				if buffer[position] != rune('m') {
					goto l108
				}
				position++
				if buffer[position] != rune('e') {
					goto l108
				}
				position++
				if buffer[position] != rune('r') {
					goto l108
				}
				position++
				if buffer[position] != rune('g') {
					goto l108
				}
				position++
				if buffer[position] != rune('e') {
					goto l108
				}
				position++
				depth--
				add(ruleSimpleMerge, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 27 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if buffer[position] != rune('a') {
					goto l110
				}
				position++
				if buffer[position] != rune('u') {
					goto l110
				}
				position++
				if buffer[position] != rune('t') {
					goto l110
				}
				position++
				if buffer[position] != rune('o') {
					goto l110
				}
				position++
				depth--
				add(ruleAuto, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 28 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l114
					}
					position++
					goto l115
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
			l115:
				if !_rules[ruleKey]() {
					goto l112
				}
			l116:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l119
						}
						position++
						if !_rules[ruleKey]() {
							goto l119
						}
						goto l118
					l119:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
						if buffer[position] != rune('.') {
							goto l117
						}
						position++
						if buffer[position] != rune('[') {
							goto l117
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l117
						}
						position++
					l120:
						{
							position121, tokenIndex121, depth121 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l121
							}
							position++
							goto l120
						l121:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
						}
						if buffer[position] != rune(']') {
							goto l117
						}
						position++
					}
				l118:
					goto l116
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				depth--
				add(ruleReference, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 29 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
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
						goto l122
					}
					position++
				}
			l124:
			l128:
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					{
						position130, tokenIndex130, depth130 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l131
						}
						position++
						goto l130
					l131:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l132
						}
						position++
						goto l130
					l132:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l133
						}
						position++
						goto l130
					l133:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
						if buffer[position] != rune('_') {
							goto l134
						}
						position++
						goto l130
					l134:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
						if buffer[position] != rune('-') {
							goto l129
						}
						position++
					}
				l130:
					goto l128
				l129:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
				}
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l135
					}
					position++
					{
						position137, tokenIndex137, depth137 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l138
						}
						position++
						goto l137
					l138:
						position, tokenIndex, depth = position137, tokenIndex137, depth137
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l139
						}
						position++
						goto l137
					l139:
						position, tokenIndex, depth = position137, tokenIndex137, depth137
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l140
						}
						position++
						goto l137
					l140:
						position, tokenIndex, depth = position137, tokenIndex137, depth137
						if buffer[position] != rune('_') {
							goto l135
						}
						position++
					}
				l137:
				l141:
					{
						position142, tokenIndex142, depth142 := position, tokenIndex, depth
						{
							position143, tokenIndex143, depth143 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l144
							}
							position++
							goto l143
						l144:
							position, tokenIndex, depth = position143, tokenIndex143, depth143
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l145
							}
							position++
							goto l143
						l145:
							position, tokenIndex, depth = position143, tokenIndex143, depth143
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l146
							}
							position++
							goto l143
						l146:
							position, tokenIndex, depth = position143, tokenIndex143, depth143
							if buffer[position] != rune('_') {
								goto l147
							}
							position++
							goto l143
						l147:
							position, tokenIndex, depth = position143, tokenIndex143, depth143
							if buffer[position] != rune('-') {
								goto l142
							}
							position++
						}
					l143:
						goto l141
					l142:
						position, tokenIndex, depth = position142, tokenIndex142, depth142
					}
					goto l136
				l135:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
				}
			l136:
				depth--
				add(ruleKey, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 30 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position149 := position
				depth++
			l150:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					{
						position152, tokenIndex152, depth152 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l153
						}
						position++
						goto l152
					l153:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('\t') {
							goto l154
						}
						position++
						goto l152
					l154:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('\n') {
							goto l155
						}
						position++
						goto l152
					l155:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('\r') {
							goto l151
						}
						position++
					}
				l152:
					goto l150
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
				depth--
				add(rulews, position149)
			}
			return true
		},
		/* 31 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l161
					}
					position++
					goto l160
				l161:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					if buffer[position] != rune('\t') {
						goto l162
					}
					position++
					goto l160
				l162:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					if buffer[position] != rune('\n') {
						goto l163
					}
					position++
					goto l160
				l163:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					if buffer[position] != rune('\r') {
						goto l156
					}
					position++
				}
			l160:
			l158:
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					{
						position164, tokenIndex164, depth164 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l165
						}
						position++
						goto l164
					l165:
						position, tokenIndex, depth = position164, tokenIndex164, depth164
						if buffer[position] != rune('\t') {
							goto l166
						}
						position++
						goto l164
					l166:
						position, tokenIndex, depth = position164, tokenIndex164, depth164
						if buffer[position] != rune('\n') {
							goto l167
						}
						position++
						goto l164
					l167:
						position, tokenIndex, depth = position164, tokenIndex164, depth164
						if buffer[position] != rune('\r') {
							goto l159
						}
						position++
					}
				l164:
					goto l158
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
				depth--
				add(rulereq_ws, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
	}
	p.rules = _rules
}
