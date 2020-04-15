package hcl

type unwrapExpression interface {
	UnwrapExpression() Expression
}

// UnwrapExpressionUntil is similar to UnwrapExpression except it gives the
// caller an opportunity to test each level of unwrapping to see each a
// particular expression is accepted.
//
// This could be used, for example, to unwrap until a particular other
// interface is satisfied, regardless of wrap wrapping level it is satisfied
// at.
//
// The given callback function must return false to continue wrapping, or
// true to accept and return the proposed expression given. If the callback
// function rejects even the final, physical expression then the result of
// this function is nil.
func UnwrapExpressionUntil(expr Expression, until func(Expression) bool) Expression {
	for {
		if until(expr) {
			return expr
		}
		unwrap, wrapped := expr.(unwrapExpression)
		if !wrapped {
			return nil
		}
		expr = unwrap.UnwrapExpression()
		if expr == nil {
			return nil
		}
	}
}
