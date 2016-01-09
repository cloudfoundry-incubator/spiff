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
	ruleLambda
	ruleLambdaRef
	ruleLambdaExpr
	ruleNextName
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
	"Lambda",
	"LambdaRef",
	"LambdaExpr",
	"NextName",
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
	rules  [43]func() bool
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
		/* 14 Level0 <- <(Call / Grouped / Boolean / Nil / String / Integer / List / Merge / Auto / Mapping / Lambda / Reference)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if !_rules[ruleCall]() {
						goto l47
					}
					goto l46
				l47:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleGrouped]() {
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
					if !_rules[ruleLambda]() {
						goto l57
					}
					goto l46
				l57:
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
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if buffer[position] != rune('(') {
					goto l58
				}
				position++
				if !_rules[ruleExpression]() {
					goto l58
				}
				if buffer[position] != rune(')') {
					goto l58
				}
				position++
				depth--
				add(ruleGrouped, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 16 Call <- <((Reference / Grouped) '(' Arguments ')')> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				{
					position62, tokenIndex62, depth62 := position, tokenIndex, depth
					if !_rules[ruleReference]() {
						goto l63
					}
					goto l62
				l63:
					position, tokenIndex, depth = position62, tokenIndex62, depth62
					if !_rules[ruleGrouped]() {
						goto l60
					}
				}
			l62:
				if buffer[position] != rune('(') {
					goto l60
				}
				position++
				if !_rules[ruleArguments]() {
					goto l60
				}
				if buffer[position] != rune(')') {
					goto l60
				}
				position++
				depth--
				add(ruleCall, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 17 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
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
						goto l64
					}
					position++
				}
			l68:
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					{
						position72, tokenIndex72, depth72 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l73
						}
						position++
						goto l72
					l73:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l74
						}
						position++
						goto l72
					l74:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l75
						}
						position++
						goto l72
					l75:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
						if buffer[position] != rune('_') {
							goto l67
						}
						position++
					}
				l72:
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
				depth--
				add(ruleName, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 18 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l76
				}
			l78:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l79
					}
					goto l78
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
				depth--
				add(ruleArguments, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 19 NextExpression <- <(',' Expression)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				if buffer[position] != rune(',') {
					goto l80
				}
				position++
				if !_rules[ruleExpression]() {
					goto l80
				}
				depth--
				add(ruleNextExpression, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 20 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l84
					}
					position++
					goto l85
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
			l85:
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
						goto l82
					}
					position++
				}
			l88:
			l86:
				{
					position87, tokenIndex87, depth87 := position, tokenIndex, depth
					{
						position90, tokenIndex90, depth90 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l91
						}
						position++
						goto l90
					l91:
						position, tokenIndex, depth = position90, tokenIndex90, depth90
						if buffer[position] != rune('_') {
							goto l87
						}
						position++
					}
				l90:
					goto l86
				l87:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
				}
				depth--
				add(ruleInteger, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 21 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if buffer[position] != rune('"') {
					goto l92
				}
				position++
			l94:
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					{
						position96, tokenIndex96, depth96 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l97
						}
						position++
						if buffer[position] != rune('"') {
							goto l97
						}
						position++
						goto l96
					l97:
						position, tokenIndex, depth = position96, tokenIndex96, depth96
						{
							position98, tokenIndex98, depth98 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l98
							}
							position++
							goto l95
						l98:
							position, tokenIndex, depth = position98, tokenIndex98, depth98
						}
						if !matchDot() {
							goto l95
						}
					}
				l96:
					goto l94
				l95:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
				}
				if buffer[position] != rune('"') {
					goto l92
				}
				position++
				depth--
				add(ruleString, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 22 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l102
					}
					position++
					if buffer[position] != rune('r') {
						goto l102
					}
					position++
					if buffer[position] != rune('u') {
						goto l102
					}
					position++
					if buffer[position] != rune('e') {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if buffer[position] != rune('f') {
						goto l99
					}
					position++
					if buffer[position] != rune('a') {
						goto l99
					}
					position++
					if buffer[position] != rune('l') {
						goto l99
					}
					position++
					if buffer[position] != rune('s') {
						goto l99
					}
					position++
					if buffer[position] != rune('e') {
						goto l99
					}
					position++
				}
			l101:
				depth--
				add(ruleBoolean, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 23 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l106
					}
					position++
					if buffer[position] != rune('i') {
						goto l106
					}
					position++
					if buffer[position] != rune('l') {
						goto l106
					}
					position++
					goto l105
				l106:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if buffer[position] != rune('~') {
						goto l103
					}
					position++
				}
			l105:
				depth--
				add(ruleNil, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 24 List <- <('[' Contents? ']')> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				if buffer[position] != rune('[') {
					goto l107
				}
				position++
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l109
					}
					goto l110
				l109:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
				}
			l110:
				if buffer[position] != rune(']') {
					goto l107
				}
				position++
				depth--
				add(ruleList, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 25 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l111
				}
			l113:
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l114
					}
					goto l113
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
				depth--
				add(ruleContents, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 26 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l118
					}
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleSimpleMerge]() {
						goto l115
					}
				}
			l117:
				depth--
				add(ruleMerge, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 27 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				if buffer[position] != rune('m') {
					goto l119
				}
				position++
				if buffer[position] != rune('e') {
					goto l119
				}
				position++
				if buffer[position] != rune('r') {
					goto l119
				}
				position++
				if buffer[position] != rune('g') {
					goto l119
				}
				position++
				if buffer[position] != rune('e') {
					goto l119
				}
				position++
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l121
					}
					if !_rules[ruleRequired]() {
						goto l121
					}
					goto l119
				l121:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
				}
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l122
					}
					{
						position124, tokenIndex124, depth124 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l125
						}
						goto l124
					l125:
						position, tokenIndex, depth = position124, tokenIndex124, depth124
						if !_rules[ruleOn]() {
							goto l122
						}
					}
				l124:
					goto l123
				l122:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
				}
			l123:
				if !_rules[rulereq_ws]() {
					goto l119
				}
				if !_rules[ruleReference]() {
					goto l119
				}
				depth--
				add(ruleRefMerge, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
			return false
		},
		/* 28 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				if buffer[position] != rune('m') {
					goto l126
				}
				position++
				if buffer[position] != rune('e') {
					goto l126
				}
				position++
				if buffer[position] != rune('r') {
					goto l126
				}
				position++
				if buffer[position] != rune('g') {
					goto l126
				}
				position++
				if buffer[position] != rune('e') {
					goto l126
				}
				position++
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l128
					}
					{
						position130, tokenIndex130, depth130 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l131
						}
						goto l130
					l131:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
						if !_rules[ruleRequired]() {
							goto l132
						}
						goto l130
					l132:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
						if !_rules[ruleOn]() {
							goto l128
						}
					}
				l130:
					goto l129
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
			l129:
				depth--
				add(ruleSimpleMerge, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 29 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				if buffer[position] != rune('r') {
					goto l133
				}
				position++
				if buffer[position] != rune('e') {
					goto l133
				}
				position++
				if buffer[position] != rune('p') {
					goto l133
				}
				position++
				if buffer[position] != rune('l') {
					goto l133
				}
				position++
				if buffer[position] != rune('a') {
					goto l133
				}
				position++
				if buffer[position] != rune('c') {
					goto l133
				}
				position++
				if buffer[position] != rune('e') {
					goto l133
				}
				position++
				depth--
				add(ruleReplace, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 30 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if buffer[position] != rune('r') {
					goto l135
				}
				position++
				if buffer[position] != rune('e') {
					goto l135
				}
				position++
				if buffer[position] != rune('q') {
					goto l135
				}
				position++
				if buffer[position] != rune('u') {
					goto l135
				}
				position++
				if buffer[position] != rune('i') {
					goto l135
				}
				position++
				if buffer[position] != rune('r') {
					goto l135
				}
				position++
				if buffer[position] != rune('e') {
					goto l135
				}
				position++
				if buffer[position] != rune('d') {
					goto l135
				}
				position++
				depth--
				add(ruleRequired, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 31 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if buffer[position] != rune('o') {
					goto l137
				}
				position++
				if buffer[position] != rune('n') {
					goto l137
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l137
				}
				if !_rules[ruleName]() {
					goto l137
				}
				depth--
				add(ruleOn, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 32 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				if buffer[position] != rune('a') {
					goto l139
				}
				position++
				if buffer[position] != rune('u') {
					goto l139
				}
				position++
				if buffer[position] != rune('t') {
					goto l139
				}
				position++
				if buffer[position] != rune('o') {
					goto l139
				}
				position++
				depth--
				add(ruleAuto, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 33 Mapping <- <('m' 'a' 'p' '[' Expression (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				if buffer[position] != rune('m') {
					goto l141
				}
				position++
				if buffer[position] != rune('a') {
					goto l141
				}
				position++
				if buffer[position] != rune('p') {
					goto l141
				}
				position++
				if buffer[position] != rune('[') {
					goto l141
				}
				position++
				if !_rules[ruleExpression]() {
					goto l141
				}
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if buffer[position] != rune('|') {
						goto l141
					}
					position++
					if !_rules[ruleExpression]() {
						goto l141
					}
				}
			l143:
				if buffer[position] != rune(']') {
					goto l141
				}
				position++
				depth--
				add(ruleMapping, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 34 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				if buffer[position] != rune('l') {
					goto l145
				}
				position++
				if buffer[position] != rune('a') {
					goto l145
				}
				position++
				if buffer[position] != rune('m') {
					goto l145
				}
				position++
				if buffer[position] != rune('b') {
					goto l145
				}
				position++
				if buffer[position] != rune('d') {
					goto l145
				}
				position++
				if buffer[position] != rune('a') {
					goto l145
				}
				position++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l148
					}
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[ruleLambdaExpr]() {
						goto l145
					}
				}
			l147:
				depth--
				add(ruleLambda, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 35 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l149
				}
				if !_rules[ruleExpression]() {
					goto l149
				}
				depth--
				add(ruleLambdaRef, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 36 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if !_rules[rulews]() {
					goto l151
				}
				if buffer[position] != rune('|') {
					goto l151
				}
				position++
				if !_rules[rulews]() {
					goto l151
				}
				if !_rules[ruleName]() {
					goto l151
				}
			l153:
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
				if !_rules[rulews]() {
					goto l151
				}
				if buffer[position] != rune('|') {
					goto l151
				}
				position++
				if !_rules[rulews]() {
					goto l151
				}
				if buffer[position] != rune('-') {
					goto l151
				}
				position++
				if buffer[position] != rune('>') {
					goto l151
				}
				position++
				if !_rules[ruleExpression]() {
					goto l151
				}
				depth--
				add(ruleLambdaExpr, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 37 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if !_rules[rulews]() {
					goto l155
				}
				if buffer[position] != rune(',') {
					goto l155
				}
				position++
				if !_rules[rulews]() {
					goto l155
				}
				if !_rules[ruleName]() {
					goto l155
				}
				depth--
				add(ruleNextName, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 38 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l159
					}
					position++
					goto l160
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
			l160:
				if !_rules[ruleKey]() {
					goto l157
				}
			l161:
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					{
						position163, tokenIndex163, depth163 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l164
						}
						position++
						if !_rules[ruleKey]() {
							goto l164
						}
						goto l163
					l164:
						position, tokenIndex, depth = position163, tokenIndex163, depth163
						if buffer[position] != rune('.') {
							goto l162
						}
						position++
						if buffer[position] != rune('[') {
							goto l162
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l162
						}
						position++
					l165:
						{
							position166, tokenIndex166, depth166 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l166
							}
							position++
							goto l165
						l166:
							position, tokenIndex, depth = position166, tokenIndex166, depth166
						}
						if buffer[position] != rune(']') {
							goto l162
						}
						position++
					}
				l163:
					goto l161
				l162:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
				}
				depth--
				add(ruleReference, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 39 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position167, tokenIndex167, depth167 := position, tokenIndex, depth
			{
				position168 := position
				depth++
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l170
					}
					position++
					goto l169
				l170:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l171
					}
					position++
					goto l169
				l171:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l172
					}
					position++
					goto l169
				l172:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
					if buffer[position] != rune('_') {
						goto l167
					}
					position++
				}
			l169:
			l173:
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l176
						}
						position++
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l177
						}
						position++
						goto l175
					l177:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l178
						}
						position++
						goto l175
					l178:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('_') {
							goto l179
						}
						position++
						goto l175
					l179:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('-') {
							goto l174
						}
						position++
					}
				l175:
					goto l173
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l180
					}
					position++
					{
						position182, tokenIndex182, depth182 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l183
						}
						position++
						goto l182
					l183:
						position, tokenIndex, depth = position182, tokenIndex182, depth182
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l184
						}
						position++
						goto l182
					l184:
						position, tokenIndex, depth = position182, tokenIndex182, depth182
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l185
						}
						position++
						goto l182
					l185:
						position, tokenIndex, depth = position182, tokenIndex182, depth182
						if buffer[position] != rune('_') {
							goto l180
						}
						position++
					}
				l182:
				l186:
					{
						position187, tokenIndex187, depth187 := position, tokenIndex, depth
						{
							position188, tokenIndex188, depth188 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l189
							}
							position++
							goto l188
						l189:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l190
							}
							position++
							goto l188
						l190:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l191
							}
							position++
							goto l188
						l191:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if buffer[position] != rune('_') {
								goto l192
							}
							position++
							goto l188
						l192:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if buffer[position] != rune('-') {
								goto l187
							}
							position++
						}
					l188:
						goto l186
					l187:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
					}
					goto l181
				l180:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
				}
			l181:
				depth--
				add(ruleKey, position168)
			}
			return true
		l167:
			position, tokenIndex, depth = position167, tokenIndex167, depth167
			return false
		},
		/* 40 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position194 := position
				depth++
			l195:
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					{
						position197, tokenIndex197, depth197 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l198
						}
						position++
						goto l197
					l198:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
						if buffer[position] != rune('\t') {
							goto l199
						}
						position++
						goto l197
					l199:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
						if buffer[position] != rune('\n') {
							goto l200
						}
						position++
						goto l197
					l200:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
						if buffer[position] != rune('\r') {
							goto l196
						}
						position++
					}
				l197:
					goto l195
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				depth--
				add(rulews, position194)
			}
			return true
		},
		/* 41 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l206
					}
					position++
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('\t') {
						goto l207
					}
					position++
					goto l205
				l207:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('\n') {
						goto l208
					}
					position++
					goto l205
				l208:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('\r') {
						goto l201
					}
					position++
				}
			l205:
			l203:
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					{
						position209, tokenIndex209, depth209 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l210
						}
						position++
						goto l209
					l210:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
						if buffer[position] != rune('\t') {
							goto l211
						}
						position++
						goto l209
					l211:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
						if buffer[position] != rune('\n') {
							goto l212
						}
						position++
						goto l209
					l212:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
						if buffer[position] != rune('\r') {
							goto l204
						}
						position++
					}
				l209:
					goto l203
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
				depth--
				add(rulereq_ws, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
	}
	p.rules = _rules
}
