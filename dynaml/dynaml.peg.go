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
	ruleLevel7
	ruleOr
	ruleLevel6
	ruleConditional
	ruleLevel5
	ruleConcatenation
	ruleLevel4
	ruleLogOr
	ruleLogAnd
	ruleLevel3
	ruleComparison
	ruleCompareOp
	ruleLevel2
	ruleAddition
	ruleSubtraction
	ruleLevel1
	ruleMultiplication
	ruleDivision
	ruleModulo
	ruleLevel0
	ruleChained
	ruleChainedQualifiedExpression
	ruleChainedCall
	ruleArguments
	ruleNextExpression
	ruleNot
	ruleGrouped
	ruleRange
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
	ruleName
	ruleReference
	ruleFollowUpRef
	ruleKey
	ruleIndex
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
	"Level7",
	"Or",
	"Level6",
	"Conditional",
	"Level5",
	"Concatenation",
	"Level4",
	"LogOr",
	"LogAnd",
	"Level3",
	"Comparison",
	"CompareOp",
	"Level2",
	"Addition",
	"Subtraction",
	"Level1",
	"Multiplication",
	"Division",
	"Modulo",
	"Level0",
	"Chained",
	"ChainedQualifiedExpression",
	"ChainedCall",
	"Arguments",
	"NextExpression",
	"Not",
	"Grouped",
	"Range",
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
	"Name",
	"Reference",
	"FollowUpRef",
	"Key",
	"Index",
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
	rules  [57]func() bool
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
		/* 2 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				if !_rules[rulews]() {
					goto l7
				}
				{
					position9, tokenIndex9, depth9 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex, depth = position9, tokenIndex9, depth9
					if !_rules[ruleLevel7]() {
						goto l7
					}
				}
			l9:
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
		/* 3 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l11
				}
			l13:
				{
					position14, tokenIndex14, depth14 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l14
					}
					if !_rules[ruleOr]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position14, tokenIndex14, depth14
				}
				depth--
				add(ruleLevel7, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 4 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				if buffer[position] != rune('|') {
					goto l15
				}
				position++
				if buffer[position] != rune('|') {
					goto l15
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l15
				}
				if !_rules[ruleLevel6]() {
					goto l15
				}
				depth--
				add(ruleOr, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 5 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				{
					position19, tokenIndex19, depth19 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l20
					}
					goto l19
				l20:
					position, tokenIndex, depth = position19, tokenIndex19, depth19
					if !_rules[ruleLevel5]() {
						goto l17
					}
				}
			l19:
				depth--
				add(ruleLevel6, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 6 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l21
				}
				if !_rules[rulews]() {
					goto l21
				}
				if buffer[position] != rune('?') {
					goto l21
				}
				position++
				if !_rules[ruleExpression]() {
					goto l21
				}
				if buffer[position] != rune(':') {
					goto l21
				}
				position++
				if !_rules[ruleExpression]() {
					goto l21
				}
				depth--
				add(ruleConditional, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 7 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l23
				}
			l25:
				{
					position26, tokenIndex26, depth26 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l26
					}
					goto l25
				l26:
					position, tokenIndex, depth = position26, tokenIndex26, depth26
				}
				depth--
				add(ruleLevel5, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 8 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l27
				}
				if !_rules[ruleLevel4]() {
					goto l27
				}
				depth--
				add(ruleConcatenation, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 9 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l29
				}
			l31:
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l32
					}
					{
						position33, tokenIndex33, depth33 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l34
						}
						goto l33
					l34:
						position, tokenIndex, depth = position33, tokenIndex33, depth33
						if !_rules[ruleLogAnd]() {
							goto l32
						}
					}
				l33:
					goto l31
				l32:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
				}
				depth--
				add(ruleLevel4, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 10 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				if buffer[position] != rune('-') {
					goto l35
				}
				position++
				if buffer[position] != rune('o') {
					goto l35
				}
				position++
				if buffer[position] != rune('r') {
					goto l35
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l35
				}
				if !_rules[ruleLevel3]() {
					goto l35
				}
				depth--
				add(ruleLogOr, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 11 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if buffer[position] != rune('-') {
					goto l37
				}
				position++
				if buffer[position] != rune('a') {
					goto l37
				}
				position++
				if buffer[position] != rune('n') {
					goto l37
				}
				position++
				if buffer[position] != rune('d') {
					goto l37
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l37
				}
				if !_rules[ruleLevel3]() {
					goto l37
				}
				depth--
				add(ruleLogAnd, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 12 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l39
				}
			l41:
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l42
					}
					if !_rules[ruleComparison]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
				}
				depth--
				add(ruleLevel3, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 13 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l43
				}
				if !_rules[rulereq_ws]() {
					goto l43
				}
				if !_rules[ruleLevel2]() {
					goto l43
				}
				depth--
				add(ruleComparison, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 14 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l48
					}
					position++
					if buffer[position] != rune('=') {
						goto l48
					}
					position++
					goto l47
				l48:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('!') {
						goto l49
					}
					position++
					if buffer[position] != rune('=') {
						goto l49
					}
					position++
					goto l47
				l49:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('<') {
						goto l50
					}
					position++
					if buffer[position] != rune('=') {
						goto l50
					}
					position++
					goto l47
				l50:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('>') {
						goto l51
					}
					position++
					if buffer[position] != rune('=') {
						goto l51
					}
					position++
					goto l47
				l51:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('>') {
						goto l52
					}
					position++
					goto l47
				l52:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('<') {
						goto l53
					}
					position++
					goto l47
				l53:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('>') {
						goto l45
					}
					position++
				}
			l47:
				depth--
				add(ruleCompareOp, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 15 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l54
				}
			l56:
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l57
					}
					{
						position58, tokenIndex58, depth58 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l59
						}
						goto l58
					l59:
						position, tokenIndex, depth = position58, tokenIndex58, depth58
						if !_rules[ruleSubtraction]() {
							goto l57
						}
					}
				l58:
					goto l56
				l57:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
				}
				depth--
				add(ruleLevel2, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 16 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if buffer[position] != rune('+') {
					goto l60
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l60
				}
				if !_rules[ruleLevel1]() {
					goto l60
				}
				depth--
				add(ruleAddition, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 17 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				if buffer[position] != rune('-') {
					goto l62
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l62
				}
				if !_rules[ruleLevel1]() {
					goto l62
				}
				depth--
				add(ruleSubtraction, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 18 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l64
				}
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l67
					}
					{
						position68, tokenIndex68, depth68 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l69
						}
						goto l68
					l69:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if !_rules[ruleDivision]() {
							goto l70
						}
						goto l68
					l70:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if !_rules[ruleModulo]() {
							goto l67
						}
					}
				l68:
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
				depth--
				add(ruleLevel1, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 19 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				if buffer[position] != rune('*') {
					goto l71
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l71
				}
				if !_rules[ruleLevel0]() {
					goto l71
				}
				depth--
				add(ruleMultiplication, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 20 Division <- <('/' req_ws Level0)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if buffer[position] != rune('/') {
					goto l73
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l73
				}
				if !_rules[ruleLevel0]() {
					goto l73
				}
				depth--
				add(ruleDivision, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 21 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if buffer[position] != rune('%') {
					goto l75
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l75
				}
				if !_rules[ruleLevel0]() {
					goto l75
				}
				depth--
				add(ruleModulo, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 22 Level0 <- <(String / Integer / Boolean / Nil / Not / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if !_rules[ruleString]() {
						goto l80
					}
					goto l79
				l80:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleInteger]() {
						goto l81
					}
					goto l79
				l81:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleBoolean]() {
						goto l82
					}
					goto l79
				l82:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleNil]() {
						goto l83
					}
					goto l79
				l83:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleNot]() {
						goto l84
					}
					goto l79
				l84:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleMerge]() {
						goto l85
					}
					goto l79
				l85:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleAuto]() {
						goto l86
					}
					goto l79
				l86:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleLambda]() {
						goto l87
					}
					goto l79
				l87:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleChained]() {
						goto l77
					}
				}
			l79:
				depth--
				add(ruleLevel0, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 23 Chained <- <((Mapping / List / Range / ((Grouped / Reference) ChainedCall*)) (ChainedQualifiedExpression ChainedCall+)* ChainedQualifiedExpression?)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l91
					}
					goto l90
				l91:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleList]() {
						goto l92
					}
					goto l90
				l92:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleRange]() {
						goto l93
					}
					goto l90
				l93:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					{
						position94, tokenIndex94, depth94 := position, tokenIndex, depth
						if !_rules[ruleGrouped]() {
							goto l95
						}
						goto l94
					l95:
						position, tokenIndex, depth = position94, tokenIndex94, depth94
						if !_rules[ruleReference]() {
							goto l88
						}
					}
				l94:
				l96:
					{
						position97, tokenIndex97, depth97 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l97
						}
						goto l96
					l97:
						position, tokenIndex, depth = position97, tokenIndex97, depth97
					}
				}
			l90:
			l98:
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l99
					}
					if !_rules[ruleChainedCall]() {
						goto l99
					}
				l100:
					{
						position101, tokenIndex101, depth101 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l101
						}
						goto l100
					l101:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
					}
					goto l98
				l99:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
				}
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l102
					}
					goto l103
				l102:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
				}
			l103:
				depth--
				add(ruleChained, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 24 ChainedQualifiedExpression <- <('.' FollowUpRef)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if buffer[position] != rune('.') {
					goto l104
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l104
				}
				depth--
				add(ruleChainedQualifiedExpression, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 25 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				if buffer[position] != rune('(') {
					goto l106
				}
				position++
				if !_rules[ruleArguments]() {
					goto l106
				}
				if buffer[position] != rune(')') {
					goto l106
				}
				position++
				depth--
				add(ruleChainedCall, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 26 Arguments <- <(Expression NextExpression*)> */
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
				add(ruleArguments, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 27 NextExpression <- <(',' Expression)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				if buffer[position] != rune(',') {
					goto l112
				}
				position++
				if !_rules[ruleExpression]() {
					goto l112
				}
				depth--
				add(ruleNextExpression, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 28 Not <- <('!' ws Level0)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				if buffer[position] != rune('!') {
					goto l114
				}
				position++
				if !_rules[rulews]() {
					goto l114
				}
				if !_rules[ruleLevel0]() {
					goto l114
				}
				depth--
				add(ruleNot, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 29 Grouped <- <('(' Expression ')')> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
				if buffer[position] != rune('(') {
					goto l116
				}
				position++
				if !_rules[ruleExpression]() {
					goto l116
				}
				if buffer[position] != rune(')') {
					goto l116
				}
				position++
				depth--
				add(ruleGrouped, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 30 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				if buffer[position] != rune('[') {
					goto l118
				}
				position++
				if !_rules[ruleExpression]() {
					goto l118
				}
				if buffer[position] != rune('.') {
					goto l118
				}
				position++
				if buffer[position] != rune('.') {
					goto l118
				}
				position++
				if !_rules[ruleExpression]() {
					goto l118
				}
				if buffer[position] != rune(']') {
					goto l118
				}
				position++
				depth--
				add(ruleRange, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 31 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l122
					}
					position++
					goto l123
				l122:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
				}
			l123:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l120
				}
				position++
			l124:
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					{
						position126, tokenIndex126, depth126 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l127
						}
						position++
						goto l126
					l127:
						position, tokenIndex, depth = position126, tokenIndex126, depth126
						if buffer[position] != rune('_') {
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
				add(ruleInteger, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 32 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				if buffer[position] != rune('"') {
					goto l128
				}
				position++
			l130:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					{
						position132, tokenIndex132, depth132 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l133
						}
						position++
						if buffer[position] != rune('"') {
							goto l133
						}
						position++
						goto l132
					l133:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
						{
							position134, tokenIndex134, depth134 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l134
							}
							position++
							goto l131
						l134:
							position, tokenIndex, depth = position134, tokenIndex134, depth134
						}
						if !matchDot() {
							goto l131
						}
					}
				l132:
					goto l130
				l131:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
				}
				if buffer[position] != rune('"') {
					goto l128
				}
				position++
				depth--
				add(ruleString, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 33 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l138
					}
					position++
					if buffer[position] != rune('r') {
						goto l138
					}
					position++
					if buffer[position] != rune('u') {
						goto l138
					}
					position++
					if buffer[position] != rune('e') {
						goto l138
					}
					position++
					goto l137
				l138:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
					if buffer[position] != rune('f') {
						goto l135
					}
					position++
					if buffer[position] != rune('a') {
						goto l135
					}
					position++
					if buffer[position] != rune('l') {
						goto l135
					}
					position++
					if buffer[position] != rune('s') {
						goto l135
					}
					position++
					if buffer[position] != rune('e') {
						goto l135
					}
					position++
				}
			l137:
				depth--
				add(ruleBoolean, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 34 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l142
					}
					position++
					if buffer[position] != rune('i') {
						goto l142
					}
					position++
					if buffer[position] != rune('l') {
						goto l142
					}
					position++
					goto l141
				l142:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
					if buffer[position] != rune('~') {
						goto l139
					}
					position++
				}
			l141:
				depth--
				add(ruleNil, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 35 List <- <('[' Contents? ']')> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				if buffer[position] != rune('[') {
					goto l143
				}
				position++
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l145
					}
					goto l146
				l145:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
				}
			l146:
				if buffer[position] != rune(']') {
					goto l143
				}
				position++
				depth--
				add(ruleList, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 36 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l147
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l150
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(ruleContents, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 37 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[ruleSimpleMerge]() {
						goto l151
					}
				}
			l153:
				depth--
				add(ruleMerge, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 38 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if buffer[position] != rune('m') {
					goto l155
				}
				position++
				if buffer[position] != rune('e') {
					goto l155
				}
				position++
				if buffer[position] != rune('r') {
					goto l155
				}
				position++
				if buffer[position] != rune('g') {
					goto l155
				}
				position++
				if buffer[position] != rune('e') {
					goto l155
				}
				position++
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l157
					}
					if !_rules[ruleRequired]() {
						goto l157
					}
					goto l155
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l158
					}
					{
						position160, tokenIndex160, depth160 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l161
						}
						goto l160
					l161:
						position, tokenIndex, depth = position160, tokenIndex160, depth160
						if !_rules[ruleOn]() {
							goto l158
						}
					}
				l160:
					goto l159
				l158:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
				}
			l159:
				if !_rules[rulereq_ws]() {
					goto l155
				}
				if !_rules[ruleReference]() {
					goto l155
				}
				depth--
				add(ruleRefMerge, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 39 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune('m') {
					goto l162
				}
				position++
				if buffer[position] != rune('e') {
					goto l162
				}
				position++
				if buffer[position] != rune('r') {
					goto l162
				}
				position++
				if buffer[position] != rune('g') {
					goto l162
				}
				position++
				if buffer[position] != rune('e') {
					goto l162
				}
				position++
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l164
					}
					{
						position166, tokenIndex166, depth166 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l167
						}
						goto l166
					l167:
						position, tokenIndex, depth = position166, tokenIndex166, depth166
						if !_rules[ruleRequired]() {
							goto l168
						}
						goto l166
					l168:
						position, tokenIndex, depth = position166, tokenIndex166, depth166
						if !_rules[ruleOn]() {
							goto l164
						}
					}
				l166:
					goto l165
				l164:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
			l165:
				depth--
				add(ruleSimpleMerge, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 40 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				if buffer[position] != rune('r') {
					goto l169
				}
				position++
				if buffer[position] != rune('e') {
					goto l169
				}
				position++
				if buffer[position] != rune('p') {
					goto l169
				}
				position++
				if buffer[position] != rune('l') {
					goto l169
				}
				position++
				if buffer[position] != rune('a') {
					goto l169
				}
				position++
				if buffer[position] != rune('c') {
					goto l169
				}
				position++
				if buffer[position] != rune('e') {
					goto l169
				}
				position++
				depth--
				add(ruleReplace, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 41 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if buffer[position] != rune('r') {
					goto l171
				}
				position++
				if buffer[position] != rune('e') {
					goto l171
				}
				position++
				if buffer[position] != rune('q') {
					goto l171
				}
				position++
				if buffer[position] != rune('u') {
					goto l171
				}
				position++
				if buffer[position] != rune('i') {
					goto l171
				}
				position++
				if buffer[position] != rune('r') {
					goto l171
				}
				position++
				if buffer[position] != rune('e') {
					goto l171
				}
				position++
				if buffer[position] != rune('d') {
					goto l171
				}
				position++
				depth--
				add(ruleRequired, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 42 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position173, tokenIndex173, depth173 := position, tokenIndex, depth
			{
				position174 := position
				depth++
				if buffer[position] != rune('o') {
					goto l173
				}
				position++
				if buffer[position] != rune('n') {
					goto l173
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l173
				}
				if !_rules[ruleName]() {
					goto l173
				}
				depth--
				add(ruleOn, position174)
			}
			return true
		l173:
			position, tokenIndex, depth = position173, tokenIndex173, depth173
			return false
		},
		/* 43 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				if buffer[position] != rune('a') {
					goto l175
				}
				position++
				if buffer[position] != rune('u') {
					goto l175
				}
				position++
				if buffer[position] != rune('t') {
					goto l175
				}
				position++
				if buffer[position] != rune('o') {
					goto l175
				}
				position++
				depth--
				add(ruleAuto, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 44 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{
				position178 := position
				depth++
				if buffer[position] != rune('m') {
					goto l177
				}
				position++
				if buffer[position] != rune('a') {
					goto l177
				}
				position++
				if buffer[position] != rune('p') {
					goto l177
				}
				position++
				if buffer[position] != rune('[') {
					goto l177
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l177
				}
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l180
					}
					goto l179
				l180:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
					if buffer[position] != rune('|') {
						goto l177
					}
					position++
					if !_rules[ruleExpression]() {
						goto l177
					}
				}
			l179:
				if buffer[position] != rune(']') {
					goto l177
				}
				position++
				depth--
				add(ruleMapping, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 45 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if buffer[position] != rune('l') {
					goto l181
				}
				position++
				if buffer[position] != rune('a') {
					goto l181
				}
				position++
				if buffer[position] != rune('m') {
					goto l181
				}
				position++
				if buffer[position] != rune('b') {
					goto l181
				}
				position++
				if buffer[position] != rune('d') {
					goto l181
				}
				position++
				if buffer[position] != rune('a') {
					goto l181
				}
				position++
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l184
					}
					goto l183
				l184:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
					if !_rules[ruleLambdaExpr]() {
						goto l181
					}
				}
			l183:
				depth--
				add(ruleLambda, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 46 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l185
				}
				if !_rules[ruleExpression]() {
					goto l185
				}
				depth--
				add(ruleLambdaRef, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 47 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if !_rules[rulews]() {
					goto l187
				}
				if buffer[position] != rune('|') {
					goto l187
				}
				position++
				if !_rules[rulews]() {
					goto l187
				}
				if !_rules[ruleName]() {
					goto l187
				}
			l189:
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l190
					}
					goto l189
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
				if !_rules[rulews]() {
					goto l187
				}
				if buffer[position] != rune('|') {
					goto l187
				}
				position++
				if !_rules[rulews]() {
					goto l187
				}
				if buffer[position] != rune('-') {
					goto l187
				}
				position++
				if buffer[position] != rune('>') {
					goto l187
				}
				position++
				if !_rules[ruleExpression]() {
					goto l187
				}
				depth--
				add(ruleLambdaExpr, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 48 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if !_rules[rulews]() {
					goto l191
				}
				if buffer[position] != rune(',') {
					goto l191
				}
				position++
				if !_rules[rulews]() {
					goto l191
				}
				if !_rules[ruleName]() {
					goto l191
				}
				depth--
				add(ruleNextName, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 49 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l198
					}
					position++
					goto l197
				l198:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l199
					}
					position++
					goto l197
				l199:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l200
					}
					position++
					goto l197
				l200:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if buffer[position] != rune('_') {
						goto l193
					}
					position++
				}
			l197:
			l195:
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					{
						position201, tokenIndex201, depth201 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l202
						}
						position++
						goto l201
					l202:
						position, tokenIndex, depth = position201, tokenIndex201, depth201
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l203
						}
						position++
						goto l201
					l203:
						position, tokenIndex, depth = position201, tokenIndex201, depth201
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l204
						}
						position++
						goto l201
					l204:
						position, tokenIndex, depth = position201, tokenIndex201, depth201
						if buffer[position] != rune('_') {
							goto l196
						}
						position++
					}
				l201:
					goto l195
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				depth--
				add(ruleName, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 50 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l207
					}
					position++
					goto l208
				l207:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
				}
			l208:
				if !_rules[ruleKey]() {
					goto l205
				}
			l209:
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l210
					}
					position++
					{
						position211, tokenIndex211, depth211 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l212
						}
						goto l211
					l212:
						position, tokenIndex, depth = position211, tokenIndex211, depth211
						if !_rules[ruleIndex]() {
							goto l210
						}
					}
				l211:
					goto l209
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
				depth--
				add(ruleReference, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 51 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l216
					}
					goto l215
				l216:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
					if !_rules[ruleIndex]() {
						goto l213
					}
				}
			l215:
			l217:
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l218
					}
					position++
					{
						position219, tokenIndex219, depth219 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l220
						}
						goto l219
					l220:
						position, tokenIndex, depth = position219, tokenIndex219, depth219
						if !_rules[ruleIndex]() {
							goto l218
						}
					}
				l219:
					goto l217
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
				depth--
				add(ruleFollowUpRef, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 52 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l224
					}
					position++
					goto l223
				l224:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l225
					}
					position++
					goto l223
				l225:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l226
					}
					position++
					goto l223
				l226:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if buffer[position] != rune('_') {
						goto l221
					}
					position++
				}
			l223:
			l227:
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					{
						position229, tokenIndex229, depth229 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l230
						}
						position++
						goto l229
					l230:
						position, tokenIndex, depth = position229, tokenIndex229, depth229
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l231
						}
						position++
						goto l229
					l231:
						position, tokenIndex, depth = position229, tokenIndex229, depth229
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l232
						}
						position++
						goto l229
					l232:
						position, tokenIndex, depth = position229, tokenIndex229, depth229
						if buffer[position] != rune('_') {
							goto l233
						}
						position++
						goto l229
					l233:
						position, tokenIndex, depth = position229, tokenIndex229, depth229
						if buffer[position] != rune('-') {
							goto l228
						}
						position++
					}
				l229:
					goto l227
				l228:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
				}
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l234
					}
					position++
					{
						position236, tokenIndex236, depth236 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l237
						}
						position++
						goto l236
					l237:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l238
						}
						position++
						goto l236
					l238:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l239
						}
						position++
						goto l236
					l239:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('_') {
							goto l234
						}
						position++
					}
				l236:
				l240:
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						{
							position242, tokenIndex242, depth242 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l243
							}
							position++
							goto l242
						l243:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l244
							}
							position++
							goto l242
						l244:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l245
							}
							position++
							goto l242
						l245:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if buffer[position] != rune('_') {
								goto l246
							}
							position++
							goto l242
						l246:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if buffer[position] != rune('-') {
								goto l241
							}
							position++
						}
					l242:
						goto l240
					l241:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
					}
					goto l235
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
			l235:
				depth--
				add(ruleKey, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 53 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				if buffer[position] != rune('[') {
					goto l247
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l247
				}
				position++
			l249:
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l250
					}
					position++
					goto l249
				l250:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
				}
				if buffer[position] != rune(']') {
					goto l247
				}
				position++
				depth--
				add(ruleIndex, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 54 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position252 := position
				depth++
			l253:
				{
					position254, tokenIndex254, depth254 := position, tokenIndex, depth
					{
						position255, tokenIndex255, depth255 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l256
						}
						position++
						goto l255
					l256:
						position, tokenIndex, depth = position255, tokenIndex255, depth255
						if buffer[position] != rune('\t') {
							goto l257
						}
						position++
						goto l255
					l257:
						position, tokenIndex, depth = position255, tokenIndex255, depth255
						if buffer[position] != rune('\n') {
							goto l258
						}
						position++
						goto l255
					l258:
						position, tokenIndex, depth = position255, tokenIndex255, depth255
						if buffer[position] != rune('\r') {
							goto l254
						}
						position++
					}
				l255:
					goto l253
				l254:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
				}
				depth--
				add(rulews, position252)
			}
			return true
		},
		/* 55 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l264
					}
					position++
					goto l263
				l264:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
					if buffer[position] != rune('\t') {
						goto l265
					}
					position++
					goto l263
				l265:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
					if buffer[position] != rune('\n') {
						goto l266
					}
					position++
					goto l263
				l266:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
					if buffer[position] != rune('\r') {
						goto l259
					}
					position++
				}
			l263:
			l261:
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					{
						position267, tokenIndex267, depth267 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l268
						}
						position++
						goto l267
					l268:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if buffer[position] != rune('\t') {
							goto l269
						}
						position++
						goto l267
					l269:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if buffer[position] != rune('\n') {
							goto l270
						}
						position++
						goto l267
					l270:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if buffer[position] != rune('\r') {
							goto l262
						}
						position++
					}
				l267:
					goto l261
				l262:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
				}
				depth--
				add(rulereq_ws, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
	}
	p.rules = _rules
}
