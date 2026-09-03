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

// getBaseURL resolves the application's base URL from environment configuration
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

// UserService provides business logic operations for user accounts, auth, and sessions
type UserService struct {
	userrepo       *repository.UserRepository
	sessionService *SessionService
	roleRepo *repository.RoleRepository
}

// NewUserService instantiates a new UserService with user and session dependencies
func NewUserService(userrepo *repository.UserRepository, sessionService *SessionService, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{
		userrepo:       userrepo,
		sessionService: sessionService,
		roleRepo: roleRepo,
	}
}

// SignUp handles user registration and triggers email verification dispatch
func (s *UserService) SignUp(dto *dto.SignUpDTO) (model.User, error) {
	// 1. Hash user password securely
	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		return model.User{}, err
	}

	// 2. Generate cryptographically secure verification token
	token := utils.SignUpToken()

	// 3. Assemble new user entity
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

	// 4. Save user record to database
	err = s.userrepo.CreateUser(&user)
	if err != nil {
		return model.User{}, err
	}

	// 5. Construct verification email payload
	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", getBaseURL(), token)
	subject := "Verify your email - Backend Institutions"
	body := fmt.Sprintf(`<h1>Hello %s,</h1>
<p>Thank you for signing up. Please verify your email by clicking the link below:</p>
<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; border-radius: 5px;">Verify Email Address</a></p>
<p>Or copy and paste this link in your browser:<br/><a href="%s">%s</a></p>
<p>This link will expire in 24 hours.</p>`, user.Name, verifyURL, verifyURL, verifyURL)

	// 6. Asynchronously send verification email via gRPC
	go func(email, subject, body string) {
		if sendErr := grpc.SendEmail(email, subject, body, "signup"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(user.Email, subject, body)

	return user, nil
}

// SignUpWithRole registers a user and automatically assigns a specified role
func (s *UserService) SignUpWithRole(
	dto *dto.SignUpDTO,
	roleName string,
) (model.User, error) {

	// 1. Clean role name
	roleName = strings.TrimSpace(roleName)

	if roleName == "" {
		return model.User{}, errors.New("role is required")
	}

	// 2. Check whether role exists
	role, err := s.roleRepo.GetRoleByName(roleName)
	if err != nil {
		return model.User{}, errors.New("role not found")
	}

	// 3. Hash password
	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		return model.User{}, err
	}

	// 4. Generate email verification token
	token := utils.SignUpToken()

	// 5. Create user object
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

	// 6. Create user in database
	if err := s.userrepo.CreateUser(&user); err != nil {
		return model.User{}, err
	}
	

	// 7. Map user with role
	if err := s.userrepo.AssignRole(user.ID, role.ID); err != nil {
		return model.User{}, err
	}

	// 8. Generate verification URL
	verifyURL := fmt.Sprintf(
		"%s/auth/verify?token=%s",
		getBaseURL(),
		token,
	)

	// 9. Prepare email
	subject := "Verify your email - Backend Institutions"

	body := fmt.Sprintf(`
		<h1>Hello %s,</h1>

		<p>Thank you for signing up.</p>

		<p>
			<a href="%s">
				Verify Email Address
			</a>
		</p>

		<p>This link will expire in 24 hours.</p>
	`,
		user.Name,
		verifyURL,
	)

	// 10. Send email in background
	go func(email, subject, body string) {

		 grpc.SendEmail(
			email,
			subject,
			body,
			"signup",
		); 
	}(user.Email, subject, body)

	return user, nil
}

// CheckUserRole checks if user has been assigned a specific role
func (s *UserService) CheckUserRole(userID uint, targetRole string) (bool, error) {
	return s.userrepo.CheckUserRole(userID, targetRole)
}

// UpdateStudentID links a student ID to a user account
func (s *UserService) UpdateStudentID(userID uint, studentID uint) error {
	return s.userrepo.UpdateStudentID(userID, studentID)
}

// GetUserByID fetches user model by user ID
func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	return s.userrepo.GetUserByID(userID)
}

// UpdateFacultyID links a faculty ID to a user account
func (s *UserService) UpdateFacultyID(userID uint, facultyID uint) error {
	return s.userrepo.UpdateFacultyID(userID, facultyID)
}

// SignIn authenticates user credentials, checks activation, and creates active session tokens
func (s *UserService) SignIn(dto *dto.SignInDTO, c fiber.Ctx) (string, string, uint, string, string, error) {
	// 1. Fetch user by email
	user, err := s.userrepo.FindByEmail(dto.Email)
	if err != nil {
		return "", "", 0, "", "", errors.New("invalid email or password")
	}

	// 2. Verify account is active and verified
	if !user.IsActive {
		return "", "", 0, "", "", errors.New("account is inactive")
	}

	if !user.IsVerified {
		return "", "", 0, "", "", errors.New("please verify your email before signing in")
	}

	// 3. Compare password hash
	err = utils.ComparePassword(user.Password, dto.Password)
	if err != nil {
		return "", "", 0, "", "", errors.New("invalid email or password")
	}


	// 5. Resolve user's primary assigned role
	
	rolesList, _ := s.userrepo.FetchUserRoles(user.ID)
	

if len(rolesList) == 0 {
	return "", "", 0, "", "", errors.New("user role not found")
}

primaryRole := rolesList[0].Name
	
	
	

	// 6. Send security login notification in background
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

	// 7. Generate JWT access and refresh tokens
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

	// 8. Persist active session in database
	_, err = s.sessionService.CreateSession(user.ID, userAgent, sessionID, accessToken, refreshToken)
	if err != nil {
		return "", "", 0, "", "", err
	}

	return accessToken, refreshToken, user.ID, sessionID, primaryRole, nil
}

