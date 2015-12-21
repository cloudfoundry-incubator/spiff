package dynaml

import (
	"container/list"
	"strconv"
	"strings"
)

func Parse(source string, path []string) (Expression, error) {
	grammar := &DynamlGrammar{Buffer: source}
	grammar.Init()

	err := grammar.Parse()
	if err != nil {
		return nil, err
	}

	return buildExpression(grammar, path), nil
}

func buildExpression(grammar *DynamlGrammar, path []string) Expression {
	tokens := &tokenStack{}

	for token := range grammar.Tokens() {
		contents := grammar.Buffer[token.begin:token.end]

		switch token.pegRule {
		case ruleDynaml:
			return tokens.Pop()
		case ruleAuto:
			tokens.Push(AutoExpr{path})
		case ruleMerge:
			tokens.Push(MergeExpr{path})
		case ruleReference:
			tokens.Push(ReferenceExpr{strings.Split(contents, ".")})
		case ruleInteger:
			val, err := strconv.ParseInt(contents, 10, 64)
			if err != nil {
				panic(err)
			}

			tokens.Push(IntegerExpr{val})
		case ruleNil:
			tokens.Push(NilExpr{})
		case ruleBoolean:
			tokens.Push(BooleanExpr{contents == "true"})
		case ruleString:
			val := strings.Replace(contents[1:len(contents)-1], `\"`, `"`, -1)
			tokens.Push(StringExpr{val})
		case ruleOr:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(OrExpr{A: lhs, B: rhs})
		case ruleConcatenation:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(ConcatenationExpr{A: lhs, B: rhs})
		case ruleAddition:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(AdditionExpr{A: lhs, B: rhs})
		case ruleSubtraction:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(SubtractionExpr{A: lhs, B: rhs})
		case ruleCall:
			tokens.Push(CallExpr{
				Name:      tokens.functionName,
				Arguments: tokens.PopSeq(),
			})
		case ruleName:
			tokens.functionName = contents
		case ruleList:
			seq := tokens.PopSeq()
			tokens.Push(ListExpr{seq})
		case ruleComma, ruleContents, ruleArguments:
			expr := tokens.Pop()
			tokens.PushToSeq(expr)
		case ruleGrouped:
		case ruleLevel0, ruleLevel1, ruleLevel2:
		case ruleExpression:
		case rulews:
		case rulereq_ws:
		default:
			panic("unhandled:" + rul3s[token.pegRule])
		}
	}

	panic("unreachable")
}

type tokenStack struct {
	list.List

	seq []Expression

	functionName string
}

func (s *tokenStack) Pop() Expression {
	front := s.Front()
	if front == nil {
		return nil
	}

	s.Remove(front)

	return front.Value.(Expression)
}

func (s *tokenStack) Push(expr Expression) {
	s.PushFront(expr)
}

func (s *tokenStack) PushToSeq(expr Expression) {
	if s.seq == nil {
		s.seq = []Expression{}
	}

	s.seq = append(s.seq, expr)
}

func (s *tokenStack) PopSeq() []Expression {
	seq := s.seq
	s.seq = nil
	return seq
}
