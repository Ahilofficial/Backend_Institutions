package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/grpc"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"backend_institutions/internal/utils"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func getBaseURL() string {
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL != "" {
		return baseURL
	}
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8090"
	}
	return "http://localhost:" + port
}

type UserService struct {
	userrepo       *repository.UserRepository
	sessionService *SessionService
}

func NewUserService(userrepo *repository.UserRepository, sessionService *SessionService) *UserService {
	return &UserService{
		userrepo:       userrepo,
		sessionService: sessionService,
	}
}

func (s *UserService) SignUp(dto *dto.SignUpDTO) (model.User, error) {
	if dto.Role != "" {
		return s.SignUpWithRole(dto, dto.Role)
	}

	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		return model.User{}, err
	}

	token := utils.SignUpToken()

	user := model.User{
		Name:              dto.Name,
		Email:             dto.Email,
		Phone:             dto.Phone,
		Password:          hashedPassword,
		IsActive:          false,
		IsVerified:        false,
		VerificationToken: token,
		TokenExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	err = s.userrepo.CreateUser(&user)
	if err != nil {
		return model.User{}, err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", getBaseURL(), token)
	subject := "Verify your email - Backend Institutions"
	body := fmt.Sprintf(`<h1>Hello %s,</h1>
<p>Thank you for signing up. Please verify your email by clicking the link below:</p>
<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; border-radius: 5px;">Verify Email Address</a></p>
<p>Or copy and paste this link in your browser:<br/><a href="%s">%s</a></p>
<p>This link will expire in 24 hours.</p>`, user.Name, verifyURL, verifyURL, verifyURL)

	go func(email, subject, body string) {
		if sendErr := grpc.SendEmail(email, subject, body, "signup"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(user.Email, subject, body)

	return user, nil
}

func (s *UserService) SignUpWithRole(dto *dto.SignUpDTO, roleName string) (model.User, error) {
	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		return model.User{}, err
	}

	token := utils.SignUpToken()

	user := model.User{
		Name:              dto.Name,
		Email:             dto.Email,
		Phone:             dto.Phone,
		Password:          hashedPassword,
		IsActive:          false,
		IsVerified:        false,
		VerificationToken: token,
		TokenExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	err = s.userrepo.CreateUser(&user)
	if err != nil {
		return model.User{}, err
	}

	if strings.TrimSpace(roleName) != "" {
		_ = s.userrepo.AssignRoleByName(user.ID, strings.TrimSpace(roleName))
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", getBaseURL(), token)
	subject := "Verify your email - Backend Institutions"
	body := fmt.Sprintf(`<h1>Hello %s,</h1>
<p>Thank you for signing up. Please verify your email by clicking the link below:</p>
<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; border-radius: 5px;">Verify Email Address</a></p>
<p>Or copy and paste this link in your browser:<br/><a href="%s">%s</a></p>
<p>This link will expire in 24 hours.</p>`, user.Name, verifyURL, verifyURL, verifyURL)

	go func(email, subject, body string) {
		if sendErr := grpc.SendEmail(email, subject, body, "signup"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(user.Email, subject, body)

	return user, nil
}

func (s *UserService) SendVerificationEmail(userID uint) error {
	user, err := s.userrepo.FindByID(userID)
	if err != nil {
		return err
	}

	token := utils.SignUpToken()
	if token == "" {
		return errors.New("failed to generate verification token")
	}

	user.IsActive = false
	user.IsVerified = false
	user.VerificationToken = token
	user.TokenExpiresAt = time.Now().Add(24 * time.Hour)

	if err := s.userrepo.UpdateUser(&user); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", getBaseURL(), token)
	subject := "Verify your email - Backend Institutions"
	body := fmt.Sprintf(`<h1>Hello %s,</h1>
<p>Please verify your email address by clicking the link below:</p>
<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; border-radius: 5px;">Verify Email Address</a></p>
<p>Or copy and paste this link in your browser:<br/><a href="%s">%s</a></p>
<p>This link will expire in 24 hours.</p>`, user.Name, verifyURL, verifyURL, verifyURL)

	go func(email, subject, body string) {
		if err := grpc.SendEmail(email, subject, body, "signup"); err != nil {
			log.Printf("Failed to send verification email via gRPC: %v", err)
		}
	}(user.Email, subject, body)

	return nil
}

func (s *UserService) CheckUserRole(userID uint, targetRole string) (bool, error) {
	return s.userrepo.CheckUserRole(userID, targetRole)
}
func (s *UserService) UpdateStudentID(userID uint, studentID uint) error {
	return s.userrepo.UpdateStudentID(userID, studentID)
}
func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	return s.userrepo.GetUserByID(userID)
}
func (s *UserService) UpdatePrincipalID(userID uint, principalID uint) error {
	return s.userrepo.UpdatePrincipalID(userID, principalID)
}
func (s *UserService) UpdateFacultyID(userID uint, facultyID uint) error {
	return s.userrepo.UpdateFacultyID(userID, facultyID)
}

func (s *UserService) SignIn(dto *dto.SignInDTO, c fiber.Ctx) (string, string, uint, string, string, error) {
	user, err := s.userrepo.FindByEmail(dto.Email)
	if err != nil {
		return "", "", 0, "", "", errors.New("invalid email or password")
	}

	if !user.IsActive {
		return "", "", 0, "", "", errors.New("account is inactive")
	}

	if !user.IsVerified {
		return "", "", 0, "", "", errors.New("please verify your email before signing in")
	}

	err = utils.ComparePassword(user.Password, dto.Password)
	if err != nil {
		return "", "", 0, "", "", errors.New("invalid email or password")
	}

	// Auto-detect and sync profile IDs if mapped in respective tables
	if user.StudentID == 0 {
		user.StudentID, _ = s.userrepo.GetUserStudentID(user.ID)
	}
	if user.FacultyID == 0 {
		user.FacultyID, _ = s.userrepo.GetUserFacultyID(user.ID)
	}
	if user.PrincipalID == 0 {
		user.PrincipalID, _ = s.userrepo.GetUserPrincipalID(user.ID)
	}

	primaryRole := ""
	rolesList, _ := s.userrepo.FetchUserRoles(user.ID)
	if len(rolesList) > 0 {
		primaryRole = rolesList[0].Name
	}

	go func(email, name string) {
		subject := "New Sign-In"

		body := fmt.Sprintf(
			"Hello %s,\n\nYour account has just been signed in successfully.\n\nIf this wasn't you, please contact support immediately.",
			name,
		)

		if sendErr := grpc.SendEmail(email, subject, body, "signin"); sendErr != nil {
			log.Printf("Failed to send sign-in email via gRPC: %v\n", sendErr)
		}
	}(user.Email, user.Name)
	
	sessionID := uuid.New().String()
	userAgent := c.Get("User-Agent")

	accessToken, err := utils.GenerateAccessToken(user.ID, sessionID)
	if err != nil {
		return "", "", 0, "", "", err
	}
	refreshToken, err := utils.GenerateRefreshToken(user.ID, sessionID)
	if err != nil {
		return "", "", 0, "", "", err
	}
	_, err = s.sessionService.CreateSession(user.ID, userAgent, sessionID, accessToken, refreshToken)
	if err != nil {
		return "", "", 0, "", "", err
	}

	return accessToken, refreshToken, user.ID, sessionID, primaryRole, nil
}

func (s *UserService) AssignRole(userID uint, roleName string) error {
	role, err := s.userrepo.FindRoleByName(roleName)
	if err != nil {
		return errors.New("role not found: " + roleName)
	}
	return s.userrepo.AssignRoleToUser(userID, role.ID)
}

func (s *UserService) DeleteUserService(id uint) error {
	return s.userrepo.DeleteUser(id)
}
func (s *UserService) ForgotPasswordService(mail dto.ForgotPasswordDTO) (model.User, error) {
	fetchemail, err := s.userrepo.ForgotPasswordRepo(mail)
	if err != nil {
		return model.User{}, err
	}
	subject := "Forgot Password mail"
	token := utils.ReseTToken()
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", getBaseURL(), token)

	body := fmt.Sprintf(`
	<h2>Hello %s,</h2>

	<p>We received a request to reset your password for your <strong>Backend Institutions</strong> account.</p>

	<p>Click the button below to reset your password:</p>

	<p>
		<a href="%s"
		   style="
				background-color:#4CAF50;
				color:white;
				padding:12px 24px;
				text-decoration:none;
				border-radius:5px;
				display:inline-block;">
			Reset Password
		</a>
	</p>

	<p>If the button doesn't work, copy and paste this link into your browser:</p>

	<p>%s</p>

	<p><strong>This link will expire in 15 minutes.</strong></p>

	<p>If you did not request a password reset, you can safely ignore this email. Your password will remain unchanged.</p>

	<br>

	<p>Regards,</p>
	<p><strong>Backend Institutions Team</strong></p>
`, fetchemail.Name, resetURL, resetURL)

	go func(email, subject, body, resetURL string) {
		if sendErr := grpc.SendEmail(email, subject, body, "forgot-password"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(fetchemail.Email, subject, body, resetURL)

	err = s.userrepo.UpdateResetToken(fetchemail)

	return fetchemail, err

}

func (s *UserService) ResetPasswordService(token string, reset dto.ResetPassword) error {
	user, err := s.userrepo.FetchUsertoken(token)
	if err != nil {
		return err
	}
	err = utils.ComparePassword(user.Password, reset.CurrentPassword)
	if err != nil {
		return errors.New("current password is incorrect")
	}
	hashedPassword, err := utils.HashPassword(reset.NewPassword)
	if err != nil {
		return err
	}
	fmt.Println("User ID:", user.ID)
	fmt.Println("New Password:", reset.NewPassword)
	fmt.Println("Hashed Password:", hashedPassword)
	err = s.userrepo.UpdatePassword(user.ID, hashedPassword)

	if err != nil {
		fmt.Println("Cant able to update the password")
	}
	return err

}

func (s *UserService) VerifyEmail(token string) error {
	user, err := s.userrepo.FindByVerificationToken(token)
	if err != nil {
		return errors.New("invalid verification link")
	}

	if user.IsVerified {
		return errors.New("email already verified")
	}

	if time.Now().After(user.TokenExpiresAt) {
		return errors.New("verification link expired")
	}

	user.IsVerified = true
	user.IsActive = true
	user.VerificationToken = ""
	user.TokenExpiresAt = time.Time{}

	return s.userrepo.UpdateUser(&user)
}

func (s *UserService) Logout(dto *dto.LogoutDTO) error {
	return s.userrepo.Logout(dto)
}

func (s *UserService) ResendMail(email string) error {
	user, err := s.userrepo.FindByEmail(email)
	if err != nil {
		return errors.New("cannot find email")
	}

	if user.IsVerified {
		return errors.New("email is already verified")
	}

	token := utils.SignUpToken()
	if token == "" {
		return errors.New("failed to generate verification token")
	}

	user.VerificationToken = token
	user.TokenExpiresAt = time.Now().Add(24 * time.Hour)
	err = s.userrepo.UpdateUser(&user)
	if err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", getBaseURL(), token)
	subject := "Verify your email - Backend Institutions"
	body := fmt.Sprintf(`<h1>Hello %s,</h1>
<p>Thank you for signing up. Please verify your email by clicking the link below:</p>
<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; border-radius: 5px;">Verify Email Address</a></p>
<p>Or copy and paste this link in your browser:<br/><a href="%s">%s</a></p>
<p>This link will expire in 24 hours.</p>
<p>If you did not create this account, please ignore this email.</p>`, user.Name, verifyURL, verifyURL, verifyURL)

	go func(email, subject, body string) {
		if sendErr := grpc.SendEmail(email, subject, body, "signup"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(user.Email, subject, body)

	return nil
}

func (s *UserService) GetProfileByID(id uint) (model.User, error) {
	user, err := s.userrepo.FindByID(id)
	if err != nil {
		return model.User{}, err
	}
	if user.StudentID == 0 {
		user.StudentID, _ = s.userrepo.GetUserStudentID(id)
	}
	if user.FacultyID == 0 {
		user.FacultyID, _ = s.userrepo.GetUserFacultyID(id)
	}
	if user.PrincipalID == 0 {
		user.PrincipalID, _ = s.userrepo.GetUserPrincipalID(id)
	}
	roles, _ := s.userrepo.FetchUserRoles(id)
	user.Roles = roles
	return user, nil
}
