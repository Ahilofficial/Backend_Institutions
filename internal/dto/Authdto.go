package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"regexp"
	"strings"

	"github.com/jinzhu/copier"
)

type ForgotPasswordDTO struct {
	Email string `json:"email"`
}

type ResetPassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type ResendResetPassword struct {
	ResetToken string `json:"reset_token"`
}

type SignUpDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type SignInDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogoutDTO struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"refresh_token"`
}

type ResendMailSignUp struct {
	Email string `json:"email"`
}

type AuthResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       uint   `json:"user_id"`
	SessionID    string `json:"session_id"`
	Role         string `json:"role,omitempty"`
}

type AssignRoleDTO struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
}

type UserResponseDTO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	IsActive bool   `json:"isactive"`
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9._%+-]{0,62}[a-zA-Z0-9])?@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)
	phoneRegex = regexp.MustCompile(`^[0-9]{10}$`)
)

func (dto *SignUpDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
	dto.Phone = strings.TrimSpace(dto.Phone)
}

func (dto *SignUpDTO) Validate() error {
	if dto.Name == "" || dto.Email == "" || dto.Phone == "" || dto.Password == "" {
		return errors.New("all fields are required")
	}

	if !emailRegex.MatchString(dto.Email) {
		return errors.New("invalid email format")
	}

	if !phoneRegex.MatchString(dto.Phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

func (dto *SignInDTO) Sanitize() {
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
}

func (dto *SignInDTO) Validate() error {
	if dto.Email == "" || dto.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
}

func (dto *AssignRoleDTO) Sanitize() {
	dto.Role = strings.TrimSpace(strings.ToLower(dto.Role))
}

func (dto *AssignRoleDTO) Validate() error {
	if dto.UserID == 0 {
		return errors.New("user_id is required")
	}

	if dto.Role == "" {
		return errors.New("role is required")
	}

	return nil
}

func ToUserResponseDTO(user *model.User) UserResponseDTO {
	var dto UserResponseDTO
	copier.Copy(&dto, user)
	return dto
}

func ToUserResponseListDTO(users []model.User) []UserResponseDTO {
	list := make([]UserResponseDTO, len(users))
	for i, u := range users {
		list[i] = ToUserResponseDTO(&u)
	}
	return list
}
