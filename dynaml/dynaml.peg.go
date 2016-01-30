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
	ruleTemplate
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
	ruleSubstitution
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
	"Template",
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
	"Substitution",
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
	rules  [59]func() bool
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
		/* 0 Dynaml <- <((Prefer / Template / Expression) !.)> */
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
					if !_rules[ruleTemplate]() {
						goto l4
					}
					goto l2
				l4:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleExpression]() {
						goto l0
					}
				}
			l2:
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !matchDot() {
						goto l5
					}
					goto l0
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
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
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
				if !_rules[rulews]() {
					goto l6
				}
				if buffer[position] != rune('p') {
					goto l6
				}
				position++
				if buffer[position] != rune('r') {
					goto l6
				}
				position++
				if buffer[position] != rune('e') {
					goto l6
				}
				position++
				if buffer[position] != rune('f') {
					goto l6
				}
				position++
				if buffer[position] != rune('e') {
					goto l6
				}
				position++
				if buffer[position] != rune('r') {
					goto l6
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l6
				}
				if !_rules[ruleExpression]() {
					goto l6
				}
				depth--
				add(rulePrefer, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 2 Template <- <(ws ('&' 't' 'e' 'm' 'p' 'l' 'a' 't' 'e') ws)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				if !_rules[rulews]() {
					goto l8
				}
				if buffer[position] != rune('&') {
					goto l8
				}
				position++
				if buffer[position] != rune('t') {
					goto l8
				}
				position++
				if buffer[position] != rune('e') {
					goto l8
				}
				position++
				if buffer[position] != rune('m') {
					goto l8
				}
				position++
				if buffer[position] != rune('p') {
					goto l8
				}
				position++
				if buffer[position] != rune('l') {
					goto l8
				}
				position++
				if buffer[position] != rune('a') {
					goto l8
				}
				position++
				if buffer[position] != rune('t') {
					goto l8
				}
				position++
				if buffer[position] != rune('e') {
					goto l8
				}
				position++
				if !_rules[rulews]() {
					goto l8
				}
				depth--
				add(ruleTemplate, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 3 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position10, tokenIndex10, depth10 := position, tokenIndex, depth
			{
				position11 := position
				depth++
				if !_rules[rulews]() {
					goto l10
				}
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l13
					}
					goto l12
				l13:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
					if !_rules[ruleLevel7]() {
						goto l10
					}
				}
			l12:
				if !_rules[rulews]() {
					goto l10
				}
				depth--
				add(ruleExpression, position11)
			}
			return true
		l10:
			position, tokenIndex, depth = position10, tokenIndex10, depth10
			return false
		},
		/* 4 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l14
				}
			l16:
				{
					position17, tokenIndex17, depth17 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l17
					}
					if !_rules[ruleOr]() {
						goto l17
					}
					goto l16
				l17:
					position, tokenIndex, depth = position17, tokenIndex17, depth17
				}
				depth--
				add(ruleLevel7, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 5 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				if buffer[position] != rune('|') {
					goto l18
				}
				position++
				if buffer[position] != rune('|') {
					goto l18
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l18
				}
				if !_rules[ruleLevel6]() {
					goto l18
				}
				depth--
				add(ruleOr, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 6 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
					if !_rules[ruleLevel5]() {
						goto l20
					}
				}
			l22:
				depth--
				add(ruleLevel6, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 7 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l24
				}
				if !_rules[rulews]() {
					goto l24
				}
				if buffer[position] != rune('?') {
					goto l24
				}
				position++
				if !_rules[ruleExpression]() {
					goto l24
				}
				if buffer[position] != rune(':') {
					goto l24
				}
				position++
				if !_rules[ruleExpression]() {
					goto l24
				}
				depth--
				add(ruleConditional, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 8 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l26
				}
			l28:
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
				}
				depth--
				add(ruleLevel5, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 9 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l30
				}
				if !_rules[ruleLevel4]() {
					goto l30
				}
				depth--
				add(ruleConcatenation, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 10 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l32
				}
			l34:
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l35
					}
					{
						position36, tokenIndex36, depth36 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l37
						}
						goto l36
					l37:
						position, tokenIndex, depth = position36, tokenIndex36, depth36
						if !_rules[ruleLogAnd]() {
							goto l35
						}
					}
				l36:
					goto l34
				l35:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
				}
				depth--
				add(ruleLevel4, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 11 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if buffer[position] != rune('-') {
					goto l38
				}
				position++
				if buffer[position] != rune('o') {
					goto l38
				}
				position++
				if buffer[position] != rune('r') {
					goto l38
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l38
				}
				if !_rules[ruleLevel3]() {
					goto l38
				}
				depth--
				add(ruleLogOr, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 12 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if buffer[position] != rune('-') {
					goto l40
				}
				position++
				if buffer[position] != rune('a') {
					goto l40
				}
				position++
				if buffer[position] != rune('n') {
					goto l40
				}
				position++
				if buffer[position] != rune('d') {
					goto l40
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l40
				}
				if !_rules[ruleLevel3]() {
					goto l40
				}
				depth--
				add(ruleLogAnd, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 13 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l42
				}
			l44:
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l45
					}
					if !_rules[ruleComparison]() {
						goto l45
					}
					goto l44
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
				depth--
				add(ruleLevel3, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 14 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l46
				}
				if !_rules[rulereq_ws]() {
					goto l46
				}
				if !_rules[ruleLevel2]() {
					goto l46
				}
				depth--
				add(ruleComparison, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 15 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l51
					}
					position++
					if buffer[position] != rune('=') {
						goto l51
					}
					position++
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('!') {
						goto l52
					}
					position++
					if buffer[position] != rune('=') {
						goto l52
					}
					position++
					goto l50
				l52:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('<') {
						goto l53
					}
					position++
					if buffer[position] != rune('=') {
						goto l53
					}
					position++
					goto l50
				l53:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('>') {
						goto l54
					}
					position++
					if buffer[position] != rune('=') {
						goto l54
					}
					position++
					goto l50
				l54:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('>') {
						goto l55
					}
					position++
					goto l50
				l55:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('<') {
						goto l56
					}
					position++
					goto l50
				l56:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('>') {
						goto l48
					}
					position++
				}
			l50:
				depth--
				add(ruleCompareOp, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 16 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l57
				}
			l59:
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l60
					}
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l62
						}
						goto l61
					l62:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if !_rules[ruleSubtraction]() {
							goto l60
						}
					}
				l61:
					goto l59
				l60:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
				}
				depth--
				add(ruleLevel2, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 17 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position63, tokenIndex63, depth63 := position, tokenIndex, depth
			{
				position64 := position
				depth++
				if buffer[position] != rune('+') {
					goto l63
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l63
				}
				if !_rules[ruleLevel1]() {
					goto l63
				}
				depth--
				add(ruleAddition, position64)
			}
			return true
		l63:
			position, tokenIndex, depth = position63, tokenIndex63, depth63
			return false
		},
		/* 18 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				if buffer[position] != rune('-') {
					goto l65
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l65
				}
				if !_rules[ruleLevel1]() {
					goto l65
				}
				depth--
				add(ruleSubtraction, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 19 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l67
				}
			l69:
				{
					position70, tokenIndex70, depth70 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l70
					}
					{
						position71, tokenIndex71, depth71 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l72
						}
						goto l71
					l72:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleDivision]() {
							goto l73
						}
						goto l71
					l73:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleModulo]() {
							goto l70
						}
					}
				l71:
					goto l69
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
				depth--
				add(ruleLevel1, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 20 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune('*') {
					goto l74
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l74
				}
				if !_rules[ruleLevel0]() {
					goto l74
				}
				depth--
				add(ruleMultiplication, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 21 Division <- <('/' req_ws Level0)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != rune('/') {
					goto l76
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l76
				}
				if !_rules[ruleLevel0]() {
					goto l76
				}
				depth--
				add(ruleDivision, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 22 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if buffer[position] != rune('%') {
					goto l78
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l78
				}
				if !_rules[ruleLevel0]() {
					goto l78
				}
				depth--
				add(ruleModulo, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 23 Level0 <- <(String / Integer / Boolean / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				{
					position82, tokenIndex82, depth82 := position, tokenIndex, depth
					if !_rules[ruleString]() {
						goto l83
					}
					goto l82
				l83:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleInteger]() {
						goto l84
					}
					goto l82
				l84:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleBoolean]() {
						goto l85
					}
					goto l82
				l85:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleNil]() {
						goto l86
					}
					goto l82
				l86:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleNot]() {
						goto l87
					}
					goto l82
				l87:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleSubstitution]() {
						goto l88
					}
					goto l82
				l88:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleMerge]() {
						goto l89
					}
					goto l82
				l89:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleAuto]() {
						goto l90
					}
					goto l82
				l90:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleLambda]() {
						goto l91
					}
					goto l82
				l91:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleChained]() {
						goto l80
					}
				}
			l82:
				depth--
				add(ruleLevel0, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 24 Chained <- <((Mapping / List / Range / ((Grouped / Reference) ChainedCall*)) (ChainedQualifiedExpression ChainedCall+)* ChainedQualifiedExpression?)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l95
					}
					goto l94
				l95:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					if !_rules[ruleList]() {
						goto l96
					}
					goto l94
				l96:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					if !_rules[ruleRange]() {
						goto l97
					}
					goto l94
				l97:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					{
						position98, tokenIndex98, depth98 := position, tokenIndex, depth
						if !_rules[ruleGrouped]() {
							goto l99
						}
						goto l98
					l99:
						position, tokenIndex, depth = position98, tokenIndex98, depth98
						if !_rules[ruleReference]() {
							goto l92
						}
					}
				l98:
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
				}
			l94:
			l102:
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l103
					}
					if !_rules[ruleChainedCall]() {
						goto l103
					}
				l104:
					{
						position105, tokenIndex105, depth105 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l105
						}
						goto l104
					l105:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
					}
					goto l102
				l103:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
				}
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l106
					}
					goto l107
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
			l107:
				depth--
				add(ruleChained, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 25 ChainedQualifiedExpression <- <('.' FollowUpRef)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				if buffer[position] != rune('.') {
					goto l108
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l108
				}
				depth--
				add(ruleChainedQualifiedExpression, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 26 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if buffer[position] != rune('(') {
					goto l110
				}
				position++
				if !_rules[ruleArguments]() {
					goto l110
				}
				if buffer[position] != rune(')') {
					goto l110
				}
				position++
				depth--
				add(ruleChainedCall, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 27 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l112
				}
			l114:
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l115
					}
					goto l114
				l115:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
				}
				depth--
				add(ruleArguments, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 28 NextExpression <- <(',' Expression)> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
				if buffer[position] != rune(',') {
					goto l116
				}
				position++
				if !_rules[ruleExpression]() {
					goto l116
				}
				depth--
				add(ruleNextExpression, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 29 Substitution <- <('*' Level0)> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				if buffer[position] != rune('*') {
					goto l118
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l118
				}
				depth--
				add(ruleSubstitution, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 30 Not <- <('!' ws Level0)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				if buffer[position] != rune('!') {
					goto l120
				}
				position++
				if !_rules[rulews]() {
					goto l120
				}
				if !_rules[ruleLevel0]() {
					goto l120
				}
				depth--
				add(ruleNot, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 31 Grouped <- <('(' Expression ')')> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				if buffer[position] != rune('(') {
					goto l122
				}
				position++
				if !_rules[ruleExpression]() {
					goto l122
				}
				if buffer[position] != rune(')') {
					goto l122
				}
				position++
				depth--
				add(ruleGrouped, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 32 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				if buffer[position] != rune('[') {
					goto l124
				}
				position++
				if !_rules[ruleExpression]() {
					goto l124
				}
				if buffer[position] != rune('.') {
					goto l124
				}
				position++
				if buffer[position] != rune('.') {
					goto l124
				}
				position++
				if !_rules[ruleExpression]() {
					goto l124
				}
				if buffer[position] != rune(']') {
					goto l124
				}
				position++
				depth--
				add(ruleRange, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 33 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l128
					}
					position++
					goto l129
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
			l129:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l126
				}
				position++
			l130:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					{
						position132, tokenIndex132, depth132 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l133
						}
						position++
						goto l132
					l133:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
						if buffer[position] != rune('_') {
							goto l131
						}
						position++
					}
				l132:
					goto l130
				l131:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
				}
				depth--
				add(ruleInteger, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 34 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if buffer[position] != rune('"') {
					goto l134
				}
				position++
			l136:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					{
						position138, tokenIndex138, depth138 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l139
						}
						position++
						if buffer[position] != rune('"') {
							goto l139
						}
						position++
						goto l138
					l139:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						{
							position140, tokenIndex140, depth140 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l140
							}
							position++
							goto l137
						l140:
							position, tokenIndex, depth = position140, tokenIndex140, depth140
						}
						if !matchDot() {
							goto l137
						}
					}
				l138:
					goto l136
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
				if buffer[position] != rune('"') {
					goto l134
				}
				position++
				depth--
				add(ruleString, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 35 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l144
					}
					position++
					if buffer[position] != rune('r') {
						goto l144
					}
					position++
					if buffer[position] != rune('u') {
						goto l144
					}
					position++
					if buffer[position] != rune('e') {
						goto l144
					}
					position++
					goto l143
				l144:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if buffer[position] != rune('f') {
						goto l141
					}
					position++
					if buffer[position] != rune('a') {
						goto l141
					}
					position++
					if buffer[position] != rune('l') {
						goto l141
					}
					position++
					if buffer[position] != rune('s') {
						goto l141
					}
					position++
					if buffer[position] != rune('e') {
						goto l141
					}
					position++
				}
			l143:
				depth--
				add(ruleBoolean, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 36 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l148
					}
					position++
					if buffer[position] != rune('i') {
						goto l148
					}
					position++
					if buffer[position] != rune('l') {
						goto l148
					}
					position++
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if buffer[position] != rune('~') {
						goto l145
					}
					position++
				}
			l147:
				depth--
				add(ruleNil, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 37 List <- <('[' Contents? ']')> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				if buffer[position] != rune('[') {
					goto l149
				}
				position++
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l151
					}
					goto l152
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
			l152:
				if buffer[position] != rune(']') {
					goto l149
				}
				position++
				depth--
				add(ruleList, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 38 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l153
				}
			l155:
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l156
					}
					goto l155
				l156:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
				}
				depth--
				add(ruleContents, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 39 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l160
					}
					goto l159
				l160:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
					if !_rules[ruleSimpleMerge]() {
						goto l157
					}
				}
			l159:
				depth--
				add(ruleMerge, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 40 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				if buffer[position] != rune('m') {
					goto l161
				}
				position++
				if buffer[position] != rune('e') {
					goto l161
				}
				position++
				if buffer[position] != rune('r') {
					goto l161
				}
				position++
				if buffer[position] != rune('g') {
					goto l161
				}
				position++
				if buffer[position] != rune('e') {
					goto l161
				}
				position++
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l163
					}
					if !_rules[ruleRequired]() {
						goto l163
					}
					goto l161
				l163:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
				}
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
				if !_rules[rulereq_ws]() {
					goto l161
				}
				if !_rules[ruleReference]() {
					goto l161
				}
				depth--
				add(ruleRefMerge, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 41 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if buffer[position] != rune('m') {
					goto l168
				}
				position++
				if buffer[position] != rune('e') {
					goto l168
				}
				position++
				if buffer[position] != rune('r') {
					goto l168
				}
				position++
				if buffer[position] != rune('g') {
					goto l168
				}
				position++
				if buffer[position] != rune('e') {
					goto l168
				}
				position++
				{
					position170, tokenIndex170, depth170 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l170
					}
					{
						position172, tokenIndex172, depth172 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l173
						}
						goto l172
					l173:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						if !_rules[ruleRequired]() {
							goto l174
						}
						goto l172
					l174:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						if !_rules[ruleOn]() {
							goto l170
						}
					}
				l172:
					goto l171
				l170:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
				}
			l171:
				depth--
				add(ruleSimpleMerge, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 42 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				if buffer[position] != rune('r') {
					goto l175
				}
				position++
				if buffer[position] != rune('e') {
					goto l175
				}
				position++
				if buffer[position] != rune('p') {
					goto l175
				}
				position++
				if buffer[position] != rune('l') {
					goto l175
				}
				position++
				if buffer[position] != rune('a') {
					goto l175
				}
				position++
				if buffer[position] != rune('c') {
					goto l175
				}
				position++
				if buffer[position] != rune('e') {
					goto l175
				}
				position++
				depth--
				add(ruleReplace, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 43 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{
				position178 := position
				depth++
				if buffer[position] != rune('r') {
					goto l177
				}
				position++
				if buffer[position] != rune('e') {
					goto l177
				}
				position++
				if buffer[position] != rune('q') {
					goto l177
				}
				position++
				if buffer[position] != rune('u') {
					goto l177
				}
				position++
				if buffer[position] != rune('i') {
					goto l177
				}
				position++
				if buffer[position] != rune('r') {
					goto l177
				}
				position++
				if buffer[position] != rune('e') {
					goto l177
				}
				position++
				if buffer[position] != rune('d') {
					goto l177
				}
				position++
				depth--
				add(ruleRequired, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 44 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if buffer[position] != rune('o') {
					goto l179
				}
				position++
				if buffer[position] != rune('n') {
					goto l179
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l179
				}
				if !_rules[ruleName]() {
					goto l179
				}
				depth--
				add(ruleOn, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 45 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if buffer[position] != rune('a') {
					goto l181
				}
				position++
				if buffer[position] != rune('u') {
					goto l181
				}
				position++
				if buffer[position] != rune('t') {
					goto l181
				}
				position++
				if buffer[position] != rune('o') {
					goto l181
				}
				position++
				depth--
				add(ruleAuto, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 46 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('m') {
					goto l183
				}
				position++
				if buffer[position] != rune('a') {
					goto l183
				}
				position++
				if buffer[position] != rune('p') {
					goto l183
				}
				position++
				if buffer[position] != rune('[') {
					goto l183
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l183
				}
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l186
					}
					goto l185
				l186:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if buffer[position] != rune('|') {
						goto l183
					}
					position++
					if !_rules[ruleExpression]() {
						goto l183
					}
				}
			l185:
				if buffer[position] != rune(']') {
					goto l183
				}
				position++
				depth--
				add(ruleMapping, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 47 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if buffer[position] != rune('l') {
					goto l187
				}
				position++
				if buffer[position] != rune('a') {
					goto l187
				}
				position++
				if buffer[position] != rune('m') {
					goto l187
				}
				position++
				if buffer[position] != rune('b') {
					goto l187
				}
				position++
				if buffer[position] != rune('d') {
					goto l187
				}
				position++
				if buffer[position] != rune('a') {
					goto l187
				}
				position++
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l190
					}
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if !_rules[ruleLambdaExpr]() {
						goto l187
					}
				}
			l189:
				depth--
				add(ruleLambda, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 48 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l191
				}
				if !_rules[ruleExpression]() {
					goto l191
				}
				depth--
				add(ruleLambdaRef, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 49 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				if !_rules[rulews]() {
					goto l193
				}
				if buffer[position] != rune('|') {
					goto l193
				}
				position++
				if !_rules[rulews]() {
					goto l193
				}
				if !_rules[ruleName]() {
					goto l193
				}
			l195:
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l196
					}
					goto l195
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				if !_rules[rulews]() {
					goto l193
				}
				if buffer[position] != rune('|') {
					goto l193
				}
				position++
				if !_rules[rulews]() {
					goto l193
				}
				if buffer[position] != rune('-') {
					goto l193
				}
				position++
				if buffer[position] != rune('>') {
					goto l193
				}
				position++
				if !_rules[ruleExpression]() {
					goto l193
				}
				depth--
				add(ruleLambdaExpr, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 50 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position197, tokenIndex197, depth197 := position, tokenIndex, depth
			{
				position198 := position
				depth++
				if !_rules[rulews]() {
					goto l197
				}
				if buffer[position] != rune(',') {
					goto l197
				}
				position++
				if !_rules[rulews]() {
					goto l197
				}
				if !_rules[ruleName]() {
					goto l197
				}
				depth--
				add(ruleNextName, position198)
			}
			return true
		l197:
			position, tokenIndex, depth = position197, tokenIndex197, depth197
			return false
		},
		/* 51 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l204
					}
					position++
					goto l203
				l204:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l205
					}
					position++
					goto l203
				l205:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l206
					}
					position++
					goto l203
				l206:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					if buffer[position] != rune('_') {
						goto l199
					}
					position++
				}
			l203:
			l201:
				{
					position202, tokenIndex202, depth202 := position, tokenIndex, depth
					{
						position207, tokenIndex207, depth207 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l208
						}
						position++
						goto l207
					l208:
						position, tokenIndex, depth = position207, tokenIndex207, depth207
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l209
						}
						position++
						goto l207
					l209:
						position, tokenIndex, depth = position207, tokenIndex207, depth207
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l210
						}
						position++
						goto l207
					l210:
						position, tokenIndex, depth = position207, tokenIndex207, depth207
						if buffer[position] != rune('_') {
							goto l202
						}
						position++
					}
				l207:
					goto l201
				l202:
					position, tokenIndex, depth = position202, tokenIndex202, depth202
				}
				depth--
				add(ruleName, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 52 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l213
					}
					position++
					goto l214
				l213:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
				}
			l214:
				if !_rules[ruleKey]() {
					goto l211
				}
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
				add(ruleReference, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 53 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l222
					}
					goto l221
				l222:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
					if !_rules[ruleIndex]() {
						goto l219
					}
				}
			l221:
			l223:
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l224
					}
					position++
					{
						position225, tokenIndex225, depth225 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l226
						}
						goto l225
					l226:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
						if !_rules[ruleIndex]() {
							goto l224
						}
					}
				l225:
					goto l223
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
				depth--
				add(ruleFollowUpRef, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 54 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
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
						goto l227
					}
					position++
				}
			l229:
			l233:
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					{
						position235, tokenIndex235, depth235 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l236
						}
						position++
						goto l235
					l236:
						position, tokenIndex, depth = position235, tokenIndex235, depth235
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l237
						}
						position++
						goto l235
					l237:
						position, tokenIndex, depth = position235, tokenIndex235, depth235
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l238
						}
						position++
						goto l235
					l238:
						position, tokenIndex, depth = position235, tokenIndex235, depth235
						if buffer[position] != rune('_') {
							goto l239
						}
						position++
						goto l235
					l239:
						position, tokenIndex, depth = position235, tokenIndex235, depth235
						if buffer[position] != rune('-') {
							goto l234
						}
						position++
					}
				l235:
					goto l233
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
				{
					position240, tokenIndex240, depth240 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l240
					}
					position++
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
							goto l240
						}
						position++
					}
				l242:
				l246:
					{
						position247, tokenIndex247, depth247 := position, tokenIndex, depth
						{
							position248, tokenIndex248, depth248 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l249
							}
							position++
							goto l248
						l249:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l250
							}
							position++
							goto l248
						l250:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l251
							}
							position++
							goto l248
						l251:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if buffer[position] != rune('_') {
								goto l252
							}
							position++
							goto l248
						l252:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if buffer[position] != rune('-') {
								goto l247
							}
							position++
						}
					l248:
						goto l246
					l247:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
					}
					goto l241
				l240:
					position, tokenIndex, depth = position240, tokenIndex240, depth240
				}
			l241:
				depth--
				add(ruleKey, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 55 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position253, tokenIndex253, depth253 := position, tokenIndex, depth
			{
				position254 := position
				depth++
				if buffer[position] != rune('[') {
					goto l253
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l253
				}
				position++
			l255:
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l256
					}
					position++
					goto l255
				l256:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
				}
				if buffer[position] != rune(']') {
					goto l253
				}
				position++
				depth--
				add(ruleIndex, position254)
			}
			return true
		l253:
			position, tokenIndex, depth = position253, tokenIndex253, depth253
			return false
		},
		/* 56 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position258 := position
				depth++
			l259:
				{
					position260, tokenIndex260, depth260 := position, tokenIndex, depth
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
							goto l260
						}
						position++
					}
				l261:
					goto l259
				l260:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
				}
				depth--
				add(rulews, position258)
			}
			return true
		},
		/* 57 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				{
					position269, tokenIndex269, depth269 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l270
					}
					position++
					goto l269
				l270:
					position, tokenIndex, depth = position269, tokenIndex269, depth269
					if buffer[position] != rune('\t') {
						goto l271
					}
					position++
					goto l269
				l271:
					position, tokenIndex, depth = position269, tokenIndex269, depth269
					if buffer[position] != rune('\n') {
						goto l272
					}
					position++
					goto l269
				l272:
					position, tokenIndex, depth = position269, tokenIndex269, depth269
					if buffer[position] != rune('\r') {
						goto l265
					}
					position++
				}
			l269:
			l267:
				{
					position268, tokenIndex268, depth268 := position, tokenIndex, depth
					{
						position273, tokenIndex273, depth273 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l274
						}
						position++
						goto l273
					l274:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if buffer[position] != rune('\t') {
							goto l275
						}
						position++
						goto l273
					l275:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if buffer[position] != rune('\n') {
							goto l276
						}
						position++
						goto l273
					l276:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if buffer[position] != rune('\r') {
							goto l268
						}
						position++
					}
				l273:
					goto l267
				l268:
					position, tokenIndex, depth = position268, tokenIndex268, depth268
				}
				depth--
				add(rulereq_ws, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
	}
	p.rules = _rules
}
