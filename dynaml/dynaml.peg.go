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
	ruleOn
	ruleAuto
	ruleMapping
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
	"On",
	"Auto",
	"Mapping",
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
		/* 14 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Auto / Mapping / Reference)> */
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
					if !_rules[ruleMapping]() {
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
		/* 19 NextExpression <- <(',' Expression)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if buffer[position] != rune(',') {
					goto l77
				}
				position++
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
		/* 20 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l81
					}
					position++
					goto l82
				l81:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
				}
			l82:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('_') {
						goto l79
					}
					position++
				}
			l85:
			l83:
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
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
							goto l84
						}
						position++
					}
				l87:
					goto l83
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
				depth--
				add(ruleInteger, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 21 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('"') {
					goto l89
				}
				position++
			l91:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					{
						position93, tokenIndex93, depth93 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l94
						}
						position++
						if buffer[position] != rune('"') {
							goto l94
						}
						position++
						goto l93
					l94:
						position, tokenIndex, depth = position93, tokenIndex93, depth93
						{
							position95, tokenIndex95, depth95 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l95
							}
							position++
							goto l92
						l95:
							position, tokenIndex, depth = position95, tokenIndex95, depth95
						}
						if !matchDot() {
							goto l92
						}
					}
				l93:
					goto l91
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
				if buffer[position] != rune('"') {
					goto l89
				}
				position++
				depth--
				add(ruleString, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 22 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position96, tokenIndex96, depth96 := position, tokenIndex, depth
			{
				position97 := position
				depth++
				{
					position98, tokenIndex98, depth98 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l99
					}
					position++
					if buffer[position] != rune('r') {
						goto l99
					}
					position++
					if buffer[position] != rune('u') {
						goto l99
					}
					position++
					if buffer[position] != rune('e') {
						goto l99
					}
					position++
					goto l98
				l99:
					position, tokenIndex, depth = position98, tokenIndex98, depth98
					if buffer[position] != rune('f') {
						goto l96
					}
					position++
					if buffer[position] != rune('a') {
						goto l96
					}
					position++
					if buffer[position] != rune('l') {
						goto l96
					}
					position++
					if buffer[position] != rune('s') {
						goto l96
					}
					position++
					if buffer[position] != rune('e') {
						goto l96
					}
					position++
				}
			l98:
				depth--
				add(ruleBoolean, position97)
			}
			return true
		l96:
			position, tokenIndex, depth = position96, tokenIndex96, depth96
			return false
		},
		/* 23 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l103
					}
					position++
					if buffer[position] != rune('i') {
						goto l103
					}
					position++
					if buffer[position] != rune('l') {
						goto l103
					}
					position++
					goto l102
				l103:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
					if buffer[position] != rune('~') {
						goto l100
					}
					position++
				}
			l102:
				depth--
				add(ruleNil, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 24 List <- <('[' Contents? ']')> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if buffer[position] != rune('[') {
					goto l104
				}
				position++
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l106
					}
					goto l107
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
			l107:
				if buffer[position] != rune(']') {
					goto l104
				}
				position++
				depth--
				add(ruleList, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 25 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l108
				}
			l110:
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l111
					}
					goto l110
				l111:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
				}
				depth--
				add(ruleContents, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 26 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l115
					}
					goto l114
				l115:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleSimpleMerge]() {
						goto l112
					}
				}
			l114:
				depth--
				add(ruleMerge, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 27 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
				if buffer[position] != rune('m') {
					goto l116
				}
				position++
				if buffer[position] != rune('e') {
					goto l116
				}
				position++
				if buffer[position] != rune('r') {
					goto l116
				}
				position++
				if buffer[position] != rune('g') {
					goto l116
				}
				position++
				if buffer[position] != rune('e') {
					goto l116
				}
				position++
				{
					position118, tokenIndex118, depth118 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l118
					}
					if !_rules[ruleRequired]() {
						goto l118
					}
					goto l116
				l118:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
				}
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l119
					}
					{
						position121, tokenIndex121, depth121 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l122
						}
						goto l121
					l122:
						position, tokenIndex, depth = position121, tokenIndex121, depth121
						if !_rules[ruleOn]() {
							goto l119
						}
					}
				l121:
					goto l120
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
				}
			l120:
				if !_rules[rulereq_ws]() {
					goto l116
				}
				if !_rules[ruleReference]() {
					goto l116
				}
				depth--
				add(ruleRefMerge, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 28 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
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
							goto l129
						}
						goto l127
					l129:
						position, tokenIndex, depth = position127, tokenIndex127, depth127
						if !_rules[ruleOn]() {
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
		/* 29 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
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
				if buffer[position] != rune('p') {
					goto l130
				}
				position++
				if buffer[position] != rune('l') {
					goto l130
				}
				position++
				if buffer[position] != rune('a') {
					goto l130
				}
				position++
				if buffer[position] != rune('c') {
					goto l130
				}
				position++
				if buffer[position] != rune('e') {
					goto l130
				}
				position++
				depth--
				add(ruleReplace, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 30 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if buffer[position] != rune('r') {
					goto l132
				}
				position++
				if buffer[position] != rune('e') {
					goto l132
				}
				position++
				if buffer[position] != rune('q') {
					goto l132
				}
				position++
				if buffer[position] != rune('u') {
					goto l132
				}
				position++
				if buffer[position] != rune('i') {
					goto l132
				}
				position++
				if buffer[position] != rune('r') {
					goto l132
				}
				position++
				if buffer[position] != rune('e') {
					goto l132
				}
				position++
				if buffer[position] != rune('d') {
					goto l132
				}
				position++
				depth--
				add(ruleRequired, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 31 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if buffer[position] != rune('o') {
					goto l134
				}
				position++
				if buffer[position] != rune('n') {
					goto l134
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l134
				}
				if !_rules[ruleName]() {
					goto l134
				}
				depth--
				add(ruleOn, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 32 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				if buffer[position] != rune('a') {
					goto l136
				}
				position++
				if buffer[position] != rune('u') {
					goto l136
				}
				position++
				if buffer[position] != rune('t') {
					goto l136
				}
				position++
				if buffer[position] != rune('o') {
					goto l136
				}
				position++
				depth--
				add(ruleAuto, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 33 Mapping <- <('m' 'a' 'p' '[' Expression '|' ws Name (ws ',' ws Name)? ws '|' ws ('-' '>') Expression ']')> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if buffer[position] != rune('m') {
					goto l138
				}
				position++
				if buffer[position] != rune('a') {
					goto l138
				}
				position++
				if buffer[position] != rune('p') {
					goto l138
				}
				position++
				if buffer[position] != rune('[') {
					goto l138
				}
				position++
				if !_rules[ruleExpression]() {
					goto l138
				}
				if buffer[position] != rune('|') {
					goto l138
				}
				position++
				if !_rules[rulews]() {
					goto l138
				}
				if !_rules[ruleName]() {
					goto l138
				}
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l140
					}
					if buffer[position] != rune(',') {
						goto l140
					}
					position++
					if !_rules[rulews]() {
						goto l140
					}
					if !_rules[ruleName]() {
						goto l140
					}
					goto l141
				l140:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
				}
			l141:
				if !_rules[rulews]() {
					goto l138
				}
				if buffer[position] != rune('|') {
					goto l138
				}
				position++
				if !_rules[rulews]() {
					goto l138
				}
				if buffer[position] != rune('-') {
					goto l138
				}
				position++
				if buffer[position] != rune('>') {
					goto l138
				}
				position++
				if !_rules[ruleExpression]() {
					goto l138
				}
				if buffer[position] != rune(']') {
					goto l138
				}
				position++
				depth--
				add(ruleMapping, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 34 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l144
					}
					position++
					goto l145
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
			l145:
				if !_rules[ruleKey]() {
					goto l142
				}
			l146:
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					{
						position148, tokenIndex148, depth148 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l149
						}
						position++
						if !_rules[ruleKey]() {
							goto l149
						}
						goto l148
					l149:
						position, tokenIndex, depth = position148, tokenIndex148, depth148
						if buffer[position] != rune('.') {
							goto l147
						}
						position++
						if buffer[position] != rune('[') {
							goto l147
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l147
						}
						position++
					l150:
						{
							position151, tokenIndex151, depth151 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l151
							}
							position++
							goto l150
						l151:
							position, tokenIndex, depth = position151, tokenIndex151, depth151
						}
						if buffer[position] != rune(']') {
							goto l147
						}
						position++
					}
				l148:
					goto l146
				l147:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
				}
				depth--
				add(ruleReference, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 35 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l155
					}
					position++
					goto l154
				l155:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l156
					}
					position++
					goto l154
				l156:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l157
					}
					position++
					goto l154
				l157:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
					if buffer[position] != rune('_') {
						goto l152
					}
					position++
				}
			l154:
			l158:
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					{
						position160, tokenIndex160, depth160 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l161
						}
						position++
						goto l160
					l161:
						position, tokenIndex, depth = position160, tokenIndex160, depth160
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l162
						}
						position++
						goto l160
					l162:
						position, tokenIndex, depth = position160, tokenIndex160, depth160
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l163
						}
						position++
						goto l160
					l163:
						position, tokenIndex, depth = position160, tokenIndex160, depth160
						if buffer[position] != rune('_') {
							goto l164
						}
						position++
						goto l160
					l164:
						position, tokenIndex, depth = position160, tokenIndex160, depth160
						if buffer[position] != rune('-') {
							goto l159
						}
						position++
					}
				l160:
					goto l158
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
				{
					position165, tokenIndex165, depth165 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l165
					}
					position++
					{
						position167, tokenIndex167, depth167 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l168
						}
						position++
						goto l167
					l168:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l169
						}
						position++
						goto l167
					l169:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l170
						}
						position++
						goto l167
					l170:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
						if buffer[position] != rune('_') {
							goto l165
						}
						position++
					}
				l167:
				l171:
					{
						position172, tokenIndex172, depth172 := position, tokenIndex, depth
						{
							position173, tokenIndex173, depth173 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l174
							}
							position++
							goto l173
						l174:
							position, tokenIndex, depth = position173, tokenIndex173, depth173
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l175
							}
							position++
							goto l173
						l175:
							position, tokenIndex, depth = position173, tokenIndex173, depth173
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l176
							}
							position++
							goto l173
						l176:
							position, tokenIndex, depth = position173, tokenIndex173, depth173
							if buffer[position] != rune('_') {
								goto l177
							}
							position++
							goto l173
						l177:
							position, tokenIndex, depth = position173, tokenIndex173, depth173
							if buffer[position] != rune('-') {
								goto l172
							}
							position++
						}
					l173:
						goto l171
					l172:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
					}
					goto l166
				l165:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
				}
			l166:
				depth--
				add(ruleKey, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 36 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position179 := position
				depth++
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
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
							goto l181
						}
						position++
					}
				l182:
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				depth--
				add(rulews, position179)
			}
			return true
		},
		/* 37 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l191
					}
					position++
					goto l190
				l191:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
					if buffer[position] != rune('\t') {
						goto l192
					}
					position++
					goto l190
				l192:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
					if buffer[position] != rune('\n') {
						goto l193
					}
					position++
					goto l190
				l193:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
					if buffer[position] != rune('\r') {
						goto l186
					}
					position++
				}
			l190:
			l188:
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					{
						position194, tokenIndex194, depth194 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l195
						}
						position++
						goto l194
					l195:
						position, tokenIndex, depth = position194, tokenIndex194, depth194
						if buffer[position] != rune('\t') {
							goto l196
						}
						position++
						goto l194
					l196:
						position, tokenIndex, depth = position194, tokenIndex194, depth194
						if buffer[position] != rune('\n') {
							goto l197
						}
						position++
						goto l194
					l197:
						position, tokenIndex, depth = position194, tokenIndex194, depth194
						if buffer[position] != rune('\r') {
							goto l189
						}
						position++
					}
				l194:
					goto l188
				l189:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
				}
				depth--
				add(rulereq_ws, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
	}
	p.rules = _rules
}
