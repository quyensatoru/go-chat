package middleware

import (
	contextkey "backend/internal/common/contextKey"
	"backend/internal/response"
	"backend/internal/service"
	"context"

	"github.com/gin-gonic/gin"
)

func FirebaseAuthMiddleware(firebaseSerivce *service.FirebaseService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("auth_token")

		if token == "" || err != nil {
			response.Forbidden(ctx, "Unauthorization")
			return
		}

		verify, err := firebaseSerivce.VerifyToken(ctx.Request.Context(), token)

		if err != nil {
			response.Forbidden(ctx, "invalid or expired token")
			return
		}

		newCtx := context.WithValue(ctx.Request.Context(), contextkey.UserFirebase, verify)

		ctx.Request = ctx.Request.WithContext(newCtx)

		ctx.Next()

	}
}
