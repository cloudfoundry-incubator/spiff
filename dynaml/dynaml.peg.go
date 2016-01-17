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
	ruleQualifiedExpression
	ruleNot
	ruleGrouped
	ruleRange
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
	"QualifiedExpression",
	"Not",
	"Grouped",
	"Range",
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
	rules  [56]func() bool
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
		/* 22 Level0 <- <(QualifiedExpression / Not / Call / Grouped / Boolean / Nil / String / Integer / List / Range / Merge / Auto / Mapping / Lambda / Reference)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if !_rules[ruleQualifiedExpression]() {
						goto l80
					}
					goto l79
				l80:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleNot]() {
						goto l81
					}
					goto l79
				l81:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleCall]() {
						goto l82
					}
					goto l79
				l82:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleGrouped]() {
						goto l83
					}
					goto l79
				l83:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleBoolean]() {
						goto l84
					}
					goto l79
				l84:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleNil]() {
						goto l85
					}
					goto l79
				l85:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleString]() {
						goto l86
					}
					goto l79
				l86:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleInteger]() {
						goto l87
					}
					goto l79
				l87:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleList]() {
						goto l88
					}
					goto l79
				l88:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleRange]() {
						goto l89
					}
					goto l79
				l89:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleMerge]() {
						goto l90
					}
					goto l79
				l90:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleAuto]() {
						goto l91
					}
					goto l79
				l91:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleMapping]() {
						goto l92
					}
					goto l79
				l92:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleLambda]() {
						goto l93
					}
					goto l79
				l93:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if !_rules[ruleReference]() {
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
		/* 23 QualifiedExpression <- <((Call / Grouped / List / Range) '.' FollowUpRef)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[ruleCall]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleGrouped]() {
						goto l98
					}
					goto l96
				l98:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleList]() {
						goto l99
					}
					goto l96
				l99:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleRange]() {
						goto l94
					}
				}
			l96:
				if buffer[position] != rune('.') {
					goto l94
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l94
				}
				depth--
				add(ruleQualifiedExpression, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 24 Not <- <('!' ws Level0)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('!') {
					goto l100
				}
				position++
				if !_rules[rulews]() {
					goto l100
				}
				if !_rules[ruleLevel0]() {
					goto l100
				}
				depth--
				add(ruleNot, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 25 Grouped <- <('(' Expression ')')> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if buffer[position] != rune('(') {
					goto l102
				}
				position++
				if !_rules[ruleExpression]() {
					goto l102
				}
				if buffer[position] != rune(')') {
					goto l102
				}
				position++
				depth--
				add(ruleGrouped, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 26 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if buffer[position] != rune('[') {
					goto l104
				}
				position++
				if !_rules[ruleExpression]() {
					goto l104
				}
				if buffer[position] != rune('.') {
					goto l104
				}
				position++
				if buffer[position] != rune('.') {
					goto l104
				}
				position++
				if !_rules[ruleExpression]() {
					goto l104
				}
				if buffer[position] != rune(']') {
					goto l104
				}
				position++
				depth--
				add(ruleRange, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 27 Call <- <((Reference / Grouped) '(' Arguments ')')> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if !_rules[ruleReference]() {
						goto l109
					}
					goto l108
				l109:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleGrouped]() {
						goto l106
					}
				}
			l108:
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
				add(ruleCall, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 28 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l115
					}
					position++
					goto l114
				l115:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l116
					}
					position++
					goto l114
				l116:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l117
					}
					position++
					goto l114
				l117:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if buffer[position] != rune('_') {
						goto l110
					}
					position++
				}
			l114:
			l112:
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
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
							goto l113
						}
						position++
					}
				l118:
					goto l112
				l113:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
				}
				depth--
				add(ruleName, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 29 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l122
				}
			l124:
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l125
					}
					goto l124
				l125:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
				}
				depth--
				add(ruleArguments, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 30 NextExpression <- <(',' Expression)> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				if buffer[position] != rune(',') {
					goto l126
				}
				position++
				if !_rules[ruleExpression]() {
					goto l126
				}
				depth--
				add(ruleNextExpression, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 31 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				{
					position130, tokenIndex130, depth130 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l130
					}
					position++
					goto l131
				l130:
					position, tokenIndex, depth = position130, tokenIndex130, depth130
				}
			l131:
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l135
					}
					position++
					goto l134
				l135:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
					if buffer[position] != rune('_') {
						goto l128
					}
					position++
				}
			l134:
			l132:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					{
						position136, tokenIndex136, depth136 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l137
						}
						position++
						goto l136
					l137:
						position, tokenIndex, depth = position136, tokenIndex136, depth136
						if buffer[position] != rune('_') {
							goto l133
						}
						position++
					}
				l136:
					goto l132
				l133:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
				}
				depth--
				add(ruleInteger, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 32 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if buffer[position] != rune('"') {
					goto l138
				}
				position++
			l140:
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					{
						position142, tokenIndex142, depth142 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l143
						}
						position++
						if buffer[position] != rune('"') {
							goto l143
						}
						position++
						goto l142
					l143:
						position, tokenIndex, depth = position142, tokenIndex142, depth142
						{
							position144, tokenIndex144, depth144 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l144
							}
							position++
							goto l141
						l144:
							position, tokenIndex, depth = position144, tokenIndex144, depth144
						}
						if !matchDot() {
							goto l141
						}
					}
				l142:
					goto l140
				l141:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
				}
				if buffer[position] != rune('"') {
					goto l138
				}
				position++
				depth--
				add(ruleString, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 33 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l148
					}
					position++
					if buffer[position] != rune('r') {
						goto l148
					}
					position++
					if buffer[position] != rune('u') {
						goto l148
					}
					position++
					if buffer[position] != rune('e') {
						goto l148
					}
					position++
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if buffer[position] != rune('f') {
						goto l145
					}
					position++
					if buffer[position] != rune('a') {
						goto l145
					}
					position++
					if buffer[position] != rune('l') {
						goto l145
					}
					position++
					if buffer[position] != rune('s') {
						goto l145
					}
					position++
					if buffer[position] != rune('e') {
						goto l145
					}
					position++
				}
			l147:
				depth--
				add(ruleBoolean, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 34 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l152
					}
					position++
					if buffer[position] != rune('i') {
						goto l152
					}
					position++
					if buffer[position] != rune('l') {
						goto l152
					}
					position++
					goto l151
				l152:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if buffer[position] != rune('~') {
						goto l149
					}
					position++
				}
			l151:
				depth--
				add(ruleNil, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 35 List <- <('[' Contents? ']')> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if buffer[position] != rune('[') {
					goto l153
				}
				position++
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l155
					}
					goto l156
				l155:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
				}
			l156:
				if buffer[position] != rune(']') {
					goto l153
				}
				position++
				depth--
				add(ruleList, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 36 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l157
				}
			l159:
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l160
					}
					goto l159
				l160:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
				}
				depth--
				add(ruleContents, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 37 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l164
					}
					goto l163
				l164:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
					if !_rules[ruleSimpleMerge]() {
						goto l161
					}
				}
			l163:
				depth--
				add(ruleMerge, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 38 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				if buffer[position] != rune('m') {
					goto l165
				}
				position++
				if buffer[position] != rune('e') {
					goto l165
				}
				position++
				if buffer[position] != rune('r') {
					goto l165
				}
				position++
				if buffer[position] != rune('g') {
					goto l165
				}
				position++
				if buffer[position] != rune('e') {
					goto l165
				}
				position++
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l167
					}
					if !_rules[ruleRequired]() {
						goto l167
					}
					goto l165
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l168
					}
					{
						position170, tokenIndex170, depth170 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l171
						}
						goto l170
					l171:
						position, tokenIndex, depth = position170, tokenIndex170, depth170
						if !_rules[ruleOn]() {
							goto l168
						}
					}
				l170:
					goto l169
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
			l169:
				if !_rules[rulereq_ws]() {
					goto l165
				}
				if !_rules[ruleReference]() {
					goto l165
				}
				depth--
				add(ruleRefMerge, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 39 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				if buffer[position] != rune('m') {
					goto l172
				}
				position++
				if buffer[position] != rune('e') {
					goto l172
				}
				position++
				if buffer[position] != rune('r') {
					goto l172
				}
				position++
				if buffer[position] != rune('g') {
					goto l172
				}
				position++
				if buffer[position] != rune('e') {
					goto l172
				}
				position++
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l174
					}
					{
						position176, tokenIndex176, depth176 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l177
						}
						goto l176
					l177:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						if !_rules[ruleRequired]() {
							goto l178
						}
						goto l176
					l178:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						if !_rules[ruleOn]() {
							goto l174
						}
					}
				l176:
					goto l175
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
			l175:
				depth--
				add(ruleSimpleMerge, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 40 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if buffer[position] != rune('r') {
					goto l179
				}
				position++
				if buffer[position] != rune('e') {
					goto l179
				}
				position++
				if buffer[position] != rune('p') {
					goto l179
				}
				position++
				if buffer[position] != rune('l') {
					goto l179
				}
				position++
				if buffer[position] != rune('a') {
					goto l179
				}
				position++
				if buffer[position] != rune('c') {
					goto l179
				}
				position++
				if buffer[position] != rune('e') {
					goto l179
				}
				position++
				depth--
				add(ruleReplace, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 41 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if buffer[position] != rune('r') {
					goto l181
				}
				position++
				if buffer[position] != rune('e') {
					goto l181
				}
				position++
				if buffer[position] != rune('q') {
					goto l181
				}
				position++
				if buffer[position] != rune('u') {
					goto l181
				}
				position++
				if buffer[position] != rune('i') {
					goto l181
				}
				position++
				if buffer[position] != rune('r') {
					goto l181
				}
				position++
				if buffer[position] != rune('e') {
					goto l181
				}
				position++
				if buffer[position] != rune('d') {
					goto l181
				}
				position++
				depth--
				add(ruleRequired, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 42 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('o') {
					goto l183
				}
				position++
				if buffer[position] != rune('n') {
					goto l183
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l183
				}
				if !_rules[ruleName]() {
					goto l183
				}
				depth--
				add(ruleOn, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 43 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if buffer[position] != rune('a') {
					goto l185
				}
				position++
				if buffer[position] != rune('u') {
					goto l185
				}
				position++
				if buffer[position] != rune('t') {
					goto l185
				}
				position++
				if buffer[position] != rune('o') {
					goto l185
				}
				position++
				depth--
				add(ruleAuto, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 44 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if buffer[position] != rune('m') {
					goto l187
				}
				position++
				if buffer[position] != rune('a') {
					goto l187
				}
				position++
				if buffer[position] != rune('p') {
					goto l187
				}
				position++
				if buffer[position] != rune('[') {
					goto l187
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l187
				}
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l190
					}
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if buffer[position] != rune('|') {
						goto l187
					}
					position++
					if !_rules[ruleExpression]() {
						goto l187
					}
				}
			l189:
				if buffer[position] != rune(']') {
					goto l187
				}
				position++
				depth--
				add(ruleMapping, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 45 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if buffer[position] != rune('l') {
					goto l191
				}
				position++
				if buffer[position] != rune('a') {
					goto l191
				}
				position++
				if buffer[position] != rune('m') {
					goto l191
				}
				position++
				if buffer[position] != rune('b') {
					goto l191
				}
				position++
				if buffer[position] != rune('d') {
					goto l191
				}
				position++
				if buffer[position] != rune('a') {
					goto l191
				}
				position++
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if !_rules[ruleLambdaExpr]() {
						goto l191
					}
				}
			l193:
				depth--
				add(ruleLambda, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 46 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l195
				}
				if !_rules[ruleExpression]() {
					goto l195
				}
				depth--
				add(ruleLambdaRef, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 47 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position197, tokenIndex197, depth197 := position, tokenIndex, depth
			{
				position198 := position
				depth++
				if !_rules[rulews]() {
					goto l197
				}
				if buffer[position] != rune('|') {
					goto l197
				}
				position++
				if !_rules[rulews]() {
					goto l197
				}
				if !_rules[ruleName]() {
					goto l197
				}
			l199:
				{
					position200, tokenIndex200, depth200 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l200
					}
					goto l199
				l200:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
				}
				if !_rules[rulews]() {
					goto l197
				}
				if buffer[position] != rune('|') {
					goto l197
				}
				position++
				if !_rules[rulews]() {
					goto l197
				}
				if buffer[position] != rune('-') {
					goto l197
				}
				position++
				if buffer[position] != rune('>') {
					goto l197
				}
				position++
				if !_rules[ruleExpression]() {
					goto l197
				}
				depth--
				add(ruleLambdaExpr, position198)
			}
			return true
		l197:
			position, tokenIndex, depth = position197, tokenIndex197, depth197
			return false
		},
		/* 48 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				if !_rules[rulews]() {
					goto l201
				}
				if buffer[position] != rune(',') {
					goto l201
				}
				position++
				if !_rules[rulews]() {
					goto l201
				}
				if !_rules[ruleName]() {
					goto l201
				}
				depth--
				add(ruleNextName, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 49 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l205
					}
					position++
					goto l206
				l205:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
				}
			l206:
				if !_rules[ruleKey]() {
					goto l203
				}
			l207:
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l208
					}
					position++
					{
						position209, tokenIndex209, depth209 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l210
						}
						goto l209
					l210:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
						if !_rules[ruleIndex]() {
							goto l208
						}
					}
				l209:
					goto l207
				l208:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
				}
				depth--
				add(ruleReference, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 50 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l214
					}
					goto l213
				l214:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
					if !_rules[ruleIndex]() {
						goto l211
					}
				}
			l213:
			l215:
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l216
					}
					position++
					{
						position217, tokenIndex217, depth217 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l218
						}
						goto l217
					l218:
						position, tokenIndex, depth = position217, tokenIndex217, depth217
						if !_rules[ruleIndex]() {
							goto l216
						}
					}
				l217:
					goto l215
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
				depth--
				add(ruleFollowUpRef, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 51 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l222
					}
					position++
					goto l221
				l222:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l223
					}
					position++
					goto l221
				l223:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l224
					}
					position++
					goto l221
				l224:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
					if buffer[position] != rune('_') {
						goto l219
					}
					position++
				}
			l221:
			l225:
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						position227, tokenIndex227, depth227 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l228
						}
						position++
						goto l227
					l228:
						position, tokenIndex, depth = position227, tokenIndex227, depth227
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l229
						}
						position++
						goto l227
					l229:
						position, tokenIndex, depth = position227, tokenIndex227, depth227
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l230
						}
						position++
						goto l227
					l230:
						position, tokenIndex, depth = position227, tokenIndex227, depth227
						if buffer[position] != rune('_') {
							goto l231
						}
						position++
						goto l227
					l231:
						position, tokenIndex, depth = position227, tokenIndex227, depth227
						if buffer[position] != rune('-') {
							goto l226
						}
						position++
					}
				l227:
					goto l225
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
				{
					position232, tokenIndex232, depth232 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l232
					}
					position++
					{
						position234, tokenIndex234, depth234 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l235
						}
						position++
						goto l234
					l235:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l236
						}
						position++
						goto l234
					l236:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l237
						}
						position++
						goto l234
					l237:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if buffer[position] != rune('_') {
							goto l232
						}
						position++
					}
				l234:
				l238:
					{
						position239, tokenIndex239, depth239 := position, tokenIndex, depth
						{
							position240, tokenIndex240, depth240 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l241
							}
							position++
							goto l240
						l241:
							position, tokenIndex, depth = position240, tokenIndex240, depth240
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l242
							}
							position++
							goto l240
						l242:
							position, tokenIndex, depth = position240, tokenIndex240, depth240
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l243
							}
							position++
							goto l240
						l243:
							position, tokenIndex, depth = position240, tokenIndex240, depth240
							if buffer[position] != rune('_') {
								goto l244
							}
							position++
							goto l240
						l244:
							position, tokenIndex, depth = position240, tokenIndex240, depth240
							if buffer[position] != rune('-') {
								goto l239
							}
							position++
						}
					l240:
						goto l238
					l239:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
					}
					goto l233
				l232:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
				}
			l233:
				depth--
				add(ruleKey, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 52 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if buffer[position] != rune('[') {
					goto l245
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l245
				}
				position++
			l247:
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l248
					}
					position++
					goto l247
				l248:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
				}
				if buffer[position] != rune(']') {
					goto l245
				}
				position++
				depth--
				add(ruleIndex, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 53 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position250 := position
				depth++
			l251:
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					{
						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l254
						}
						position++
						goto l253
					l254:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if buffer[position] != rune('\t') {
							goto l255
						}
						position++
						goto l253
					l255:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if buffer[position] != rune('\n') {
							goto l256
						}
						position++
						goto l253
					l256:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if buffer[position] != rune('\r') {
							goto l252
						}
						position++
					}
				l253:
					goto l251
				l252:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
				}
				depth--
				add(rulews, position250)
			}
			return true
		},
		/* 54 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				{
					position261, tokenIndex261, depth261 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l262
					}
					position++
					goto l261
				l262:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if buffer[position] != rune('\t') {
						goto l263
					}
					position++
					goto l261
				l263:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if buffer[position] != rune('\n') {
						goto l264
					}
					position++
					goto l261
				l264:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if buffer[position] != rune('\r') {
						goto l257
					}
					position++
				}
			l261:
			l259:
				{
					position260, tokenIndex260, depth260 := position, tokenIndex, depth
					{
						position265, tokenIndex265, depth265 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l266
						}
						position++
						goto l265
					l266:
						position, tokenIndex, depth = position265, tokenIndex265, depth265
						if buffer[position] != rune('\t') {
							goto l267
						}
						position++
						goto l265
					l267:
						position, tokenIndex, depth = position265, tokenIndex265, depth265
						if buffer[position] != rune('\n') {
							goto l268
						}
						position++
						goto l265
					l268:
						position, tokenIndex, depth = position265, tokenIndex265, depth265
						if buffer[position] != rune('\r') {
							goto l260
						}
						position++
					}
				l265:
					goto l259
				l260:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
				}
				depth--
				add(rulereq_ws, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
	}
	p.rules = _rules
}
