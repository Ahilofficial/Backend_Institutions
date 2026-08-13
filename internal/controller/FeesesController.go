package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/model"
	"backend_institutions/internal/services"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type FeesController struct {
	feesService *services.FeesService
}

func NewFeesController(feesService *services.FeesService) *FeesController {
	return &FeesController{feesService: feesService}
}

func (cl *FeesController) GetInactiveFeesController(c fiber.Ctx) error {
	fees, err := cl.feesService.GetInactiveFeesService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Inactive fees fetched successfully",
		dto.ToFeesResponseListDTO(fees),
	)
}

func (cl *FeesController) CreateFeesController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	var body dto.CreateFeesDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	var feeObj = model.Fees{
		PaymentMode:   body.PaymentMode,
		TotalAmount:   body.TotalAmount,
		StudentID:     body.StudentID,
		TotalPaid:     0,
		PendingAmount: body.TotalAmount,
	}

	fees, err := cl.feesService.CreateFeesService(
		userID,
		&feeObj,
	)
	if err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Fees created successfully",
		dto.ToFeesResponseDTO(&fees),
	)
}

func (cl *FeesController) GetAllFeesController(c fiber.Ctx) error {
	search := c.Query("search")
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	fees, total, err := cl.feesService.GetFeesServicePaginated(search, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return helper.Success(
		c,
		"Fees fetched successfully",
		fiber.Map{
			"items":       dto.ToFeesResponseListDTO(fees),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}

func (cl *FeesController) GetFeesByIDController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid fees id")
	}

	fees, err := cl.feesService.GetFeesServiceById(
		userID,
		uint(id),
	)
	if err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Fees fetched successfully",
		dto.ToFeesResponseDTO(&fees),
	)
}

func (cl *FeesController) UpdateFeesController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid fees id")
	}

	var body dto.UpdateFeesDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.feesService.UpdateFeesService(
		userID,
		uint(id),
		&body,
	); err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	// Get updated fee.
	updated, err := cl.feesService.GetFeesServiceById(
		userID,
		uint(id),
	)
	if err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"Fees updated successfully",
		dto.ToFeesResponseDTO(&updated),
	)
}

func (cl *FeesController) DeleteFeesController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid fees id")
	}

	if err := cl.feesService.DeleteFeesService(
		userID,
		uint(id),
	); err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Fees deleted successfully",
		nil,
	)
}

func (cl *FeesController) FetchAllFeesController(c fiber.Ctx) error {
	fees, err := cl.feesService.GetFeesService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"All fees fetched successfully",
		dto.ToFeesResponseListDTO(fees),
	)
}

func (cl *FeesController) GetPaymentByIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid payment id")
	}

	payment, err := cl.feesService.GetPaymentByID(userID, uint(id))
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Payment fetched successfully",
		payment,
	)
}

func (cl *FeesController) GetPaymentByFeeIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	feeIDStr := c.Params("fee_id")

	feeID, err := strconv.ParseUint(feeIDStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid fee id")
	}

	payments, err := cl.feesService.GetPaymentByFeeID(userID, uint(feeID))
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Payments fetched successfully",
		payments,
	)
}

func (c *FeesController) CreatePayment(ctx fiber.Ctx) error {

	userID, ok := ctx.Locals("user_id").(uint)
	if !ok {
		return helper.Error(ctx, 401, "Invalid user")
	}

	var req dto.CreatePaymentDTO

	if err := ctx.Bind().Body(&req); err != nil {
		return helper.Error(ctx, 400, "Invalid request body")
	}

	req.Sanitize()

	if err := req.Validate(); err != nil {
		return helper.Error(ctx, 400, err.Error())
	}

	payment, err := c.feesService.CreatePayment(
		userID,
		&req,
	)
	if err != nil {

		if err.Error() == "access denied" {
			return helper.Error(ctx, 403, "Access denied")
		}

		return helper.Error(ctx, 400, err.Error())
	}

	return helper.Success(
		ctx,
		"Payment created successfully",
		payment,
	)
}

func (c *StudentController) FetchPaidStudents(ctx fiber.Ctx) error {
	students, err := c.studentService.FetchPaidStudentsService()
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(ctx, "Paid students fetched successfully", students)
}

func (c *StudentController) FetchNotPaidStudents(ctx fiber.Ctx) error {
	students, err := c.studentService.FetchNotPaidStudentsService()
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(ctx, "Not paid students fetched successfully", students)
}

func (c *FeesController) FetchFeesByStudentID(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(uint)
	if !ok {
		return helper.Error(ctx, 401, "Invalid user")
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return helper.Error(ctx, fiber.StatusBadRequest, "Invalid student id")
	}

	fees, err := c.feesService.FetchFeesByStudentID(userID, uint(id))
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(ctx, 403, "Access denied")
		}
		return helper.Error(ctx, fiber.StatusNotFound, "Fees not found")
	}

	return helper.Success(ctx, "Fees fetched successfully", fees)
}
