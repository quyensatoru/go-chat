package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *model.User) error
	FindUserByEmail(email string) (*model.User, error)
	FindUserByID(id uint) (*model.User, error)
	GetAllUsers() ([]model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id uint) error
	ValidatePassword(hashedPassword, password string) error
	HashPassword(password string) (string, error)
	FindUserByUID(id string) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(user *model.User) error {
	// Check if user already exists
	existingUser, err := s.repo.FindByEmail(user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// Hash password before saving
	hashedPassword, err := s.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return s.repo.Create(user)
}

func (s *userService) FindUserByEmail(email string) (*model.User, error) {
	return s.repo.FindByEmail(email)
}

func (s *userService) FindUserByID(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *userService) GetAllUsers() ([]model.User, error) {
	return s.repo.FindAll()
}

func (s *userService) UpdateUser(user *model.User) error {
	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

func (s *userService) ValidatePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *userService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *userService) FindUserByUID(id string) (*model.User, error) {
	return s.repo.FindByUID(id)
}
