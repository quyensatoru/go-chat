package dto

type CreateUserDto struct {
	Email    string `json:"email" binding:"optional,email"`
	Password string `json:"password" binding:"optional"`
}
