// Layer: Security (安全基础设施层)
// Responsibility: 认证信息在Context中的存取、上下文传播
// Dependency: 不依赖业务层，被 Middleware 和 Usecase 使用
package security

import "context"

type authContextKey struct{}

func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, authContextKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	v := ctx.Value(authContextKey{})
	claims, ok := v.(Claims)
	return claims, ok
}
