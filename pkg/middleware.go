package pkg

import (
	"github.com/kataras/iris/v12"
)

// Middleware holds dependencies for middleware functions
type Middleware struct{}

// NewMiddleware creates a new Middleware instance
func NewMiddleware() *Middleware {
	return &Middleware{}
}

// TenantMiddleware extracts Tenant from headers and adds it to the context
func (m *Middleware) TenantMiddleware(ctx iris.Context) {
	tenant := ctx.GetHeader("Tenant")
	if tenant == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Missing Tenant header"})
		return
	}
	ctx.Values().Set("tenant", tenant)
	ctx.Next()
}
