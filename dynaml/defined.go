package dynaml

func (e CallExpr) defined(binding Binding) (interface{}, EvaluationInfo, bool) {
	pushed := make([]Expression, len(e.Arguments))
	ok := true
	resolved := true

	copy(pushed, e.Arguments)
	for i, _ := range pushed {
		_, _, ok = ResolveExpressionOrPushEvaluation(&pushed[i], &resolved, nil, binding)
		if resolved && !ok {
			return false, DefaultInfo(), true
		}
	}
	if !resolved {
		return e, DefaultInfo(), true
	}
	return true, DefaultInfo(), ok
}
