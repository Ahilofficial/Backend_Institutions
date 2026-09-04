package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/services"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// UserController handles user authentication, registration, password recovery, and profile endpoints
type UserController struct {
	userService *services.UserService
}

// NewUserController instantiates a new UserController with the user service dependency
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// SignUpController handles user registration with optional role assignment and email verification dispatch
func (cl *UserController) SignUpController(c fiber.Ctx) error {

	// 1. Get role from route parameter
	roleName := strings.TrimSpace(c.Params("role"))

	if roleName == "" {
		return helper.Error(c, 400, "role is required")
	}

	roleName = strings.ReplaceAll(roleName, "-", " ")
	roleName = strings.ReplaceAll(roleName, "_", " ")

	// 2. Bind request body
	var body dto.SignUpDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(
			c,
			400,
			"invalid request body: "+err.Error(),
		)
	}

	// 3. Sanitize and validate request data
	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Create user and assign role
	user, err := cl.userService.SignUpWithRole(
		&body,
		roleName,
	)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success response
	msg := fmt.Sprintf(
		"Signed up successfully as %s. Please verify your email.",
		roleName,
	)

	return helper.Success(
		c,
		msg,
		dto.ToUserResponseDTO(&user),
	)
}

// SignInController handles user login, password verification, and JWT token issuing
func (cl *UserController) SignInController(c fiber.Ctx) error {
	// 1. Bind request body
	var body dto.SignInDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 2. Validate input fields
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Authenticate credentials and generate session tokens
	accessToken, refreshToken, user_id, session_id, role, err := cl.userService.SignIn(&body, c)
	if err != nil {
		return helper.Error(c, 401, err.Error())
	}

	// 4. Return token credentials
	return helper.Success(c, "Signed in successfully", dto.AuthResponseDTO{
		UserID:       user_id,
		SessionID:    session_id,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         role,
	})
}

// VerifyEmail verifies email using the verification token sent via email
func (c *UserController) VerifyEmail(ctx fiber.Ctx) error {
	// 1. Extract verification token from query parameter
	token := ctx.Query("token")
	if token == "" {
		return helper.Error(ctx, fiber.StatusBadRequest, "Verification token is required")
	}

	// 2. Verify token and activate user account via service
	err := c.userService.VerifyEmail(token)
	if err != nil {
		return helper.Error(ctx, fiber.StatusBadRequest, err.Error())
	}

	// 3. Return confirmation
	return helper.Success(ctx, "Email verified successfully", nil)
}

// AssignRoleController assigns a specified role to a user
func (cl *UserController) AssignRoleController(c fiber.Ctx) error {
	// 1. Bind and sanitize request body
	var body dto.AssignRoleDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 2. Validate inputs
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Assign role to user via service
	if err := cl.userService.AssignRole(body.UserID, body.Role); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return response
	return helper.Success(c, "Role assigned successfully", nil)
}

// DeleteUserController soft deletes a user account
func (cl *UserController) DeleteUserController(c fiber.Ctx) error {
	// 1. Parse user ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid user id")
	}

	// 2. Delete user via service
	if err := cl.userService.DeleteUserService(uint(id)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Return response
	return helper.Success(c, "User deleted successfully", nil)
}

// ForgotPassword initiates password reset workflow and dispatches reset token email
func (cl *UserController) ForgotPassword(c fiber.Ctx) error {
	// 1. Bind request body
	var forgotpassword dto.ForgotPasswordDTO
	if err := c.Bind().Body(&forgotpassword); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 2. Sanitize and validate inputs
	forgotpassword.Sanitize()
	if err := forgotpassword.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Generate token and dispatch email via service
	_, err := cl.userService.ForgotPasswordService(forgotpassword)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return response
	return helper.Success(c, "Sent the forgot password link successfully", nil)
}

// ResetPassword applies new password using a verified reset token
func (cl *UserController) ResetPassword(c fiber.Ctx) error {
	// 1. Extract reset token from query string
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		return helper.Error(c, 400, "token is required")
	}

	// 2. Bind new password payload
	var reset dto.ResetPassword
	if err := c.Bind().Body(&reset); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 3. Sanitize and validate
	reset.Sanitize()
	if err := reset.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update user password via service
	err := cl.userService.ResetPasswordService(token, reset)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return response
	return helper.Success(c, "Password reset successfully", nil)
}

// Logout terminates current session and revokes tokens
func (cl *UserController) Logout(c fiber.Ctx) error {
	// 1. Bind logout parameters
	var body dto.LogoutDTO
	err:= c.Bind().Body(&body)
	if err != nil {
		return err
	}

	// 2. Fallback to authenticated user ID from context
	if body.UserID == 0 {
		if uID, ok := c.Locals("user_id").(uint); ok {
			body.UserID = uID
		}
	}

	// 3. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Invalidate session via service
	_ = cl.userService.Logout(&body)
	return helper.Success(c, "Logout successful", nil)
}

// ResendMail resends the email verification message to an unverified user
func (c *UserController) ResendMail(ctx fiber.Ctx) error {
	// 1. Bind request body
	var body dto.ResendMailSignUp
	if err := ctx.Bind().Body(&body); err != nil {
		return helper.Error(ctx, 400, "invalid request body: "+err.Error())
	}

	// 2. Sanitize and validate
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(ctx, 400, err.Error())
	}

	// 3. Resend email via service
	err := c.userService.ResendMail(body.Email)
	if err != nil {
		return helper.Error(ctx, 400, err.Error())
	}

	// 4. Return response
	return helper.Success(ctx, "Verification email sent successfully", nil)
}

// GetProfile fetches the user account profile details
func (cl *UserController) GetProfile(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user ID")
	}

	// 2. Fetch user profile from service
	user, err := cl.userService.GetProfileByID(userID)
	if err != nil {
		return helper.Error(c, 404, "User not found")
	}

	// 3. Return user profile DTO
	return helper.Success(c, "User profile fetched successfully", dto.ToUserResponseDTO(&user))
}