// AssignRole assigns a role to a user by role name
func (s *UserService) AssignRole(userID uint, roleName string) error {
	// 1. Look up role by name
	role, err := s.userrepo.FindRoleByName(roleName)
	if err != nil {
		return errors.New("role not found: " + roleName)
	}

	// 2. Assign role mapping in user_roles
	return s.userrepo.AssignRoleToUser(userID, role.ID)
}

// DeleteUserService soft deletes a user record
func (s *UserService) DeleteUserService(id uint) error {
	return s.userrepo.DeleteUser(id)
}

// ForgotPasswordService initiates password reset request and sends reset email
func (s *UserService) ForgotPasswordService(mail dto.ForgotPasswordDTO) (model.User, error) {
	// 1. Verify user email existence
	fetchemail, err := s.userrepo.ForgotPasswordRepo(mail)
	if err != nil {
		return model.User{}, err
	}

	// 2. Generate password reset token
	subject := "Forgot Password mail"
	token := utils.ReseTToken()
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", getBaseURL(), token)

	// 3. Assemble reset email body
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

	// 4. Asynchronously send reset email via gRPC
	go func(email, subject, body, resetURL string) {
		if sendErr := grpc.SendEmail(email, subject, body, "forgot-password"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(fetchemail.Email, subject, body, resetURL)

	// 5. Store reset token in database
	err = s.userrepo.UpdateResetToken(fetchemail)

	return fetchemail, err
}

// ResetPasswordService validates current password and updates with new password
func (s *UserService) ResetPasswordService(token string, reset dto.ResetPassword) error {
	// 1. Fetch user by reset token
	user, err := s.userrepo.FetchUsertoken(token)
	if err != nil {
		return err
	}

	// 2. Validate current password
	err = utils.ComparePassword(user.Password, reset.CurrentPassword)
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// 3. Hash new password
	hashedPassword, err := utils.HashPassword(reset.NewPassword)
	if err != nil {
		return err
	}

	// 4. Update password in database
	err = s.userrepo.UpdatePassword(user.ID, hashedPassword)
	return err
}

// VerifyEmail marks user account as verified using confirmation token
func (s *UserService) VerifyEmail(token string) error {
	// 1. Fetch user by verification token
	user, err := s.userrepo.FindByVerificationToken(token)
	if err != nil {
		return errors.New("invalid verification link")
	}

	// 2. Check if already verified
	if user.IsVerified {
		return errors.New("email already verified")
	}

	// 3. Check token expiration
	if time.Now().After(user.TokenExpiresAt) {
		return errors.New("verification link expired")
	}

	// 4. Update user verification flags and clear token
	user.IsVerified = true
	user.IsActive = true
	user.VerificationToken = ""
	user.TokenExpiresAt = time.Time{}

	return s.userrepo.UpdateUser(&user)
}

// Logout terminates user session
func (s *UserService) Logout(dto *dto.LogoutDTO) error {
	return s.userrepo.Logout(dto)
}

// ResendMail resends verification email to an unverified user
func (s *UserService) ResendMail(email string) error {
	// 1. Locate user by email
	user, err := s.userrepo.FindByEmail(email)
	if err != nil {
		return errors.New("cannot find email")
	}

	// 2. Ensure user is not already verified
	if user.IsVerified {
		return errors.New("email is already verified")
	}

	// 3. Generate new verification token
	token := utils.SignUpToken()
	if token == "" {
		return errors.New("failed to generate verification token")
	}

	// 4. Update token and expiry on user record
	user.VerificationToken = token
	user.TokenExpiresAt = time.Now().Add(24 * time.Hour)
	err = s.userrepo.UpdateUser(&user)
	if err != nil {
		return err
	}

	// 5. Construct email content
	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", getBaseURL(), token)
	subject := "Verify your email - Backend Institutions"
	body := fmt.Sprintf(`<h1>Hello %s,</h1>
<p>Thank you for signing up. Please verify your email by clicking the link below:</p>
<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; display: inline-block; border-radius: 5px;">Verify Email Address</a></p>
<p>Or copy and paste this link in your browser:<br/><a href="%s">%s</a></p>
<p>This link will expire in 24 hours.</p>
<p>If you did not create this account, please ignore this email.</p>`, user.Name, verifyURL, verifyURL, verifyURL)

	// 6. Send email via gRPC in background
	go func(email, subject, body string) {
		if sendErr := grpc.SendEmail(email, subject, body, "signup"); sendErr != nil {
			log.Printf("Failed to send verification email via gRPC: %v\n", sendErr)
		}
	}(user.Email, subject, body)

	return nil
}

// GetProfileByID fetches full profile and relations for a user ID
func (s *UserService) GetProfileByID(id uint) (model.User, error) {
	// 1. Fetch user by ID
	user, err := s.userrepo.FindByID(id)
	if err != nil {
		return model.User{}, err
	}

	// 2. Resolve student ID if missing
	if user.StudentID == 0 {
		user.StudentID, _ = s.userrepo.GetUserStudentID(id)
	}

	// 3. Resolve faculty ID if missing
	if user.FacultyID == 0 {
		user.FacultyID, _ = s.userrepo.GetUserFacultyID(id)
	}

	// 4. Fetch assigned roles
	roles, _ := s.userrepo.FetchUserRoles(id)
	user.Roles = roles

	return user, nil
}
