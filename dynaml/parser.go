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

		switch token.Rule {
		case RuleDynaml:
			return tokens.Pop()
		case RuleAuto:
			tokens.Push(AutoExpr{path})
		case RuleMerge:
			tokens.Push(MergeExpr{path})
		case RuleReference:
			tokens.Push(ReferenceExpr{strings.Split(contents, ".")})
		case RuleInteger:
			val, err := strconv.Atoi(contents)
			if err != nil {
				panic(err)
			}

			tokens.Push(IntegerExpr{val})
		case RuleNil:
			tokens.Push(NilExpr{})
		case RuleBoolean:
			tokens.Push(BooleanExpr{contents == "true"})
		case RuleString:
			val := strings.Replace(contents[1:len(contents)-1], `\"`, `"`, -1)
			tokens.Push(StringExpr{val})
		case RuleOr:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(OrExpr{A: lhs, B: rhs})
		case RuleConcatenation:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(ConcatenationExpr{A: lhs, B: rhs})
		case RuleAddition:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(AdditionExpr{A: lhs, B: rhs})
		case RuleSubtraction:
			rhs := tokens.Pop()
			lhs := tokens.Pop()

			tokens.Push(SubtractionExpr{A: lhs, B: rhs})
		case RuleCall:
			tokens.Push(CallExpr{
				Name:      tokens.functionName,
				Arguments: tokens.PopSeq(),
			})
		case RuleName:
			tokens.functionName = contents
		case RuleList:
			seq := tokens.PopSeq()
			tokens.Push(ListExpr{seq})
		case RuleComma, RuleContents, RuleArguments:
			expr := tokens.Pop()
			tokens.PushToSeq(expr)
		case RuleGrouped:
		case RuleLevel0, RuleLevel1, RuleLevel2:
		case RuleExpression:
		case Rulews:
		default:
			panic("unhandled:" + Rul3s[token.Rule])
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
