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
	rules  [36]func() bool
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
		/* 9 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
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
							goto l33
						}
						goto l31
					l33:
						position, tokenIndex, depth = position31, tokenIndex31, depth31
						if !_rules[ruleModulo]() {
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
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if buffer[position] != rune('*') {
					goto l34
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l34
				}
				if !_rules[ruleLevel0]() {
					goto l34
				}
				depth--
				add(ruleMultiplication, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 11 Division <- <('/' req_ws Level0)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if buffer[position] != rune('/') {
					goto l36
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l36
				}
				if !_rules[ruleLevel0]() {
					goto l36
				}
				depth--
				add(ruleDivision, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 12 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if buffer[position] != rune('%') {
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
				add(ruleModulo, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 13 Level0 <- <(Grouped / Call / Boolean / Nil / String / Integer / List / Merge / Auto / Reference)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleGrouped]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleCall]() {
						goto l44
					}
					goto l42
				l44:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleBoolean]() {
						goto l45
					}
					goto l42
				l45:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleNil]() {
						goto l46
					}
					goto l42
				l46:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleString]() {
						goto l47
					}
					goto l42
				l47:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleInteger]() {
						goto l48
					}
					goto l42
				l48:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleList]() {
						goto l49
					}
					goto l42
				l49:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleMerge]() {
						goto l50
					}
					goto l42
				l50:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleAuto]() {
						goto l51
					}
					goto l42
				l51:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !_rules[ruleReference]() {
						goto l40
					}
				}
			l42:
				depth--
				add(ruleLevel0, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 14 Grouped <- <('(' Expression ')')> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if buffer[position] != rune('(') {
					goto l52
				}
				position++
				if !_rules[ruleExpression]() {
					goto l52
				}
				if buffer[position] != rune(')') {
					goto l52
				}
				position++
				depth--
				add(ruleGrouped, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 15 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if !_rules[ruleName]() {
					goto l54
				}
				if buffer[position] != rune('(') {
					goto l54
				}
				position++
				if !_rules[ruleArguments]() {
					goto l54
				}
				if buffer[position] != rune(')') {
					goto l54
				}
				position++
				depth--
				add(ruleCall, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 16 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
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
						goto l56
					}
					position++
				}
			l60:
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
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
							goto l59
						}
						position++
					}
				l64:
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
				depth--
				add(ruleName, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 17 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l68
				}
			l70:
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l71
					}
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				depth--
				add(ruleArguments, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 18 NextExpression <- <(Comma Expression)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if !_rules[ruleComma]() {
					goto l72
				}
				if !_rules[ruleExpression]() {
					goto l72
				}
				depth--
				add(ruleNextExpression, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 19 Comma <- <','> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune(',') {
					goto l74
				}
				position++
				depth--
				add(ruleComma, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 20 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l78
					}
					position++
					goto l79
				l78:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
				}
			l79:
				{
					position82, tokenIndex82, depth82 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l83
					}
					position++
					goto l82
				l83:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('_') {
						goto l76
					}
					position++
				}
			l82:
			l80:
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					{
						position84, tokenIndex84, depth84 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l85
						}
						position++
						goto l84
					l85:
						position, tokenIndex, depth = position84, tokenIndex84, depth84
						if buffer[position] != rune('_') {
							goto l81
						}
						position++
					}
				l84:
					goto l80
				l81:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
				}
				depth--
				add(ruleInteger, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 21 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				if buffer[position] != rune('"') {
					goto l86
				}
				position++
			l88:
				{
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					{
						position90, tokenIndex90, depth90 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l91
						}
						position++
						if buffer[position] != rune('"') {
							goto l91
						}
						position++
						goto l90
					l91:
						position, tokenIndex, depth = position90, tokenIndex90, depth90
						{
							position92, tokenIndex92, depth92 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l92
							}
							position++
							goto l89
						l92:
							position, tokenIndex, depth = position92, tokenIndex92, depth92
						}
						if !matchDot() {
							goto l89
						}
					}
				l90:
					goto l88
				l89:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
				}
				if buffer[position] != rune('"') {
					goto l86
				}
				position++
				depth--
				add(ruleString, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 22 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l96
					}
					position++
					if buffer[position] != rune('r') {
						goto l96
					}
					position++
					if buffer[position] != rune('u') {
						goto l96
					}
					position++
					if buffer[position] != rune('e') {
						goto l96
					}
					position++
					goto l95
				l96:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if buffer[position] != rune('f') {
						goto l93
					}
					position++
					if buffer[position] != rune('a') {
						goto l93
					}
					position++
					if buffer[position] != rune('l') {
						goto l93
					}
					position++
					if buffer[position] != rune('s') {
						goto l93
					}
					position++
					if buffer[position] != rune('e') {
						goto l93
					}
					position++
				}
			l95:
				depth--
				add(ruleBoolean, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 23 Nil <- <('n' 'i' 'l')> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if buffer[position] != rune('n') {
					goto l97
				}
				position++
				if buffer[position] != rune('i') {
					goto l97
				}
				position++
				if buffer[position] != rune('l') {
					goto l97
				}
				position++
				depth--
				add(ruleNil, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 24 List <- <('[' Contents? ']')> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if buffer[position] != rune('[') {
					goto l99
				}
				position++
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l101
					}
					goto l102
				l101:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
				}
			l102:
				if buffer[position] != rune(']') {
					goto l99
				}
				position++
				depth--
				add(ruleList, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 25 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l103
				}
			l105:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l106
					}
					goto l105
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
				depth--
				add(ruleContents, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 26 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleSimpleMerge]() {
						goto l107
					}
				}
			l109:
				depth--
				add(ruleMerge, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 27 RefMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws Replace)? req_ws Reference)> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if buffer[position] != rune('m') {
					goto l111
				}
				position++
				if buffer[position] != rune('e') {
					goto l111
				}
				position++
				if buffer[position] != rune('r') {
					goto l111
				}
				position++
				if buffer[position] != rune('g') {
					goto l111
				}
				position++
				if buffer[position] != rune('e') {
					goto l111
				}
				position++
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l113
					}
					if !_rules[ruleReplace]() {
						goto l113
					}
					goto l114
				l113:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
				}
			l114:
				if !_rules[rulereq_ws]() {
					goto l111
				}
				if !_rules[ruleReference]() {
					goto l111
				}
				depth--
				add(ruleRefMerge, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 28 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws Replace)?)> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				if buffer[position] != rune('m') {
					goto l115
				}
				position++
				if buffer[position] != rune('e') {
					goto l115
				}
				position++
				if buffer[position] != rune('r') {
					goto l115
				}
				position++
				if buffer[position] != rune('g') {
					goto l115
				}
				position++
				if buffer[position] != rune('e') {
					goto l115
				}
				position++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l117
					}
					if !_rules[ruleReplace]() {
						goto l117
					}
					goto l118
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
			l118:
				depth--
				add(ruleSimpleMerge, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 29 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				if buffer[position] != rune('r') {
					goto l119
				}
				position++
				if buffer[position] != rune('e') {
					goto l119
				}
				position++
				if buffer[position] != rune('p') {
					goto l119
				}
				position++
				if buffer[position] != rune('l') {
					goto l119
				}
				position++
				if buffer[position] != rune('a') {
					goto l119
				}
				position++
				if buffer[position] != rune('c') {
					goto l119
				}
				position++
				if buffer[position] != rune('e') {
					goto l119
				}
				position++
				depth--
				add(ruleReplace, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
			return false
		},
		/* 30 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				if buffer[position] != rune('a') {
					goto l121
				}
				position++
				if buffer[position] != rune('u') {
					goto l121
				}
				position++
				if buffer[position] != rune('t') {
					goto l121
				}
				position++
				if buffer[position] != rune('o') {
					goto l121
				}
				position++
				depth--
				add(ruleAuto, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 31 Reference <- <('.'? Key (('.' Key) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l125
					}
					position++
					goto l126
				l125:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
				}
			l126:
				if !_rules[ruleKey]() {
					goto l123
				}
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					{
						position129, tokenIndex129, depth129 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l130
						}
						position++
						if !_rules[ruleKey]() {
							goto l130
						}
						goto l129
					l130:
						position, tokenIndex, depth = position129, tokenIndex129, depth129
						if buffer[position] != rune('.') {
							goto l128
						}
						position++
						if buffer[position] != rune('[') {
							goto l128
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l128
						}
						position++
					l131:
						{
							position132, tokenIndex132, depth132 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l132
							}
							position++
							goto l131
						l132:
							position, tokenIndex, depth = position132, tokenIndex132, depth132
						}
						if buffer[position] != rune(']') {
							goto l128
						}
						position++
					}
				l129:
					goto l127
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
				depth--
				add(ruleReference, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 32 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l136
					}
					position++
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l137
					}
					position++
					goto l135
				l137:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l138
					}
					position++
					goto l135
				l138:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if buffer[position] != rune('_') {
						goto l133
					}
					position++
				}
			l135:
			l139:
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					{
						position141, tokenIndex141, depth141 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l142
						}
						position++
						goto l141
					l142:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l143
						}
						position++
						goto l141
					l143:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l144
						}
						position++
						goto l141
					l144:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
						if buffer[position] != rune('_') {
							goto l145
						}
						position++
						goto l141
					l145:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
						if buffer[position] != rune('-') {
							goto l140
						}
						position++
					}
				l141:
					goto l139
				l140:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
				}
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l146
					}
					position++
					{
						position148, tokenIndex148, depth148 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l149
						}
						position++
						goto l148
					l149:
						position, tokenIndex, depth = position148, tokenIndex148, depth148
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l150
						}
						position++
						goto l148
					l150:
						position, tokenIndex, depth = position148, tokenIndex148, depth148
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l151
						}
						position++
						goto l148
					l151:
						position, tokenIndex, depth = position148, tokenIndex148, depth148
						if buffer[position] != rune('_') {
							goto l146
						}
						position++
					}
				l148:
				l152:
					{
						position153, tokenIndex153, depth153 := position, tokenIndex, depth
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
								goto l158
							}
							position++
							goto l154
						l158:
							position, tokenIndex, depth = position154, tokenIndex154, depth154
							if buffer[position] != rune('-') {
								goto l153
							}
							position++
						}
					l154:
						goto l152
					l153:
						position, tokenIndex, depth = position153, tokenIndex153, depth153
					}
					goto l147
				l146:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
				}
			l147:
				depth--
				add(ruleKey, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 33 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position160 := position
				depth++
			l161:
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					{
						position163, tokenIndex163, depth163 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l164
						}
						position++
						goto l163
					l164:
						position, tokenIndex, depth = position163, tokenIndex163, depth163
						if buffer[position] != rune('\t') {
							goto l165
						}
						position++
						goto l163
					l165:
						position, tokenIndex, depth = position163, tokenIndex163, depth163
						if buffer[position] != rune('\n') {
							goto l166
						}
						position++
						goto l163
					l166:
						position, tokenIndex, depth = position163, tokenIndex163, depth163
						if buffer[position] != rune('\r') {
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
				add(rulews, position160)
			}
			return true
		},
		/* 34 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position167, tokenIndex167, depth167 := position, tokenIndex, depth
			{
				position168 := position
				depth++
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l172
					}
					position++
					goto l171
				l172:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
					if buffer[position] != rune('\t') {
						goto l173
					}
					position++
					goto l171
				l173:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
					if buffer[position] != rune('\n') {
						goto l174
					}
					position++
					goto l171
				l174:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
					if buffer[position] != rune('\r') {
						goto l167
					}
					position++
				}
			l171:
			l169:
				{
					position170, tokenIndex170, depth170 := position, tokenIndex, depth
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l176
						}
						position++
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('\t') {
							goto l177
						}
						position++
						goto l175
					l177:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('\n') {
							goto l178
						}
						position++
						goto l175
					l178:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('\r') {
							goto l170
						}
						position++
					}
				l175:
					goto l169
				l170:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
				}
				depth--
				add(rulereq_ws, position168)
			}
			return true
		l167:
			position, tokenIndex, depth = position167, tokenIndex167, depth167
			return false
		},
	}
	p.rules = _rules
}
