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
	// Role     string `json:"role,omitempty"`
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
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	IsActive    bool   `json:"isactive"`
	Role        string `json:"role,omitempty"`
	StudentID   uint   `json:"student_id,omitempty"`
	FacultyID   uint   `json:"faculty_id,omitempty"`
}

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9._%+-]{0,62}[a-zA-Z0-9])?@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)
	phoneRegex = regexp.MustCompile(`^[0-9]{10}$`)
)

func (dto *SignUpDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
	dto.Phone = strings.TrimSpace(dto.Phone)
	// dto.Role = strings.TrimSpace(strings.ToLower(dto.Role))
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

func (dto *ForgotPasswordDTO) Sanitize() {
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
}

func (dto *ForgotPasswordDTO) Validate() error {
	if dto.Email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(dto.Email) {
		return errors.New("invalid email format")
	}
	return nil
}

func (dto *ResetPassword) Sanitize() {
	dto.CurrentPassword = strings.TrimSpace(dto.CurrentPassword)
	dto.NewPassword = strings.TrimSpace(dto.NewPassword)
}

func (dto *ResetPassword) Validate() error {
	if dto.CurrentPassword == "" {
		return errors.New("current password is required")
	}
	if dto.NewPassword == "" {
		return errors.New("new password is required")
	}
	if len(dto.NewPassword) < 6 {
		return errors.New("new password must be at least 6 characters long")
	}
	return nil
}

func (dto *ResendResetPassword) Sanitize() {
	dto.ResetToken = strings.TrimSpace(dto.ResetToken)
}

func (dto *ResendResetPassword) Validate() error {
	if dto.ResetToken == "" {
		return errors.New("reset token is required")
	}
	return nil
}

func (dto *LogoutDTO) Sanitize() {
	dto.Token = strings.TrimSpace(dto.Token)
}

func (dto *LogoutDTO) Validate() error {
	if dto.UserID == 0 {
		return errors.New("user_id is required")
	}
	return nil
}

func (dto *ResendMailSignUp) Sanitize() {
	dto.Email = strings.TrimSpace(strings.ToLower(dto.Email))
}

func (dto *ResendMailSignUp) Validate() error {
	if dto.Email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(dto.Email) {
		return errors.New("invalid email format")
	}
	return nil
}

func ToUserResponseDTO(user *model.User) UserResponseDTO {
	var dto UserResponseDTO
	copier.Copy(&dto, user)

	dto.StudentID = user.StudentID
	dto.FacultyID = user.FacultyID

	if len(user.Roles) > 0 {
		dto.Role = user.Roles[0].Name
	}

	return dto
}

func ToUserResponseListDTO(users []model.User) []UserResponseDTO {
	list := make([]UserResponseDTO, len(users))
	for i, u := range users {
		list[i] = ToUserResponseDTO(&u)
	}
	return list
}
