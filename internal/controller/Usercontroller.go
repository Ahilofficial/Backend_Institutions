package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/model"
	"backend_institutions/internal/services"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (cl *UserController) SignUpController(c fiber.Ctx) error {
	role := strings.TrimSpace(c.Params("role"))
	if role == "" {
		role = strings.TrimSpace(c.Query("role"))
	}
	role = strings.ReplaceAll(role, "-", " ")

	var body dto.SignUpDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	var user model.User
	var err error

	if role != "" {
		user, err = cl.userService.SignUpWithRole(&body, role)
	} else {
		user, err = cl.userService.SignUp(&body)
	}

	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	msg := "Signed up successfully"
	if role != "" {
		msg = fmt.Sprintf("Signed up successfully as %s", role)
	}

	return helper.Success(c, msg, dto.ToUserResponseDTO(&user))
}

func (cl *UserController) SignInController(c fiber.Ctx) error {
	var body dto.SignInDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	accessToken, refreshToken, user_id, session_id, role, err := cl.userService.SignIn(&body, c)
	if err != nil {
		return helper.Error(c, 401, err.Error())
	}

	return helper.Success(c, "Signed in successfully", dto.AuthResponseDTO{
		UserID:       user_id,
		SessionID:    session_id,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Role:         role,
	})
}

func (c *UserController) VerifyEmail(ctx fiber.Ctx) error {
	token := ctx.Query("token")

	if token == "" {
		return helper.Error(ctx, fiber.StatusBadRequest, "Verification token is required")
	}

	err := c.userService.VerifyEmail(token)
	if err != nil {
		return helper.Error(ctx, fiber.StatusBadRequest, err.Error())
	}

	return helper.Success(ctx, "Email verified successfully", nil)
}

func (cl *UserController) AssignRoleController(c fiber.Ctx) error {
	var body dto.AssignRoleDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.userService.AssignRole(body.UserID, body.Role); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Role assigned successfully", nil)
}

func (cl *UserController) DeleteUserController(c fiber.Ctx) error {
	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid user id")
	}

	if err := cl.userService.DeleteUserService(uint(id)); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "User deleted successfully", nil)
}

func (cl *UserController) ForgotPassword(c fiber.Ctx) error {
	var forgotpassword dto.ForgotPasswordDTO
	err := c.Bind().Body(&forgotpassword)
	if err != nil {
		return helper.Error(c, 400, "Invalid email format")
	}
	_, err = cl.userService.ForgotPasswordService(forgotpassword)

	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return c.SendString("Sended the forgot password link just check it")
}

func (cl *UserController) ResetPassword(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return helper.Error(c, 400, "Token is required")
	}

	var reset dto.ResetPassword
	if err := c.Bind().Body(&reset); err != nil {
		return helper.Error(c, 400, "Invalid payload format")
	}

	fmt.Println("Token:", token)
	fmt.Println("Reset Body:", reset)

	err := cl.userService.ResetPasswordService(token, reset)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Password reset successfully", nil)
}

func (cl *UserController) Logout(c fiber.Ctx) error {
	var body dto.LogoutDTO
	_ = c.Bind().Body(&body)

	if body.UserID == 0 {
		if uID, ok := c.Locals("user_id").(uint); ok {
			body.UserID = uID
		}
	}

	_ = cl.userService.Logout(&body)
	return helper.Success(c, "Logout successful", nil)
}

func (c *UserController) ResendMail(ctx fiber.Ctx) error {
	var body dto.ResendMailSignUp

	if err := ctx.Bind().Body(&body); err != nil {
		return helper.Error(ctx, 400, err.Error())
	}
	err := c.userService.ResendMail(body.Email)
	if err != nil {
		return helper.Error(ctx, 400, err.Error())
	}

	return helper.Success(ctx, "mail sent successfully", nil)
}

func (cl *UserController) GetProfile(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user ID")
	}

	user, err := cl.userService.GetProfileByID(userID)
	if err != nil {
		return helper.Error(c, 404, "User not found")
	}

	return helper.Success(c, "User profile fetched successfully", dto.ToUserResponseDTO(&user))
}
