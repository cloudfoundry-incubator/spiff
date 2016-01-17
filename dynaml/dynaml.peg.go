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
		/* 2 Expression <- <(ws Level7 ws)> */
		func() bool {
			position7, tokenIndex7, depth7 := position, tokenIndex, depth
			{
				position8 := position
				depth++
				if !_rules[rulews]() {
					goto l7
				}
				if !_rules[ruleLevel7]() {
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
		/* 3 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				if !_rules[ruleLevel6]() {
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
				add(ruleLevel7, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 4 Or <- <('|' '|' req_ws Level6)> */
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
				if !_rules[ruleLevel6]() {
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
		/* 5 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				{
					position17, tokenIndex17, depth17 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l18
					}
					goto l17
				l18:
					position, tokenIndex, depth = position17, tokenIndex17, depth17
					if !_rules[ruleLevel5]() {
						goto l15
					}
				}
			l17:
				depth--
				add(ruleLevel6, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 6 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position19, tokenIndex19, depth19 := position, tokenIndex, depth
			{
				position20 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l19
				}
				if !_rules[rulews]() {
					goto l19
				}
				if buffer[position] != rune('?') {
					goto l19
				}
				position++
				if !_rules[ruleExpression]() {
					goto l19
				}
				if buffer[position] != rune(':') {
					goto l19
				}
				position++
				if !_rules[ruleExpression]() {
					goto l19
				}
				depth--
				add(ruleConditional, position20)
			}
			return true
		l19:
			position, tokenIndex, depth = position19, tokenIndex19, depth19
			return false
		},
		/* 7 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l21
				}
			l23:
				{
					position24, tokenIndex24, depth24 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l24
					}
					goto l23
				l24:
					position, tokenIndex, depth = position24, tokenIndex24, depth24
				}
				depth--
				add(ruleLevel5, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 8 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l25
				}
				if !_rules[ruleLevel4]() {
					goto l25
				}
				depth--
				add(ruleConcatenation, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 9 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if !_rules[ruleLevel3]() {
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
						if !_rules[ruleLogOr]() {
							goto l32
						}
						goto l31
					l32:
						position, tokenIndex, depth = position31, tokenIndex31, depth31
						if !_rules[ruleLogAnd]() {
							goto l30
						}
					}
				l31:
					goto l29
				l30:
					position, tokenIndex, depth = position30, tokenIndex30, depth30
				}
				depth--
				add(ruleLevel4, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 10 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				if buffer[position] != rune('-') {
					goto l33
				}
				position++
				if buffer[position] != rune('o') {
					goto l33
				}
				position++
				if buffer[position] != rune('r') {
					goto l33
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l33
				}
				if !_rules[ruleLevel3]() {
					goto l33
				}
				depth--
				add(ruleLogOr, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 11 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				if buffer[position] != rune('-') {
					goto l35
				}
				position++
				if buffer[position] != rune('a') {
					goto l35
				}
				position++
				if buffer[position] != rune('n') {
					goto l35
				}
				position++
				if buffer[position] != rune('d') {
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
				add(ruleLogAnd, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 12 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l37
				}
			l39:
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l40
					}
					if !_rules[ruleComparison]() {
						goto l40
					}
					goto l39
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
				depth--
				add(ruleLevel3, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 13 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l41
				}
				if !_rules[rulereq_ws]() {
					goto l41
				}
				if !_rules[ruleLevel2]() {
					goto l41
				}
				depth--
				add(ruleComparison, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 14 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l46
					}
					position++
					if buffer[position] != rune('=') {
						goto l46
					}
					position++
					goto l45
				l46:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('!') {
						goto l47
					}
					position++
					if buffer[position] != rune('=') {
						goto l47
					}
					position++
					goto l45
				l47:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('<') {
						goto l48
					}
					position++
					if buffer[position] != rune('=') {
						goto l48
					}
					position++
					goto l45
				l48:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('>') {
						goto l49
					}
					position++
					if buffer[position] != rune('=') {
						goto l49
					}
					position++
					goto l45
				l49:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('>') {
						goto l50
					}
					position++
					goto l45
				l50:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('<') {
						goto l51
					}
					position++
					goto l45
				l51:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
					if buffer[position] != rune('>') {
						goto l43
					}
					position++
				}
			l45:
				depth--
				add(ruleCompareOp, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 15 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l52
				}
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l55
					}
					{
						position56, tokenIndex56, depth56 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l57
						}
						goto l56
					l57:
						position, tokenIndex, depth = position56, tokenIndex56, depth56
						if !_rules[ruleSubtraction]() {
							goto l55
						}
					}
				l56:
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				depth--
				add(ruleLevel2, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 16 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if buffer[position] != rune('+') {
					goto l58
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l58
				}
				if !_rules[ruleLevel1]() {
					goto l58
				}
				depth--
				add(ruleAddition, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 17 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if buffer[position] != rune('-') {
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
				add(ruleSubtraction, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 18 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l62
				}
			l64:
				{
					position65, tokenIndex65, depth65 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l65
					}
					{
						position66, tokenIndex66, depth66 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l67
						}
						goto l66
					l67:
						position, tokenIndex, depth = position66, tokenIndex66, depth66
						if !_rules[ruleDivision]() {
							goto l68
						}
						goto l66
					l68:
						position, tokenIndex, depth = position66, tokenIndex66, depth66
						if !_rules[ruleModulo]() {
							goto l65
						}
					}
				l66:
					goto l64
				l65:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
				}
				depth--
				add(ruleLevel1, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 19 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				if buffer[position] != rune('*') {
					goto l69
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l69
				}
				if !_rules[ruleLevel0]() {
					goto l69
				}
				depth--
				add(ruleMultiplication, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 20 Division <- <('/' req_ws Level0)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				if buffer[position] != rune('/') {
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
				add(ruleDivision, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 21 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if buffer[position] != rune('%') {
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
				add(ruleModulo, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 22 Level0 <- <(QualifiedExpression / Not / Call / Grouped / Boolean / Nil / String / Integer / List / Range / Merge / Auto / Mapping / Lambda / Reference)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					if !_rules[ruleQualifiedExpression]() {
						goto l78
					}
					goto l77
				l78:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleNot]() {
						goto l79
					}
					goto l77
				l79:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleCall]() {
						goto l80
					}
					goto l77
				l80:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleGrouped]() {
						goto l81
					}
					goto l77
				l81:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleBoolean]() {
						goto l82
					}
					goto l77
				l82:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleNil]() {
						goto l83
					}
					goto l77
				l83:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleString]() {
						goto l84
					}
					goto l77
				l84:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleInteger]() {
						goto l85
					}
					goto l77
				l85:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleList]() {
						goto l86
					}
					goto l77
				l86:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleRange]() {
						goto l87
					}
					goto l77
				l87:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleMerge]() {
						goto l88
					}
					goto l77
				l88:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleAuto]() {
						goto l89
					}
					goto l77
				l89:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleMapping]() {
						goto l90
					}
					goto l77
				l90:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleLambda]() {
						goto l91
					}
					goto l77
				l91:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if !_rules[ruleReference]() {
						goto l75
					}
				}
			l77:
				depth--
				add(ruleLevel0, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 23 QualifiedExpression <- <((Call / Grouped / List / Range) '.' FollowUpRef)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					if !_rules[ruleCall]() {
						goto l95
					}
					goto l94
				l95:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					if !_rules[ruleGrouped]() {
						goto l96
					}
					goto l94
				l96:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					if !_rules[ruleList]() {
						goto l97
					}
					goto l94
				l97:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
					if !_rules[ruleRange]() {
						goto l92
					}
				}
			l94:
				if buffer[position] != rune('.') {
					goto l92
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l92
				}
				depth--
				add(ruleQualifiedExpression, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 24 Not <- <('!' ws Level0)> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if buffer[position] != rune('!') {
					goto l98
				}
				position++
				if !_rules[rulews]() {
					goto l98
				}
				if !_rules[ruleLevel0]() {
					goto l98
				}
				depth--
				add(ruleNot, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 25 Grouped <- <('(' Expression ')')> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('(') {
					goto l100
				}
				position++
				if !_rules[ruleExpression]() {
					goto l100
				}
				if buffer[position] != rune(')') {
					goto l100
				}
				position++
				depth--
				add(ruleGrouped, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 26 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if buffer[position] != rune('[') {
					goto l102
				}
				position++
				if !_rules[ruleExpression]() {
					goto l102
				}
				if buffer[position] != rune('.') {
					goto l102
				}
				position++
				if buffer[position] != rune('.') {
					goto l102
				}
				position++
				if !_rules[ruleExpression]() {
					goto l102
				}
				if buffer[position] != rune(']') {
					goto l102
				}
				position++
				depth--
				add(ruleRange, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 27 Call <- <((Reference / Grouped) '(' Arguments ')')> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleReference]() {
						goto l107
					}
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleGrouped]() {
						goto l104
					}
				}
			l106:
				if buffer[position] != rune('(') {
					goto l104
				}
				position++
				if !_rules[ruleArguments]() {
					goto l104
				}
				if buffer[position] != rune(')') {
					goto l104
				}
				position++
				depth--
				add(ruleCall, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 28 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				{
					position112, tokenIndex112, depth112 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l113
					}
					position++
					goto l112
				l113:
					position, tokenIndex, depth = position112, tokenIndex112, depth112
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l114
					}
					position++
					goto l112
				l114:
					position, tokenIndex, depth = position112, tokenIndex112, depth112
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l115
					}
					position++
					goto l112
				l115:
					position, tokenIndex, depth = position112, tokenIndex112, depth112
					if buffer[position] != rune('_') {
						goto l108
					}
					position++
				}
			l112:
			l110:
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					{
						position116, tokenIndex116, depth116 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l117
						}
						position++
						goto l116
					l117:
						position, tokenIndex, depth = position116, tokenIndex116, depth116
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l118
						}
						position++
						goto l116
					l118:
						position, tokenIndex, depth = position116, tokenIndex116, depth116
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l119
						}
						position++
						goto l116
					l119:
						position, tokenIndex, depth = position116, tokenIndex116, depth116
						if buffer[position] != rune('_') {
							goto l111
						}
						position++
					}
				l116:
					goto l110
				l111:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
				}
				depth--
				add(ruleName, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 29 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l120
				}
			l122:
				{
					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l123
					}
					goto l122
				l123:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
				}
				depth--
				add(ruleArguments, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 30 NextExpression <- <(',' Expression)> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				if buffer[position] != rune(',') {
					goto l124
				}
				position++
				if !_rules[ruleExpression]() {
					goto l124
				}
				depth--
				add(ruleNextExpression, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 31 Integer <- <('-'? ([0-9] / '_')+)> */
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
						goto l126
					}
					position++
				}
			l132:
			l130:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
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
							goto l131
						}
						position++
					}
				l134:
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
		/* 32 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				if buffer[position] != rune('"') {
					goto l136
				}
				position++
			l138:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					{
						position140, tokenIndex140, depth140 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l141
						}
						position++
						if buffer[position] != rune('"') {
							goto l141
						}
						position++
						goto l140
					l141:
						position, tokenIndex, depth = position140, tokenIndex140, depth140
						{
							position142, tokenIndex142, depth142 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l142
							}
							position++
							goto l139
						l142:
							position, tokenIndex, depth = position142, tokenIndex142, depth142
						}
						if !matchDot() {
							goto l139
						}
					}
				l140:
					goto l138
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				if buffer[position] != rune('"') {
					goto l136
				}
				position++
				depth--
				add(ruleString, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 33 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l146
					}
					position++
					if buffer[position] != rune('r') {
						goto l146
					}
					position++
					if buffer[position] != rune('u') {
						goto l146
					}
					position++
					if buffer[position] != rune('e') {
						goto l146
					}
					position++
					goto l145
				l146:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('f') {
						goto l143
					}
					position++
					if buffer[position] != rune('a') {
						goto l143
					}
					position++
					if buffer[position] != rune('l') {
						goto l143
					}
					position++
					if buffer[position] != rune('s') {
						goto l143
					}
					position++
					if buffer[position] != rune('e') {
						goto l143
					}
					position++
				}
			l145:
				depth--
				add(ruleBoolean, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 34 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l150
					}
					position++
					if buffer[position] != rune('i') {
						goto l150
					}
					position++
					if buffer[position] != rune('l') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if buffer[position] != rune('~') {
						goto l147
					}
					position++
				}
			l149:
				depth--
				add(ruleNil, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 35 List <- <('[' Contents? ']')> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if buffer[position] != rune('[') {
					goto l151
				}
				position++
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l153
					}
					goto l154
				l153:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
				}
			l154:
				if buffer[position] != rune(']') {
					goto l151
				}
				position++
				depth--
				add(ruleList, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 36 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l155
				}
			l157:
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l158
					}
					goto l157
				l158:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
				}
				depth--
				add(ruleContents, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 37 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				{
					position161, tokenIndex161, depth161 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l162
					}
					goto l161
				l162:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
					if !_rules[ruleSimpleMerge]() {
						goto l159
					}
				}
			l161:
				depth--
				add(ruleMerge, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 38 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				if buffer[position] != rune('m') {
					goto l163
				}
				position++
				if buffer[position] != rune('e') {
					goto l163
				}
				position++
				if buffer[position] != rune('r') {
					goto l163
				}
				position++
				if buffer[position] != rune('g') {
					goto l163
				}
				position++
				if buffer[position] != rune('e') {
					goto l163
				}
				position++
				{
					position165, tokenIndex165, depth165 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l165
					}
					if !_rules[ruleRequired]() {
						goto l165
					}
					goto l163
				l165:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
				}
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l166
					}
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l169
						}
						goto l168
					l169:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
						if !_rules[ruleOn]() {
							goto l166
						}
					}
				l168:
					goto l167
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
			l167:
				if !_rules[rulereq_ws]() {
					goto l163
				}
				if !_rules[ruleReference]() {
					goto l163
				}
				depth--
				add(ruleRefMerge, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 39 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				if buffer[position] != rune('m') {
					goto l170
				}
				position++
				if buffer[position] != rune('e') {
					goto l170
				}
				position++
				if buffer[position] != rune('r') {
					goto l170
				}
				position++
				if buffer[position] != rune('g') {
					goto l170
				}
				position++
				if buffer[position] != rune('e') {
					goto l170
				}
				position++
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l172
					}
					{
						position174, tokenIndex174, depth174 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l175
						}
						goto l174
					l175:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						if !_rules[ruleRequired]() {
							goto l176
						}
						goto l174
					l176:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						if !_rules[ruleOn]() {
							goto l172
						}
					}
				l174:
					goto l173
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
			l173:
				depth--
				add(ruleSimpleMerge, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 40 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
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
				if buffer[position] != rune('p') {
					goto l177
				}
				position++
				if buffer[position] != rune('l') {
					goto l177
				}
				position++
				if buffer[position] != rune('a') {
					goto l177
				}
				position++
				if buffer[position] != rune('c') {
					goto l177
				}
				position++
				if buffer[position] != rune('e') {
					goto l177
				}
				position++
				depth--
				add(ruleReplace, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 41 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
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
				if buffer[position] != rune('q') {
					goto l179
				}
				position++
				if buffer[position] != rune('u') {
					goto l179
				}
				position++
				if buffer[position] != rune('i') {
					goto l179
				}
				position++
				if buffer[position] != rune('r') {
					goto l179
				}
				position++
				if buffer[position] != rune('e') {
					goto l179
				}
				position++
				if buffer[position] != rune('d') {
					goto l179
				}
				position++
				depth--
				add(ruleRequired, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 42 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if buffer[position] != rune('o') {
					goto l181
				}
				position++
				if buffer[position] != rune('n') {
					goto l181
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l181
				}
				if !_rules[ruleName]() {
					goto l181
				}
				depth--
				add(ruleOn, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 43 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('a') {
					goto l183
				}
				position++
				if buffer[position] != rune('u') {
					goto l183
				}
				position++
				if buffer[position] != rune('t') {
					goto l183
				}
				position++
				if buffer[position] != rune('o') {
					goto l183
				}
				position++
				depth--
				add(ruleAuto, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 44 Mapping <- <('m' 'a' 'p' '[' Expression (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if buffer[position] != rune('m') {
					goto l185
				}
				position++
				if buffer[position] != rune('a') {
					goto l185
				}
				position++
				if buffer[position] != rune('p') {
					goto l185
				}
				position++
				if buffer[position] != rune('[') {
					goto l185
				}
				position++
				if !_rules[ruleExpression]() {
					goto l185
				}
				{
					position187, tokenIndex187, depth187 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l188
					}
					goto l187
				l188:
					position, tokenIndex, depth = position187, tokenIndex187, depth187
					if buffer[position] != rune('|') {
						goto l185
					}
					position++
					if !_rules[ruleExpression]() {
						goto l185
					}
				}
			l187:
				if buffer[position] != rune(']') {
					goto l185
				}
				position++
				depth--
				add(ruleMapping, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 45 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				if buffer[position] != rune('l') {
					goto l189
				}
				position++
				if buffer[position] != rune('a') {
					goto l189
				}
				position++
				if buffer[position] != rune('m') {
					goto l189
				}
				position++
				if buffer[position] != rune('b') {
					goto l189
				}
				position++
				if buffer[position] != rune('d') {
					goto l189
				}
				position++
				if buffer[position] != rune('a') {
					goto l189
				}
				position++
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l192
					}
					goto l191
				l192:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if !_rules[ruleLambdaExpr]() {
						goto l189
					}
				}
			l191:
				depth--
				add(ruleLambda, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 46 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l193
				}
				if !_rules[ruleExpression]() {
					goto l193
				}
				depth--
				add(ruleLambdaRef, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 47 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if !_rules[rulews]() {
					goto l195
				}
				if buffer[position] != rune('|') {
					goto l195
				}
				position++
				if !_rules[rulews]() {
					goto l195
				}
				if !_rules[ruleName]() {
					goto l195
				}
			l197:
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l198
					}
					goto l197
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
				if !_rules[rulews]() {
					goto l195
				}
				if buffer[position] != rune('|') {
					goto l195
				}
				position++
				if !_rules[rulews]() {
					goto l195
				}
				if buffer[position] != rune('-') {
					goto l195
				}
				position++
				if buffer[position] != rune('>') {
					goto l195
				}
				position++
				if !_rules[ruleExpression]() {
					goto l195
				}
				depth--
				add(ruleLambdaExpr, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 48 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if !_rules[rulews]() {
					goto l199
				}
				if buffer[position] != rune(',') {
					goto l199
				}
				position++
				if !_rules[rulews]() {
					goto l199
				}
				if !_rules[ruleName]() {
					goto l199
				}
				depth--
				add(ruleNextName, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 49 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l203
					}
					position++
					goto l204
				l203:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
			l204:
				if !_rules[ruleKey]() {
					goto l201
				}
			l205:
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l206
					}
					position++
					{
						position207, tokenIndex207, depth207 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l208
						}
						goto l207
					l208:
						position, tokenIndex, depth = position207, tokenIndex207, depth207
						if !_rules[ruleIndex]() {
							goto l206
						}
					}
				l207:
					goto l205
				l206:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
				}
				depth--
				add(ruleReference, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 50 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l212
					}
					goto l211
				l212:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if !_rules[ruleIndex]() {
						goto l209
					}
				}
			l211:
			l213:
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l214
					}
					position++
					{
						position215, tokenIndex215, depth215 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l216
						}
						goto l215
					l216:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if !_rules[ruleIndex]() {
							goto l214
						}
					}
				l215:
					goto l213
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
				depth--
				add(ruleFollowUpRef, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 51 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l220
					}
					position++
					goto l219
				l220:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l221
					}
					position++
					goto l219
				l221:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l222
					}
					position++
					goto l219
				l222:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					if buffer[position] != rune('_') {
						goto l217
					}
					position++
				}
			l219:
			l223:
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					{
						position225, tokenIndex225, depth225 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l226
						}
						position++
						goto l225
					l226:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l227
						}
						position++
						goto l225
					l227:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l228
						}
						position++
						goto l225
					l228:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
						if buffer[position] != rune('_') {
							goto l229
						}
						position++
						goto l225
					l229:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
						if buffer[position] != rune('-') {
							goto l224
						}
						position++
					}
				l225:
					goto l223
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l230
					}
					position++
					{
						position232, tokenIndex232, depth232 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l233
						}
						position++
						goto l232
					l233:
						position, tokenIndex, depth = position232, tokenIndex232, depth232
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l234
						}
						position++
						goto l232
					l234:
						position, tokenIndex, depth = position232, tokenIndex232, depth232
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l235
						}
						position++
						goto l232
					l235:
						position, tokenIndex, depth = position232, tokenIndex232, depth232
						if buffer[position] != rune('_') {
							goto l230
						}
						position++
					}
				l232:
				l236:
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l239
							}
							position++
							goto l238
						l239:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l240
							}
							position++
							goto l238
						l240:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l241
							}
							position++
							goto l238
						l241:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if buffer[position] != rune('_') {
								goto l242
							}
							position++
							goto l238
						l242:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if buffer[position] != rune('-') {
								goto l237
							}
							position++
						}
					l238:
						goto l236
					l237:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
					}
					goto l231
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
			l231:
				depth--
				add(ruleKey, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 52 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				if buffer[position] != rune('[') {
					goto l243
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l243
				}
				position++
			l245:
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l246
					}
					position++
					goto l245
				l246:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
				}
				if buffer[position] != rune(']') {
					goto l243
				}
				position++
				depth--
				add(ruleIndex, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 53 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position248 := position
				depth++
			l249:
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					{
						position251, tokenIndex251, depth251 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l252
						}
						position++
						goto l251
					l252:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
						if buffer[position] != rune('\t') {
							goto l253
						}
						position++
						goto l251
					l253:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
						if buffer[position] != rune('\n') {
							goto l254
						}
						position++
						goto l251
					l254:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
						if buffer[position] != rune('\r') {
							goto l250
						}
						position++
					}
				l251:
					goto l249
				l250:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
				}
				depth--
				add(rulews, position248)
			}
			return true
		},
		/* 54 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				{
					position259, tokenIndex259, depth259 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l260
					}
					position++
					goto l259
				l260:
					position, tokenIndex, depth = position259, tokenIndex259, depth259
					if buffer[position] != rune('\t') {
						goto l261
					}
					position++
					goto l259
				l261:
					position, tokenIndex, depth = position259, tokenIndex259, depth259
					if buffer[position] != rune('\n') {
						goto l262
					}
					position++
					goto l259
				l262:
					position, tokenIndex, depth = position259, tokenIndex259, depth259
					if buffer[position] != rune('\r') {
						goto l255
					}
					position++
				}
			l259:
			l257:
				{
					position258, tokenIndex258, depth258 := position, tokenIndex, depth
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
							goto l258
						}
						position++
					}
				l263:
					goto l257
				l258:
					position, tokenIndex, depth = position258, tokenIndex258, depth258
				}
				depth--
				add(rulereq_ws, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
	}
	p.rules = _rules
}
