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
	rules  [38]func() bool
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
		/* 14 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Auto / Reference)> */
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
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if buffer[position] != rune('(') {
					goto l56
				}
				position++
				if !_rules[ruleExpression]() {
					goto l56
				}
				if buffer[position] != rune(')') {
					goto l56
				}
				position++
				depth--
				add(ruleGrouped, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 16 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if !_rules[ruleName]() {
					goto l58
				}
				if buffer[position] != rune('(') {
					goto l58
				}
				position++
				if !_rules[ruleArguments]() {
					goto l58
				}
				if buffer[position] != rune(')') {
					goto l58
				}
				position++
				depth--
				add(ruleCall, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 17 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l65
					}
					position++
					goto l64
				l65:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l66
					}
					position++
					goto l64
				l66:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l67
					}
					position++
					goto l64
				l67:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('_') {
						goto l60
					}
					position++
				}
			l64:
			l62:
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					{
						position68, tokenIndex68, depth68 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l69
						}
						position++
						goto l68
					l69:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l70
						}
						position++
						goto l68
					l70:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l71
						}
						position++
						goto l68
					l71:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if buffer[position] != rune('_') {
							goto l63
						}
						position++
					}
				l68:
					goto l62
				l63:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
				}
				depth--
				add(ruleName, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 18 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l72
				}
			l74:
				{
					position75, tokenIndex75, depth75 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l75
					}
					goto l74
				l75:
					position, tokenIndex, depth = position75, tokenIndex75, depth75
				}
				depth--
				add(ruleArguments, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 19 NextExpression <- <(Comma Expression)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if !_rules[ruleComma]() {
					goto l76
				}
				if !_rules[ruleExpression]() {
					goto l76
				}
				depth--
				add(ruleNextExpression, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 20 Comma <- <','> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if buffer[position] != rune(',') {
					goto l78
				}
				position++
				depth--
				add(ruleComma, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 21 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				{
					position82, tokenIndex82, depth82 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l82
					}
					position++
					goto l83
				l82:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
				}
			l83:
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l87
					}
					position++
					goto l86
				l87:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('_') {
						goto l80
					}
					position++
				}
			l86:
			l84:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					{
						position88, tokenIndex88, depth88 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l89
						}
						position++
						goto l88
					l89:
						position, tokenIndex, depth = position88, tokenIndex88, depth88
						if buffer[position] != rune('_') {
							goto l85
						}
						position++
					}
				l88:
					goto l84
				l85:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
				}
				depth--
				add(ruleInteger, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 22 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				if buffer[position] != rune('"') {
					goto l90
				}
				position++
			l92:
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					{
						position94, tokenIndex94, depth94 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l95
						}
						position++
						if buffer[position] != rune('"') {
							goto l95
						}
						position++
						goto l94
					l95:
						position, tokenIndex, depth = position94, tokenIndex94, depth94
						{
							position96, tokenIndex96, depth96 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l96
							}
							position++
							goto l93
						l96:
							position, tokenIndex, depth = position96, tokenIndex96, depth96
						}
						if !matchDot() {
							goto l93
						}
					}
				l94:
					goto l92
				l93:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
				}
				if buffer[position] != rune('"') {
					goto l90
				}
				position++
				depth--
				add(ruleString, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 23 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l100
					}
					position++
					if buffer[position] != rune('r') {
						goto l100
					}
					position++
					if buffer[position] != rune('u') {
						goto l100
					}
					position++
					if buffer[position] != rune('e') {
						goto l100
					}
					position++
					goto l99
				l100:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
					if buffer[position] != rune('f') {
						goto l97
					}
					position++
					if buffer[position] != rune('a') {
						goto l97
					}
					position++
					if buffer[position] != rune('l') {
						goto l97
					}
					position++
					if buffer[position] != rune('s') {
						goto l97
					}
					position++
					if buffer[position] != rune('e') {
						goto l97
					}
					position++
				}
			l99:
				depth--
				add(ruleBoolean, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 24 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l104
					}
					position++
					if buffer[position] != rune('i') {
						goto l104
					}
					position++
					if buffer[position] != rune('l') {
						goto l104
					}
					position++
					goto l103
				l104:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
					if buffer[position] != rune('~') {
						goto l101
					}
					position++
				}
			l103:
				depth--
				add(ruleNil, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 25 List <- <('[' Contents? ']')> */
		func() bool {
			position105, tokenIndex105, depth105 := position, tokenIndex, depth
			{
				position106 := position
				depth++
				if buffer[position] != rune('[') {
					goto l105
				}
				position++
				{
					position107, tokenIndex107, depth107 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l107
					}
					goto l108
				l107:
					position, tokenIndex, depth = position107, tokenIndex107, depth107
				}
			l108:
				if buffer[position] != rune(']') {
					goto l105
				}
				position++
				depth--
				add(ruleList, position106)
			}
			return true
		l105:
			position, tokenIndex, depth = position105, tokenIndex105, depth105
			return false
		},
		/* 26 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l109
				}
			l111:
				{
					position112, tokenIndex112, depth112 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l112
					}
					goto l111
				l112:
					position, tokenIndex, depth = position112, tokenIndex112, depth112
				}
				depth--
				add(ruleContents, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 27 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l116
					}
					goto l115
				l116:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if !_rules[ruleSimpleMerge]() {
						goto l113
					}
				}
			l115:
				depth--
				add(ruleMerge, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 28 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws Replace)? req_ws Reference)> */
		func() bool {
			position117, tokenIndex117, depth117 := position, tokenIndex, depth
			{
				position118 := position
				depth++
				if buffer[position] != rune('m') {
					goto l117
				}
				position++
				if buffer[position] != rune('e') {
					goto l117
				}
				position++
				if buffer[position] != rune('r') {
					goto l117
				}
				position++
				if buffer[position] != rune('g') {
					goto l117
				}
				position++
				if buffer[position] != rune('e') {
					goto l117
				}
				position++
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l119
					}
					if !_rules[ruleRequired]() {
						goto l119
					}
					goto l117
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
				}
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l120
					}
					if !_rules[ruleReplace]() {
						goto l120
					}
					goto l121
				l120:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
				}
			l121:
				if !_rules[rulereq_ws]() {
					goto l117
				}
				if !_rules[ruleReference]() {
					goto l117
				}
				depth--
				add(ruleRefMerge, position118)
			}
			return true
		l117:
			position, tokenIndex, depth = position117, tokenIndex117, depth117
			return false
		},
		/* 29 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required))?)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				if buffer[position] != rune('m') {
					goto l122
				}
				position++
				if buffer[position] != rune('e') {
					goto l122
				}
				position++
				if buffer[position] != rune('r') {
					goto l122
				}
				position++
				if buffer[position] != rune('g') {
					goto l122
				}
				position++
				if buffer[position] != rune('e') {
					goto l122
				}
				position++
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l124
					}
					{
						position126, tokenIndex126, depth126 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l127
						}
						goto l126
					l127:
						position, tokenIndex, depth = position126, tokenIndex126, depth126
						if !_rules[ruleRequired]() {
							goto l124
						}
					}
				l126:
					goto l125
				l124:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
				}
			l125:
				depth--
				add(ruleSimpleMerge, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 30 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				if buffer[position] != rune('r') {
					goto l128
				}
				position++
				if buffer[position] != rune('e') {
					goto l128
				}
				position++
				if buffer[position] != rune('p') {
					goto l128
				}
				position++
				if buffer[position] != rune('l') {
					goto l128
				}
				position++
				if buffer[position] != rune('a') {
					goto l128
				}
				position++
				if buffer[position] != rune('c') {
					goto l128
				}
				position++
				if buffer[position] != rune('e') {
					goto l128
				}
				position++
				depth--
				add(ruleReplace, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 31 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				if buffer[position] != rune('r') {
					goto l130
				}
				position++
				if buffer[position] != rune('e') {
					goto l130
				}
				position++
				if buffer[position] != rune('q') {
					goto l130
				}
				position++
				if buffer[position] != rune('u') {
					goto l130
				}
				position++
				if buffer[position] != rune('i') {
					goto l130
				}
				position++
				if buffer[position] != rune('r') {
					goto l130
				}
				position++
				if buffer[position] != rune('e') {
					goto l130
				}
				position++
				if buffer[position] != rune('d') {
					goto l130
				}
				position++
				depth--
				add(ruleRequired, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 32 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if buffer[position] != rune('a') {
					goto l132
				}
				position++
				if buffer[position] != rune('u') {
					goto l132
				}
				position++
				if buffer[position] != rune('t') {
					goto l132
				}
				position++
				if buffer[position] != rune('o') {
					goto l132
				}
				position++
				depth--
				add(ruleAuto, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 33 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				{
					position136, tokenIndex136, depth136 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l136
					}
					position++
					goto l137
				l136:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
				}
			l137:
				if !_rules[ruleKey]() {
					goto l134
				}
			l138:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					{
						position140, tokenIndex140, depth140 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l141
						}
						position++
						if !_rules[ruleKey]() {
							goto l141
						}
						goto l140
					l141:
						position, tokenIndex, depth = position140, tokenIndex140, depth140
						if buffer[position] != rune('.') {
							goto l139
						}
						position++
						if buffer[position] != rune('[') {
							goto l139
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l139
						}
						position++
					l142:
						{
							position143, tokenIndex143, depth143 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l143
							}
							position++
							goto l142
						l143:
							position, tokenIndex, depth = position143, tokenIndex143, depth143
						}
						if buffer[position] != rune(']') {
							goto l139
						}
						position++
					}
				l140:
					goto l138
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				depth--
				add(ruleReference, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 34 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position144, tokenIndex144, depth144 := position, tokenIndex, depth
			{
				position145 := position
				depth++
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l147
					}
					position++
					goto l146
				l147:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l148
					}
					position++
					goto l146
				l148:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l149
					}
					position++
					goto l146
				l149:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if buffer[position] != rune('_') {
						goto l144
					}
					position++
				}
			l146:
			l150:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					{
						position152, tokenIndex152, depth152 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l153
						}
						position++
						goto l152
					l153:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l154
						}
						position++
						goto l152
					l154:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l155
						}
						position++
						goto l152
					l155:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('_') {
							goto l156
						}
						position++
						goto l152
					l156:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('-') {
							goto l151
						}
						position++
					}
				l152:
					goto l150
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l157
					}
					position++
					{
						position159, tokenIndex159, depth159 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l160
						}
						position++
						goto l159
					l160:
						position, tokenIndex, depth = position159, tokenIndex159, depth159
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l161
						}
						position++
						goto l159
					l161:
						position, tokenIndex, depth = position159, tokenIndex159, depth159
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l162
						}
						position++
						goto l159
					l162:
						position, tokenIndex, depth = position159, tokenIndex159, depth159
						if buffer[position] != rune('_') {
							goto l157
						}
						position++
					}
				l159:
				l163:
					{
						position164, tokenIndex164, depth164 := position, tokenIndex, depth
						{
							position165, tokenIndex165, depth165 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l166
							}
							position++
							goto l165
						l166:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l167
							}
							position++
							goto l165
						l167:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l168
							}
							position++
							goto l165
						l168:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							if buffer[position] != rune('_') {
								goto l169
							}
							position++
							goto l165
						l169:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							if buffer[position] != rune('-') {
								goto l164
							}
							position++
						}
					l165:
						goto l163
					l164:
						position, tokenIndex, depth = position164, tokenIndex164, depth164
					}
					goto l158
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
			l158:
				depth--
				add(ruleKey, position145)
			}
			return true
		l144:
			position, tokenIndex, depth = position144, tokenIndex144, depth144
			return false
		},
		/* 35 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position171 := position
				depth++
			l172:
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					{
						position174, tokenIndex174, depth174 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l175
						}
						position++
						goto l174
					l175:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						if buffer[position] != rune('\t') {
							goto l176
						}
						position++
						goto l174
					l176:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						if buffer[position] != rune('\n') {
							goto l177
						}
						position++
						goto l174
					l177:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						if buffer[position] != rune('\r') {
							goto l173
						}
						position++
					}
				l174:
					goto l172
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
				depth--
				add(rulews, position171)
			}
			return true
		},
		/* 36 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l183
					}
					position++
					goto l182
				l183:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if buffer[position] != rune('\t') {
						goto l184
					}
					position++
					goto l182
				l184:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if buffer[position] != rune('\n') {
						goto l185
					}
					position++
					goto l182
				l185:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if buffer[position] != rune('\r') {
						goto l178
					}
					position++
				}
			l182:
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					{
						position186, tokenIndex186, depth186 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l187
						}
						position++
						goto l186
					l187:
						position, tokenIndex, depth = position186, tokenIndex186, depth186
						if buffer[position] != rune('\t') {
							goto l188
						}
						position++
						goto l186
					l188:
						position, tokenIndex, depth = position186, tokenIndex186, depth186
						if buffer[position] != rune('\n') {
							goto l189
						}
						position++
						goto l186
					l189:
						position, tokenIndex, depth = position186, tokenIndex186, depth186
						if buffer[position] != rune('\r') {
							goto l181
						}
						position++
					}
				l186:
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				depth--
				add(rulereq_ws, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
	}
	p.rules = _rules
}
