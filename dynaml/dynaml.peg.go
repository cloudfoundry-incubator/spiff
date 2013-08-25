package dynaml

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
)

const END_SYMBOL byte = 0

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	RuleDynaml
	RuleExpression
	RuleLevel2
	RuleOr
	RuleLevel1
	RuleConcatenation
	RuleAddition
	RuleSubtraction
	RuleLevel0
	RuleGrouped
	RuleCall
	RuleArguments
	RuleName
	RuleComma
	RuleInteger
	RuleString
	RuleBoolean
	RuleList
	RuleContents
	RuleMerge
	RuleAuto
	RuleReference
	Rulews

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"Dynaml",
	"Expression",
	"Level2",
	"Or",
	"Level1",
	"Concatenation",
	"Addition",
	"Subtraction",
	"Level0",
	"Grouped",
	"Call",
	"Arguments",
	"Name",
	"Comma",
	"Integer",
	"String",
	"Boolean",
	"List",
	"Contents",
	"Merge",
	"Auto",
	"Reference",
	"ws",

	"Pre_",
	"_In_",
	"_Suf",
}

type TokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule Rule, begin, end, next, depth int)
	Expand(index int) TokenTree
	Tokens() <-chan token32
	Error() []token32
	trim(length int)
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	Rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
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

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) PreOrder() (<-chan State16, [][]token16) {
	s, ordered := make(chan State16, 6), t.Order()
	go func() {
		var states [8]State16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
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
							write(token16{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
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

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
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
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token16{Rule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	Rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
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
		if token.Rule == RuleUnknown {
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
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) PreOrder() (<-chan State32, [][]token32) {
	s, ordered := make(chan State32, 6), t.Order()
	go func() {
		var states [8]State32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int32(depth), leaf
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
							write(token32{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
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
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
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
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
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
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token32{Rule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
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
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.GetToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) TokenTree {
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
	rules  [24]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	TokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
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
	tokens, error := e.p.TokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			Rul3s[token.Rule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *DynamlGrammar) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *DynamlGrammar) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *DynamlGrammar) Init() {
	if p.Buffer[len(p.Buffer)-1] != END_SYMBOL {
		p.Buffer = p.Buffer + string(END_SYMBOL)
	}

	var tree TokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.Buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.TokenTree = tree
		if matches {
			p.TokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule Rule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != END_SYMBOL {
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

	rules = [...]func() bool{
		nil,
		/* 0 Dynaml <- <(ws Expression ws !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[Rulews]() {
					goto l0
				}
				if !rules[RuleExpression]() {
					goto l0
				}
				if !rules[Rulews]() {
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
				add(RuleDynaml, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Expression <- <Level2> */
		func() bool {
			position3, tokenIndex3, depth3 := position, tokenIndex, depth
			{
				position4 := position
				depth++
				if !rules[RuleLevel2]() {
					goto l3
				}
				depth--
				add(RuleExpression, position4)
			}
			return true
		l3:
			position, tokenIndex, depth = position3, tokenIndex3, depth3
			return false
		},
		/* 2 Level2 <- <(Or / Level1)> */
		func() bool {
			position5, tokenIndex5, depth5 := position, tokenIndex, depth
			{
				position6 := position
				depth++
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !rules[RuleOr]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
					if !rules[RuleLevel1]() {
						goto l5
					}
				}
			l7:
				depth--
				add(RuleLevel2, position6)
			}
			return true
		l5:
			position, tokenIndex, depth = position5, tokenIndex5, depth5
			return false
		},
		/* 3 Or <- <(Level1 ws ('|' '|') ws Expression)> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				if !rules[RuleLevel1]() {
					goto l9
				}
				if !rules[Rulews]() {
					goto l9
				}
				if buffer[position] != '|' {
					goto l9
				}
				position++
				if buffer[position] != '|' {
					goto l9
				}
				position++
				if !rules[Rulews]() {
					goto l9
				}
				if !rules[RuleExpression]() {
					goto l9
				}
				depth--
				add(RuleOr, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 4 Level1 <- <(Concatenation / Addition / Subtraction / Level0)> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !rules[RuleConcatenation]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !rules[RuleAddition]() {
						goto l15
					}
					goto l13
				l15:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !rules[RuleSubtraction]() {
						goto l16
					}
					goto l13
				l16:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
					if !rules[RuleLevel0]() {
						goto l11
					}
				}
			l13:
				depth--
				add(RuleLevel1, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 5 Concatenation <- <(Level0 (' ' / '\t' / '\n' / '\r')+ Level1)> */
		func() bool {
			position17, tokenIndex17, depth17 := position, tokenIndex, depth
			{
				position18 := position
				depth++
				if !rules[RuleLevel0]() {
					goto l17
				}
				{
					position21, tokenIndex21, depth21 := position, tokenIndex, depth
					if buffer[position] != ' ' {
						goto l22
					}
					position++
					goto l21
				l22:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
					if buffer[position] != '\t' {
						goto l23
					}
					position++
					goto l21
				l23:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
					if buffer[position] != '\n' {
						goto l24
					}
					position++
					goto l21
				l24:
					position, tokenIndex, depth = position21, tokenIndex21, depth21
					if buffer[position] != '\r' {
						goto l17
					}
					position++
				}
			l21:
			l19:
				{
					position20, tokenIndex20, depth20 := position, tokenIndex, depth
					{
						position25, tokenIndex25, depth25 := position, tokenIndex, depth
						if buffer[position] != ' ' {
							goto l26
						}
						position++
						goto l25
					l26:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if buffer[position] != '\t' {
							goto l27
						}
						position++
						goto l25
					l27:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if buffer[position] != '\n' {
							goto l28
						}
						position++
						goto l25
					l28:
						position, tokenIndex, depth = position25, tokenIndex25, depth25
						if buffer[position] != '\r' {
							goto l20
						}
						position++
					}
				l25:
					goto l19
				l20:
					position, tokenIndex, depth = position20, tokenIndex20, depth20
				}
				if !rules[RuleLevel1]() {
					goto l17
				}
				depth--
				add(RuleConcatenation, position18)
			}
			return true
		l17:
			position, tokenIndex, depth = position17, tokenIndex17, depth17
			return false
		},
		/* 6 Addition <- <(Level0 ws '+' ws Level1)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if !rules[RuleLevel0]() {
					goto l29
				}
				if !rules[Rulews]() {
					goto l29
				}
				if buffer[position] != '+' {
					goto l29
				}
				position++
				if !rules[Rulews]() {
					goto l29
				}
				if !rules[RuleLevel1]() {
					goto l29
				}
				depth--
				add(RuleAddition, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 7 Subtraction <- <(Level0 ws '-' ws Level1)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if !rules[RuleLevel0]() {
					goto l31
				}
				if !rules[Rulews]() {
					goto l31
				}
				if buffer[position] != '-' {
					goto l31
				}
				position++
				if !rules[Rulews]() {
					goto l31
				}
				if !rules[RuleLevel1]() {
					goto l31
				}
				depth--
				add(RuleSubtraction, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 8 Level0 <- <(Grouped / Call / Boolean / String / Integer / List / Merge / Auto / Reference)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !rules[RuleGrouped]() {
						goto l36
					}
					goto l35
				l36:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleCall]() {
						goto l37
					}
					goto l35
				l37:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleBoolean]() {
						goto l38
					}
					goto l35
				l38:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleString]() {
						goto l39
					}
					goto l35
				l39:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleInteger]() {
						goto l40
					}
					goto l35
				l40:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleList]() {
						goto l41
					}
					goto l35
				l41:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleMerge]() {
						goto l42
					}
					goto l35
				l42:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleAuto]() {
						goto l43
					}
					goto l35
				l43:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !rules[RuleReference]() {
						goto l33
					}
				}
			l35:
				depth--
				add(RuleLevel0, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 9 Grouped <- <('(' Expression ')')> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if buffer[position] != '(' {
					goto l44
				}
				position++
				if !rules[RuleExpression]() {
					goto l44
				}
				if buffer[position] != ')' {
					goto l44
				}
				position++
				depth--
				add(RuleGrouped, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 10 Call <- <(Name '(' Arguments ')')> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !rules[RuleName]() {
					goto l46
				}
				if buffer[position] != '(' {
					goto l46
				}
				position++
				if !rules[RuleArguments]() {
					goto l46
				}
				if buffer[position] != ')' {
					goto l46
				}
				position++
				depth--
				add(RuleCall, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 11 Arguments <- <(Expression (Comma ws Expression)*)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !rules[RuleExpression]() {
					goto l48
				}
			l50:
				{
					position51, tokenIndex51, depth51 := position, tokenIndex, depth
					if !rules[RuleComma]() {
						goto l51
					}
					if !rules[Rulews]() {
						goto l51
					}
					if !rules[RuleExpression]() {
						goto l51
					}
					goto l50
				l51:
					position, tokenIndex, depth = position51, tokenIndex51, depth51
				}
				depth--
				add(RuleArguments, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 12 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if c := buffer[position]; c < 'a' || c > 'z' {
						goto l57
					}
					position++
					goto l56
				l57:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if c := buffer[position]; c < 'A' || c > 'Z' {
						goto l58
					}
					position++
					goto l56
				l58:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if c := buffer[position]; c < '0' || c > '9' {
						goto l59
					}
					position++
					goto l56
				l59:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if buffer[position] != '_' {
						goto l52
					}
					position++
				}
			l56:
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						if c := buffer[position]; c < 'a' || c > 'z' {
							goto l61
						}
						position++
						goto l60
					l61:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if c := buffer[position]; c < 'A' || c > 'Z' {
							goto l62
						}
						position++
						goto l60
					l62:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if c := buffer[position]; c < '0' || c > '9' {
							goto l63
						}
						position++
						goto l60
					l63:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if buffer[position] != '_' {
							goto l55
						}
						position++
					}
				l60:
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				depth--
				add(RuleName, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 13 Comma <- <','> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if buffer[position] != ',' {
					goto l64
				}
				position++
				depth--
				add(RuleComma, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 14 Integer <- <('-'? ([0-9] / '_')+)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				{
					position68, tokenIndex68, depth68 := position, tokenIndex, depth
					if buffer[position] != '-' {
						goto l68
					}
					position++
					goto l69
				l68:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
				}
			l69:
				{
					position72, tokenIndex72, depth72 := position, tokenIndex, depth
					if c := buffer[position]; c < '0' || c > '9' {
						goto l73
					}
					position++
					goto l72
				l73:
					position, tokenIndex, depth = position72, tokenIndex72, depth72
					if buffer[position] != '_' {
						goto l66
					}
					position++
				}
			l72:
			l70:
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					{
						position74, tokenIndex74, depth74 := position, tokenIndex, depth
						if c := buffer[position]; c < '0' || c > '9' {
							goto l75
						}
						position++
						goto l74
					l75:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						if buffer[position] != '_' {
							goto l71
						}
						position++
					}
				l74:
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				depth--
				add(RuleInteger, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 15 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != '"' {
					goto l76
				}
				position++
			l78:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					{
						position80, tokenIndex80, depth80 := position, tokenIndex, depth
						if buffer[position] != '\\' {
							goto l81
						}
						position++
						if buffer[position] != '"' {
							goto l81
						}
						position++
						goto l80
					l81:
						position, tokenIndex, depth = position80, tokenIndex80, depth80
						{
							position82, tokenIndex82, depth82 := position, tokenIndex, depth
							if buffer[position] != '"' {
								goto l82
							}
							position++
							goto l79
						l82:
							position, tokenIndex, depth = position82, tokenIndex82, depth82
						}
						if !matchDot() {
							goto l79
						}
					}
				l80:
					goto l78
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
				if buffer[position] != '"' {
					goto l76
				}
				position++
				depth--
				add(RuleString, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 16 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if buffer[position] != 't' {
						goto l86
					}
					position++
					if buffer[position] != 'r' {
						goto l86
					}
					position++
					if buffer[position] != 'u' {
						goto l86
					}
					position++
					if buffer[position] != 'e' {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != 'f' {
						goto l83
					}
					position++
					if buffer[position] != 'a' {
						goto l83
					}
					position++
					if buffer[position] != 'l' {
						goto l83
					}
					position++
					if buffer[position] != 's' {
						goto l83
					}
					position++
					if buffer[position] != 'e' {
						goto l83
					}
					position++
				}
			l85:
				depth--
				add(RuleBoolean, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 17 List <- <('[' Contents ']')> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != '[' {
					goto l87
				}
				position++
				if !rules[RuleContents]() {
					goto l87
				}
				if buffer[position] != ']' {
					goto l87
				}
				position++
				depth--
				add(RuleList, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 18 Contents <- <(Expression (Comma ws Expression)*)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if !rules[RuleExpression]() {
					goto l89
				}
			l91:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !rules[RuleComma]() {
						goto l92
					}
					if !rules[Rulews]() {
						goto l92
					}
					if !rules[RuleExpression]() {
						goto l92
					}
					goto l91
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
				depth--
				add(RuleContents, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 19 Merge <- <('m' 'e' 'r' 'g' 'e')> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if buffer[position] != 'm' {
					goto l93
				}
				position++
				if buffer[position] != 'e' {
					goto l93
				}
				position++
				if buffer[position] != 'r' {
					goto l93
				}
				position++
				if buffer[position] != 'g' {
					goto l93
				}
				position++
				if buffer[position] != 'e' {
					goto l93
				}
				position++
				depth--
				add(RuleMerge, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 20 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				if buffer[position] != 'a' {
					goto l95
				}
				position++
				if buffer[position] != 'u' {
					goto l95
				}
				position++
				if buffer[position] != 't' {
					goto l95
				}
				position++
				if buffer[position] != 'o' {
					goto l95
				}
				position++
				depth--
				add(RuleAuto, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 21 Reference <- <(([a-z] / [A-Z] / [0-9] / '_')+ (('.' ([a-z] / [A-Z] / [0-9] / '_')+) / ('.' '[' [0-9]+ ']'))*)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					if c := buffer[position]; c < 'a' || c > 'z' {
						goto l102
					}
					position++
					goto l101
				l102:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if c := buffer[position]; c < 'A' || c > 'Z' {
						goto l103
					}
					position++
					goto l101
				l103:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if c := buffer[position]; c < '0' || c > '9' {
						goto l104
					}
					position++
					goto l101
				l104:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
					if buffer[position] != '_' {
						goto l97
					}
					position++
				}
			l101:
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					{
						position105, tokenIndex105, depth105 := position, tokenIndex, depth
						if c := buffer[position]; c < 'a' || c > 'z' {
							goto l106
						}
						position++
						goto l105
					l106:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
						if c := buffer[position]; c < 'A' || c > 'Z' {
							goto l107
						}
						position++
						goto l105
					l107:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
						if c := buffer[position]; c < '0' || c > '9' {
							goto l108
						}
						position++
						goto l105
					l108:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
						if buffer[position] != '_' {
							goto l100
						}
						position++
					}
				l105:
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
			l109:
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					{
						position111, tokenIndex111, depth111 := position, tokenIndex, depth
						if buffer[position] != '.' {
							goto l112
						}
						position++
						{
							position115, tokenIndex115, depth115 := position, tokenIndex, depth
							if c := buffer[position]; c < 'a' || c > 'z' {
								goto l116
							}
							position++
							goto l115
						l116:
							position, tokenIndex, depth = position115, tokenIndex115, depth115
							if c := buffer[position]; c < 'A' || c > 'Z' {
								goto l117
							}
							position++
							goto l115
						l117:
							position, tokenIndex, depth = position115, tokenIndex115, depth115
							if c := buffer[position]; c < '0' || c > '9' {
								goto l118
							}
							position++
							goto l115
						l118:
							position, tokenIndex, depth = position115, tokenIndex115, depth115
							if buffer[position] != '_' {
								goto l112
							}
							position++
						}
					l115:
					l113:
						{
							position114, tokenIndex114, depth114 := position, tokenIndex, depth
							{
								position119, tokenIndex119, depth119 := position, tokenIndex, depth
								if c := buffer[position]; c < 'a' || c > 'z' {
									goto l120
								}
								position++
								goto l119
							l120:
								position, tokenIndex, depth = position119, tokenIndex119, depth119
								if c := buffer[position]; c < 'A' || c > 'Z' {
									goto l121
								}
								position++
								goto l119
							l121:
								position, tokenIndex, depth = position119, tokenIndex119, depth119
								if c := buffer[position]; c < '0' || c > '9' {
									goto l122
								}
								position++
								goto l119
							l122:
								position, tokenIndex, depth = position119, tokenIndex119, depth119
								if buffer[position] != '_' {
									goto l114
								}
								position++
							}
						l119:
							goto l113
						l114:
							position, tokenIndex, depth = position114, tokenIndex114, depth114
						}
						goto l111
					l112:
						position, tokenIndex, depth = position111, tokenIndex111, depth111
						if buffer[position] != '.' {
							goto l110
						}
						position++
						if buffer[position] != '[' {
							goto l110
						}
						position++
						if c := buffer[position]; c < '0' || c > '9' {
							goto l110
						}
						position++
					l123:
						{
							position124, tokenIndex124, depth124 := position, tokenIndex, depth
							if c := buffer[position]; c < '0' || c > '9' {
								goto l124
							}
							position++
							goto l123
						l124:
							position, tokenIndex, depth = position124, tokenIndex124, depth124
						}
						if buffer[position] != ']' {
							goto l110
						}
						position++
					}
				l111:
					goto l109
				l110:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
				}
				depth--
				add(RuleReference, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 22 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position126 := position
				depth++
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					{
						position129, tokenIndex129, depth129 := position, tokenIndex, depth
						if buffer[position] != ' ' {
							goto l130
						}
						position++
						goto l129
					l130:
						position, tokenIndex, depth = position129, tokenIndex129, depth129
						if buffer[position] != '\t' {
							goto l131
						}
						position++
						goto l129
					l131:
						position, tokenIndex, depth = position129, tokenIndex129, depth129
						if buffer[position] != '\n' {
							goto l132
						}
						position++
						goto l129
					l132:
						position, tokenIndex, depth = position129, tokenIndex129, depth129
						if buffer[position] != '\r' {
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
				add(Rulews, position126)
			}
			return true
		},
	}
	p.rules = rules
}
