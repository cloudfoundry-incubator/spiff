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
	rulePrefer
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
	ruleModulo
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
	ruleRequired
	ruleAuto
	ruleMap
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
	"Prefer",
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
	"Modulo",
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
	"Required",
	"Auto",
	"Map",
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
	rules  [39]func() bool
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
		/* 0 Dynaml <- <((Prefer / Expression) !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[rulePrefer]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleExpression]() {
						goto l0
					}
				}
			l2:
				{
					position4, tokenIndex4, depth4 := position, tokenIndex, depth
					if !matchDot() {
						goto l4
					}
					goto l0
				l4:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
				}
				depth--
				add(ruleDynaml, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Prefer <- <(ws ('p' 'r' 'e' 'f' 'e' 'r') req_ws Expression)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				if !_rules[rulews]() {
					goto l5
				}
				if buffer[position] != rune('p') {
					goto l5
				}
				position++
				if buffer[position] != rune('r') {
					goto l5
				}
				position++
				if buffer[position] != rune('e') {
					goto l5
				}
				position++
				if buffer[position] != rune('f') {
					goto l5
				}
				position++
				if buffer[position] != rune('e') {
					goto l5
				}
				position++
				if buffer[position] != rune('r') {
					goto l5
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l5
				}
				if !_rules[ruleExpression]() {
					goto l5
				}
				depth--
				add(rulePrefer, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 2 Expression <- <(ws Level4 ws)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				if !_rules[rulews]() {
					goto l7
				}
				if !_rules[ruleLevel4]() {
					goto l7
				}
				if !_rules[rulews]() {
					goto l7
				}
				depth--
				add(ruleExpression, position8)
			}
			return true
		l7:
			position, tokenIndex, depth = position7, tokenIndex7, depth7
			return false
		},
		/* 3 Level4 <- <(Level3 (req_ws Or)*)> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l9
				}
			l11:
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l12
					}
					if !_rules[ruleOr]() {
						goto l12
					}
					goto l11
				l12:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
				}
				depth--
				add(ruleLevel4, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 4 Or <- <('|' '|' req_ws Level3)> */
		func() bool {
			position13, tokenIndex13, depth13 := position, tokenIndex, depth
			{
				position14 := position
				depth++
				if buffer[position] != rune('|') {
					goto l13
				}
				position++
				if buffer[position] != rune('|') {
					goto l13
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l13
				}
				if !_rules[ruleLevel3]() {
					goto l13
				}
				depth--
				add(ruleOr, position14)
			}
			return true
		l13:
			position, tokenIndex, depth = position13, tokenIndex13, depth13
			return false
		},
		/* 5 Level3 <- <(Level2 Concatenation*)> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l15
				}
			l17:
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l18
					}
					goto l17
				l18:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
				}
				depth--
				add(ruleLevel3, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 6 Concatenation <- <(req_ws Level2)> */
		func() bool {
			position19, tokenIndex19, depth19 := position, tokenIndex, depth
			{
				position20 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l19
				}
				if !_rules[ruleLevel2]() {
					goto l19
				}
				depth--
				add(ruleConcatenation, position20)
			}
			return true
		l19:
			position, tokenIndex, depth = position19, tokenIndex19, depth19
			return false
		},
		/* 7 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l21
				}
			l23:
				{
					position24, tokenIndex24, depth24 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l24
					}
					{
						position25, tokenIndex25, depth25 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l26
						}
						goto l25
					l26:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if !_rules[ruleSubtraction]() {
							goto l24
						}
					}
				l25:
					goto l23
				l24:
					position, tokenIndex, depth = position24, tokenIndex24, depth24
				}
				depth--
				add(ruleLevel2, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 8 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if buffer[position] != rune('+') {
					goto l27
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l27
				}
				if !_rules[ruleLevel1]() {
					goto l27
				}
				depth--
				add(ruleAddition, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 9 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if buffer[position] != rune('-') {
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
				add(ruleSubtraction, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 10 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l31
				}
			l33:
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l34
					}
					{
						position35, tokenIndex35, depth35 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l36
						}
						goto l35
					l36:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						if !_rules[ruleDivision]() {
							goto l37
						}
						goto l35
					l37:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
						if !_rules[ruleModulo]() {
							goto l34
						}
					}
				l35:
					goto l33
				l34:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
				}
				depth--
				add(ruleLevel1, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 11 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if buffer[position] != rune('*') {
					goto l38
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l38
				}
				if !_rules[ruleLevel0]() {
					goto l38
				}
				depth--
				add(ruleMultiplication, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 12 Division <- <('/' req_ws Level0)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if buffer[position] != rune('/') {
					goto l40
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l40
				}
				if !_rules[ruleLevel0]() {
					goto l40
				}
				depth--
				add(ruleDivision, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 13 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if buffer[position] != rune('%') {
					goto l42
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l42
				}
				if !_rules[ruleLevel0]() {
					goto l42
				}
				depth--
				add(ruleModulo, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 14 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Auto / Map / Reference)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if !_rules[ruleGrouped]() {
						goto l47
					}
					goto l46
				l47:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleCall]() {
						goto l48
					}
					goto l46
				l48:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleBoolean]() {
						goto l49
					}
					goto l46
				l49:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleNil]() {
						goto l50
					}
					goto l46
				l50:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleString]() {
						goto l51
					}
					goto l46
				l51:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleInteger]() {
						goto l52
					}
					goto l46
				l52:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleList]() {
						goto l53
					}
					goto l46
				l53:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleMerge]() {
						goto l54
					}
					goto l46
				l54:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleAuto]() {
						goto l55
					}
					goto l46
				l55:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleMap]() {
						goto l56
					}
					goto l46
				l56:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleReference]() {
						goto l44
					}
				}
			l46:
				depth--
				add(ruleLevel0, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 15 Grouped <- <('(' Expression ')')> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				if buffer[position] != rune('(') {
					goto l57
				}
				position++
				if !_rules[ruleExpression]() {
					goto l57
				}
				if buffer[position] != rune(')') {
					goto l57
				}
				position++
				depth--
				add(ruleGrouped, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 16 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				if !_rules[ruleName]() {
					goto l59
				}
				if buffer[position] != rune('(') {
					goto l59
				}
				position++
				if !_rules[ruleArguments]() {
					goto l59
				}
				if buffer[position] != rune(')') {
					goto l59
				}
				position++
				depth--
				add(ruleCall, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 17 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
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
						goto l61
					}
					position++
				}
			l65:
			l63:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					{
						position69, tokenIndex69, depth69 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l70
						}
						position++
						goto l69
					l70:
						position, tokenIndex, depth = position69, tokenIndex69, depth69
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l71
						}
						position++
						goto l69
					l71:
						position, tokenIndex, depth = position69, tokenIndex69, depth69
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l72
						}
						position++
						goto l69
					l72:
						position, tokenIndex, depth = position69, tokenIndex69, depth69
						if buffer[position] != rune('_') {
							goto l64
						}
						position++
					}
				l69:
					goto l63
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
				depth--
				add(ruleName, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 18 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l73
				}
			l75:
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l76
					}
					goto l75
				l76:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
				}
				depth--
				add(ruleArguments, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 19 NextExpression <- <(Comma Expression)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if !_rules[ruleComma]() {
					goto l77
				}
				if !_rules[ruleExpression]() {
					goto l77
				}
				depth--
				add(ruleNextExpression, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 20 Comma <- <','> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				if buffer[position] != rune(',') {
					goto l79
				}
				position++
				depth--
				add(ruleComma, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 21 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l83
					}
					position++
					goto l84
				l83:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
				}
			l84:
				{
					position87, tokenIndex87, depth87 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l88
					}
					position++
					goto l87
				l88:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
					if buffer[position] != rune('_') {
						goto l81
					}
					position++
				}
			l87:
			l85:
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					{
						position89, tokenIndex89, depth89 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l90
						}
						position++
						goto l89
					l90:
						position, tokenIndex, depth = position89, tokenIndex89, depth89
						if buffer[position] != rune('_') {
							goto l86
						}
						position++
					}
				l89:
					goto l85
				l86:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
				}
				depth--
				add(ruleInteger, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 22 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if buffer[position] != rune('"') {
					goto l91
				}
				position++
			l93:
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					{
						position95, tokenIndex95, depth95 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l96
						}
						position++
						if buffer[position] != rune('"') {
							goto l96
						}
						position++
						goto l95
					l96:
						position, tokenIndex, depth = position95, tokenIndex95, depth95
						{
							position97, tokenIndex97, depth97 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l97
							}
							position++
							goto l94
						l97:
							position, tokenIndex, depth = position97, tokenIndex97, depth97
						}
						if !matchDot() {
							goto l94
						}
					}
				l95:
					goto l93
				l94:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
				}
				if buffer[position] != rune('"') {
					goto l91
				}
				position++
				depth--
				add(ruleString, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 23 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l101
					}
					position++
					if buffer[position] != rune('r') {
						goto l101
					}
					position++
					if buffer[position] != rune('u') {
						goto l101
					}
					position++
					if buffer[position] != rune('e') {
						goto l101
					}
					position++
					goto l100
				l101:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
					if buffer[position] != rune('f') {
						goto l98
					}
					position++
					if buffer[position] != rune('a') {
						goto l98
					}
					position++
					if buffer[position] != rune('l') {
						goto l98
					}
					position++
					if buffer[position] != rune('s') {
						goto l98
					}
					position++
					if buffer[position] != rune('e') {
						goto l98
					}
					position++
				}
			l100:
				depth--
				add(ruleBoolean, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 24 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l105
					}
					position++
					if buffer[position] != rune('i') {
						goto l105
					}
					position++
					if buffer[position] != rune('l') {
						goto l105
					}
					position++
					goto l104
				l105:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
					if buffer[position] != rune('~') {
						goto l102
					}
					position++
				}
			l104:
				depth--
				add(ruleNil, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 25 List <- <('[' Contents? ']')> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				if buffer[position] != rune('[') {
					goto l106
				}
				position++
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l108
					}
					goto l109
				l108:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
				}
			l109:
				if buffer[position] != rune(']') {
					goto l106
				}
				position++
				depth--
				add(ruleList, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 26 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l110
				}
			l112:
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l113
					}
					goto l112
				l113:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
				}
				depth--
				add(ruleContents, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 27 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleSimpleMerge]() {
						goto l114
					}
				}
			l116:
				depth--
				add(ruleMerge, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 28 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws Replace)? req_ws Reference)> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				if buffer[position] != rune('m') {
					goto l118
				}
				position++
				if buffer[position] != rune('e') {
					goto l118
				}
				position++
				if buffer[position] != rune('r') {
					goto l118
				}
				position++
				if buffer[position] != rune('g') {
					goto l118
				}
				position++
				if buffer[position] != rune('e') {
					goto l118
				}
				position++
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l120
					}
					if !_rules[ruleRequired]() {
						goto l120
					}
					goto l118
				l120:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
				}
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l121
					}
					if !_rules[ruleReplace]() {
						goto l121
					}
					goto l122
				l121:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
				}
			l122:
				if !_rules[rulereq_ws]() {
					goto l118
				}
				if !_rules[ruleReference]() {
					goto l118
				}
				depth--
				add(ruleRefMerge, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 29 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required))?)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				if buffer[position] != rune('m') {
					goto l123
				}
				position++
				if buffer[position] != rune('e') {
					goto l123
				}
				position++
				if buffer[position] != rune('r') {
					goto l123
				}
				position++
				if buffer[position] != rune('g') {
					goto l123
				}
				position++
				if buffer[position] != rune('e') {
					goto l123
				}
				position++
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l125
					}
					{
						position127, tokenIndex127, depth127 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l128
						}
						goto l127
					l128:
						position, tokenIndex, depth = position127, tokenIndex127, depth127
						if !_rules[ruleRequired]() {
							goto l125
						}
					}
				l127:
					goto l126
				l125:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
				}
			l126:
				depth--
				add(ruleSimpleMerge, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 30 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				if buffer[position] != rune('r') {
					goto l129
				}
				position++
				if buffer[position] != rune('e') {
					goto l129
				}
				position++
				if buffer[position] != rune('p') {
					goto l129
				}
				position++
				if buffer[position] != rune('l') {
					goto l129
				}
				position++
				if buffer[position] != rune('a') {
					goto l129
				}
				position++
				if buffer[position] != rune('c') {
					goto l129
				}
				position++
				if buffer[position] != rune('e') {
					goto l129
				}
				position++
				depth--
				add(ruleReplace, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 31 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				if buffer[position] != rune('r') {
					goto l131
				}
				position++
				if buffer[position] != rune('e') {
					goto l131
				}
				position++
				if buffer[position] != rune('q') {
					goto l131
				}
				position++
				if buffer[position] != rune('u') {
					goto l131
				}
				position++
				if buffer[position] != rune('i') {
					goto l131
				}
				position++
				if buffer[position] != rune('r') {
					goto l131
				}
				position++
				if buffer[position] != rune('e') {
					goto l131
				}
				position++
				if buffer[position] != rune('d') {
					goto l131
				}
				position++
				depth--
				add(ruleRequired, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 32 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				if buffer[position] != rune('a') {
					goto l133
				}
				position++
				if buffer[position] != rune('u') {
					goto l133
				}
				position++
				if buffer[position] != rune('t') {
					goto l133
				}
				position++
				if buffer[position] != rune('o') {
					goto l133
				}
				position++
				depth--
				add(ruleAuto, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 33 Map <- <('m' 'a' 'p' '[' Expression '|' Name ('|' '-' '>') Expression ']')> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if buffer[position] != rune('m') {
					goto l135
				}
				position++
				if buffer[position] != rune('a') {
					goto l135
				}
				position++
				if buffer[position] != rune('p') {
					goto l135
				}
				position++
				if buffer[position] != rune('[') {
					goto l135
				}
				position++
				if !_rules[ruleExpression]() {
					goto l135
				}
				if buffer[position] != rune('|') {
					goto l135
				}
				position++
				if !_rules[ruleName]() {
					goto l135
				}
				if buffer[position] != rune('|') {
					goto l135
				}
				position++
				if buffer[position] != rune('-') {
					goto l135
				}
				position++
				if buffer[position] != rune('>') {
					goto l135
				}
				position++
				if !_rules[ruleExpression]() {
					goto l135
				}
				if buffer[position] != rune(']') {
					goto l135
				}
				position++
				depth--
				add(ruleMap, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 34 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l139
					}
					position++
					goto l140
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
			l140:
				if !_rules[ruleKey]() {
					goto l137
				}
			l141:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					{
						position143, tokenIndex143, depth143 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l144
						}
						position++
						if !_rules[ruleKey]() {
							goto l144
						}
						goto l143
					l144:
						position, tokenIndex, depth = position143, tokenIndex143, depth143
						if buffer[position] != rune('.') {
							goto l142
						}
						position++
						if buffer[position] != rune('[') {
							goto l142
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l142
						}
						position++
					l145:
						{
							position146, tokenIndex146, depth146 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l146
							}
							position++
							goto l145
						l146:
							position, tokenIndex, depth = position146, tokenIndex146, depth146
						}
						if buffer[position] != rune(']') {
							goto l142
						}
						position++
					}
				l143:
					goto l141
				l142:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
				}
				depth--
				add(ruleReference, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 35 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l151
					}
					position++
					goto l149
				l151:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l152
					}
					position++
					goto l149
				l152:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if buffer[position] != rune('_') {
						goto l147
					}
					position++
				}
			l149:
			l153:
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					{
						position155, tokenIndex155, depth155 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l156
						}
						position++
						goto l155
					l156:
						position, tokenIndex, depth = position155, tokenIndex155, depth155
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l157
						}
						position++
						goto l155
					l157:
						position, tokenIndex, depth = position155, tokenIndex155, depth155
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l158
						}
						position++
						goto l155
					l158:
						position, tokenIndex, depth = position155, tokenIndex155, depth155
						if buffer[position] != rune('_') {
							goto l159
						}
						position++
						goto l155
					l159:
						position, tokenIndex, depth = position155, tokenIndex155, depth155
						if buffer[position] != rune('-') {
							goto l154
						}
						position++
					}
				l155:
					goto l153
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l160
					}
					position++
					{
						position162, tokenIndex162, depth162 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l163
						}
						position++
						goto l162
					l163:
						position, tokenIndex, depth = position162, tokenIndex162, depth162
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l164
						}
						position++
						goto l162
					l164:
						position, tokenIndex, depth = position162, tokenIndex162, depth162
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l165
						}
						position++
						goto l162
					l165:
						position, tokenIndex, depth = position162, tokenIndex162, depth162
						if buffer[position] != rune('_') {
							goto l160
						}
						position++
					}
				l162:
				l166:
					{
						position167, tokenIndex167, depth167 := position, tokenIndex, depth
						{
							position168, tokenIndex168, depth168 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l169
							}
							position++
							goto l168
						l169:
							position, tokenIndex, depth = position168, tokenIndex168, depth168
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l170
							}
							position++
							goto l168
						l170:
							position, tokenIndex, depth = position168, tokenIndex168, depth168
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l171
							}
							position++
							goto l168
						l171:
							position, tokenIndex, depth = position168, tokenIndex168, depth168
							if buffer[position] != rune('_') {
								goto l172
							}
							position++
							goto l168
						l172:
							position, tokenIndex, depth = position168, tokenIndex168, depth168
							if buffer[position] != rune('-') {
								goto l167
							}
							position++
						}
					l168:
						goto l166
					l167:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
					}
					goto l161
				l160:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
				}
			l161:
				depth--
				add(ruleKey, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 36 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position174 := position
				depth++
			l175:
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					{
						position177, tokenIndex177, depth177 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l178
						}
						position++
						goto l177
					l178:
						position, tokenIndex, depth = position177, tokenIndex177, depth177
						if buffer[position] != rune('\t') {
							goto l179
						}
						position++
						goto l177
					l179:
						position, tokenIndex, depth = position177, tokenIndex177, depth177
						if buffer[position] != rune('\n') {
							goto l180
						}
						position++
						goto l177
					l180:
						position, tokenIndex, depth = position177, tokenIndex177, depth177
						if buffer[position] != rune('\r') {
							goto l176
						}
						position++
					}
				l177:
					goto l175
				l176:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
				}
				depth--
				add(rulews, position174)
			}
			return true
		},
		/* 37 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l186
					}
					position++
					goto l185
				l186:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if buffer[position] != rune('\t') {
						goto l187
					}
					position++
					goto l185
				l187:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if buffer[position] != rune('\n') {
						goto l188
					}
					position++
					goto l185
				l188:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if buffer[position] != rune('\r') {
						goto l181
					}
					position++
				}
			l185:
			l183:
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					{
						position189, tokenIndex189, depth189 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l190
						}
						position++
						goto l189
					l190:
						position, tokenIndex, depth = position189, tokenIndex189, depth189
						if buffer[position] != rune('\t') {
							goto l191
						}
						position++
						goto l189
					l191:
						position, tokenIndex, depth = position189, tokenIndex189, depth189
						if buffer[position] != rune('\n') {
							goto l192
						}
						position++
						goto l189
					l192:
						position, tokenIndex, depth = position189, tokenIndex189, depth189
						if buffer[position] != rune('\r') {
							goto l184
						}
						position++
					}
				l189:
					goto l183
				l184:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
				}
				depth--
				add(rulereq_ws, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
	}
	p.rules = _rules
}
