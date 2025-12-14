package handler

import (
	"backend/internal/model"
	"backend/internal/response"
	"backend/internal/service"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc     service.UserService
	firebaseSvc service.FirebaseService
}

func NewUserHandler(userSvc service.UserService, firebaseSvc service.FirebaseService) *UserHandler {
	return &UserHandler{userSvc: userSvc, firebaseSvc: firebaseSvc}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userSvc.GetAllUsers()

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	var usersLookupFirebase []map[string]interface{}

	for _, user := range users {
		_user := map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"username":  user.Username,
			"updatedAt": user.UpdatedAt,
		}

		userFb, _ := h.firebaseSvc.GetUser(c.Request.Context(), user.Token)

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

	verify, err := h.firebaseSvc.VerifyToken(c, token)

	if err != nil {
		response.Forbidden(c, "invalid or expired token")
		return
	}

	log.Println("✅ Verified Firebase token for UID:", verify.UID)

	// Get user info from Firebase
	userFb, err := h.firebaseSvc.GetUser(c.Request.Context(), verify.UID)
	if err != nil {
		log.Printf("error %v", err)
		response.InternalError(c, "failed to get user info from Firebase")
		return
	}

	// Check if user exists
	existingUser, err := h.userSvc.FindUserByEmail(userFb.Email)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if existingUser != nil {
		// Update token
		existingUser.Token = verify.UID
		err = h.userSvc.UpdateUser(existingUser)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		c.SetCookie(
			"auth_token",
			token,
			120000,
			"/",
			"localhost",
			false,
			true,
		)
		c.JSON(200, existingUser)
		return
	}

	// Create new user
	newUser := &model.User{
		Email:    userFb.Email,
		Username: userFb.DisplayName,
		Password: "", // No password for Firebase users
		Token:    verify.UID,
	}

	err = h.userSvc.CreateUser(newUser)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.SetCookie(
		"auth_token",
		token,
		120000,
		"/",
		"localhost",
		false,
		true,
	)
	c.JSON(201, newUser)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	user, err := h.userSvc.FindUserByID(uint(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, user)
}
