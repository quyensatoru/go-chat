package handler

import (
	"backend/internal/response"
	"backend/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
	fb  *service.FirebaseService
}

func NewUserHandler(svc *service.UserService, fb *service.FirebaseService) *UserHandler {
	return &UserHandler{svc: svc, fb: fb}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.svc.GetAll(c.Request.Context())

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	var usersLookupFirebase []map[string]interface{}

	for _, user := range users {
		_user := map[string]interface{}{
			"_id":       user.ID,
			"email":     user.Email,
			"username":  user.Username,
			"updatedAt": user.UpdatedAt,
		}

		userFb, _ := h.fb.GetUser(c.Request.Context(), user.Token)

		if userFb != nil {
			_user["photoUrl"] = userFb.PhotoURL
		}

		usersLookupFirebase = append(usersLookupFirebase, _user)

	}

	response.Success(c, usersLookupFirebase)
}

func (h *UserHandler) CreateNewAccount(c *gin.Context) {
	authStr := c.Request.Header.Get("Authorization")

	if authStr == "" {
		response.Forbidden(c, "missing Authorization header")
		return
	}

	token := strings.TrimPrefix(authStr, "Bearer ")

	verify, err := h.fb.VerifyToken(c, token)

	if err != nil {
		response.Forbidden(c, "invalid or expired token")
		return
	}

	createdUser, err := h.svc.CreateByFirbase(c.Request.Context(), verify)

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.SetCookie(
		"auth_token",
		token,
		3600000,
		"/",
		"localhost",
		false,
		true,
	)
	c.JSON(201, createdUser)
}
