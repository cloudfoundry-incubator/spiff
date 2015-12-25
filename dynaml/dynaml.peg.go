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
	ruleName
	ruleArguments
	ruleNextExpression
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
	ruleReplace
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
	"Name",
	"Arguments",
	"NextExpression",
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
	"Replace",
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
	rules  [35]func() bool
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
		/* 15 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
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
		/* 16 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l65
				}
			l67:
				{
					position68, tokenIndex68, depth68 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l68
					}
					goto l67
				l68:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
				}
				depth--
				add(ruleArguments, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 17 NextExpression <- <(Comma Expression)> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				if !_rules[ruleComma]() {
					goto l69
				}
				if !_rules[ruleExpression]() {
					goto l69
				}
				depth--
				add(ruleNextExpression, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 18 Comma <- <','> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				if buffer[position] != rune(',') {
					goto l71
				}
				position++
				depth--
				add(ruleComma, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 19 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				{
					position75, tokenIndex75, depth75 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l75
					}
					position++
					goto l76
				l75:
					position, tokenIndex, depth = position75, tokenIndex75, depth75
				}
			l76:
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
						goto l73
					}
					position++
				}
			l79:
			l77:
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l82
						}
						position++
						goto l81
					l82:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if buffer[position] != rune('_') {
							goto l78
						}
						position++
					}
				l81:
					goto l77
				l78:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
				}
				depth--
				add(ruleInteger, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 20 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				if buffer[position] != rune('"') {
					goto l83
				}
				position++
			l85:
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					{
						position87, tokenIndex87, depth87 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l88
						}
						position++
						if buffer[position] != rune('"') {
							goto l88
						}
						position++
						goto l87
					l88:
						position, tokenIndex, depth = position87, tokenIndex87, depth87
						{
							position89, tokenIndex89, depth89 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l89
							}
							position++
							goto l86
						l89:
							position, tokenIndex, depth = position89, tokenIndex89, depth89
						}
						if !matchDot() {
							goto l86
						}
					}
				l87:
					goto l85
				l86:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
				}
				if buffer[position] != rune('"') {
					goto l83
				}
				position++
				depth--
				add(ruleString, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 21 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l93
					}
					position++
					if buffer[position] != rune('r') {
						goto l93
					}
					position++
					if buffer[position] != rune('u') {
						goto l93
					}
					position++
					if buffer[position] != rune('e') {
						goto l93
					}
					position++
					goto l92
				l93:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if buffer[position] != rune('f') {
						goto l90
					}
					position++
					if buffer[position] != rune('a') {
						goto l90
					}
					position++
					if buffer[position] != rune('l') {
						goto l90
					}
					position++
					if buffer[position] != rune('s') {
						goto l90
					}
					position++
					if buffer[position] != rune('e') {
						goto l90
					}
					position++
				}
			l92:
				depth--
				add(ruleBoolean, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 22 Nil <- <('n' 'i' 'l')> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				if buffer[position] != rune('n') {
					goto l94
				}
				position++
				if buffer[position] != rune('i') {
					goto l94
				}
				position++
				if buffer[position] != rune('l') {
					goto l94
				}
				position++
				depth--
				add(ruleNil, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 23 List <- <('[' Contents? ']')> */
		func() bool {
			position96, tokenIndex96, depth96 := position, tokenIndex, depth
			{
				position97 := position
				depth++
				if buffer[position] != rune('[') {
					goto l96
				}
				position++
				{
					position98, tokenIndex98, depth98 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l98
					}
					goto l99
				l98:
					position, tokenIndex, depth = position98, tokenIndex98, depth98
				}
			l99:
				if buffer[position] != rune(']') {
					goto l96
				}
				position++
				depth--
				add(ruleList, position97)
			}
			return true
		l96:
			position, tokenIndex, depth = position96, tokenIndex96, depth96
			return false
		},
		/* 24 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l100
				}
			l102:
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l103
					}
					goto l102
				l103:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
				}
				depth--
				add(ruleContents, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 25 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l107
					}
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleSimpleMerge]() {
						goto l104
					}
				}
			l106:
				depth--
				add(ruleMerge, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 26 RefMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws Replace)? req_ws Reference)> */
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
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l110
					}
					if !_rules[ruleReplace]() {
						goto l110
					}
					goto l111
				l110:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
				}
			l111:
				if !_rules[rulereq_ws]() {
					goto l108
				}
				if !_rules[ruleReference]() {
					goto l108
				}
				depth--
				add(ruleRefMerge, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 27 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws Replace)?)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				if buffer[position] != rune('m') {
					goto l112
				}
				position++
				if buffer[position] != rune('e') {
					goto l112
				}
				position++
				if buffer[position] != rune('r') {
					goto l112
				}
				position++
				if buffer[position] != rune('g') {
					goto l112
				}
				position++
				if buffer[position] != rune('e') {
					goto l112
				}
				position++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l114
					}
					if !_rules[ruleReplace]() {
						goto l114
					}
					goto l115
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
			l115:
				depth--
				add(ruleSimpleMerge, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 28 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
				if buffer[position] != rune('r') {
					goto l116
				}
				position++
				if buffer[position] != rune('e') {
					goto l116
				}
				position++
				if buffer[position] != rune('p') {
					goto l116
				}
				position++
				if buffer[position] != rune('l') {
					goto l116
				}
				position++
				if buffer[position] != rune('a') {
					goto l116
				}
				position++
				if buffer[position] != rune('c') {
					goto l116
				}
				position++
				if buffer[position] != rune('e') {
					goto l116
				}
				position++
				depth--
				add(ruleReplace, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 29 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				if buffer[position] != rune('a') {
					goto l118
				}
				position++
				if buffer[position] != rune('u') {
					goto l118
				}
				position++
				if buffer[position] != rune('t') {
					goto l118
				}
				position++
				if buffer[position] != rune('o') {
					goto l118
				}
				position++
				depth--
				add(ruleAuto, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 30 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l122
					}
					position++
					goto l123
				l122:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
				}
			l123:
				if !_rules[ruleKey]() {
					goto l120
				}
			l124:
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					{
						position126, tokenIndex126, depth126 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l127
						}
						position++
						if !_rules[ruleKey]() {
							goto l127
						}
						goto l126
					l127:
						position, tokenIndex, depth = position126, tokenIndex126, depth126
						if buffer[position] != rune('.') {
							goto l125
						}
						position++
						if buffer[position] != rune('[') {
							goto l125
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l125
						}
						position++
					l128:
						{
							position129, tokenIndex129, depth129 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l129
							}
							position++
							goto l128
						l129:
							position, tokenIndex, depth = position129, tokenIndex129, depth129
						}
						if buffer[position] != rune(']') {
							goto l125
						}
						position++
					}
				l126:
					goto l124
				l125:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
				}
				depth--
				add(ruleReference, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 31 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l133
					}
					position++
					goto l132
				l133:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l134
					}
					position++
					goto l132
				l134:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l135
					}
					position++
					goto l132
				l135:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if buffer[position] != rune('_') {
						goto l130
					}
					position++
				}
			l132:
			l136:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					{
						position138, tokenIndex138, depth138 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l139
						}
						position++
						goto l138
					l139:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l140
						}
						position++
						goto l138
					l140:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l141
						}
						position++
						goto l138
					l141:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if buffer[position] != rune('_') {
							goto l142
						}
						position++
						goto l138
					l142:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if buffer[position] != rune('-') {
							goto l137
						}
						position++
					}
				l138:
					goto l136
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l143
					}
					position++
					{
						position145, tokenIndex145, depth145 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l146
						}
						position++
						goto l145
					l146:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l147
						}
						position++
						goto l145
					l147:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l148
						}
						position++
						goto l145
					l148:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
						if buffer[position] != rune('_') {
							goto l143
						}
						position++
					}
				l145:
				l149:
					{
						position150, tokenIndex150, depth150 := position, tokenIndex, depth
						{
							position151, tokenIndex151, depth151 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l152
							}
							position++
							goto l151
						l152:
							position, tokenIndex, depth = position151, tokenIndex151, depth151
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l153
							}
							position++
							goto l151
						l153:
							position, tokenIndex, depth = position151, tokenIndex151, depth151
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l154
							}
							position++
							goto l151
						l154:
							position, tokenIndex, depth = position151, tokenIndex151, depth151
							if buffer[position] != rune('_') {
								goto l155
							}
							position++
							goto l151
						l155:
							position, tokenIndex, depth = position151, tokenIndex151, depth151
							if buffer[position] != rune('-') {
								goto l150
							}
							position++
						}
					l151:
						goto l149
					l150:
						position, tokenIndex, depth = position150, tokenIndex150, depth150
					}
					goto l144
				l143:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
				}
			l144:
				depth--
				add(ruleKey, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 32 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position157 := position
				depth++
			l158:
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
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
							goto l159
						}
						position++
					}
				l160:
					goto l158
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
				depth--
				add(rulews, position157)
			}
			return true
		},
		/* 33 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l169
					}
					position++
					goto l168
				l169:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
					if buffer[position] != rune('\t') {
						goto l170
					}
					position++
					goto l168
				l170:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
					if buffer[position] != rune('\n') {
						goto l171
					}
					position++
					goto l168
				l171:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
					if buffer[position] != rune('\r') {
						goto l164
					}
					position++
				}
			l168:
			l166:
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					{
						position172, tokenIndex172, depth172 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l173
						}
						position++
						goto l172
					l173:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						if buffer[position] != rune('\t') {
							goto l174
						}
						position++
						goto l172
					l174:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						if buffer[position] != rune('\n') {
							goto l175
						}
						position++
						goto l172
					l175:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						if buffer[position] != rune('\r') {
							goto l167
						}
						position++
					}
				l172:
					goto l166
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
				depth--
				add(rulereq_ws, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
	}
	p.rules = _rules
}
