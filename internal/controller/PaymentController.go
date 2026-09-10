package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"

	"backend_institutions/internal/services"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type PaymentController struct {
	paymentService *services.PaymentService
}

func NewPaymentController(paymentService *services.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
	}
}

func (cl *PaymentController) CreateDepartmentPaymentController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "unauthorized user")
	}

	var body dto.CreateDepartmentPaymentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	payment, err := cl.paymentService.CreateDepartmentPayment(userID, &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cant able to manage") ||
			strings.Contains(strings.ToLower(err.Error()), "only institution admin") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Department payment configured successfully", dto.ToDepartmentPaymentResponseDTO(payment))
}

func (cl *PaymentController) UpdateDepartmentPaymentController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "unauthorized user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid payment id")
	}

	var body dto.UpdateDepartmentPaymentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	updated, err := cl.paymentService.UpdateDepartmentPayment(userID, uint(id), &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cant able to manage") ||
			strings.Contains(strings.ToLower(err.Error()), "only institution admin") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Department payment updated successfully", dto.ToDepartmentPaymentResponseDTO(updated))
}

func (cl *PaymentController) DeleteDepartmentPaymentController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "unauthorized user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid payment id")
	}

	if err := cl.paymentService.DeleteDepartmentPayment(userID, uint(id)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cant able to manage") ||
			strings.Contains(strings.ToLower(err.Error()), "only institution admin") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Department payment deleted successfully", nil)
}

func (cl *PaymentController) GetDepartmentPaymentBySemesterController(c fiber.Ctx) error {
	deptIDStr := c.Params("departmentId")
	deptID, err := strconv.ParseUint(deptIDStr, 10, 32)
	if err != nil || deptID == 0 {
		return helper.Error(c, 400, "invalid department id")
	}

	semStr := c.Params("semester")
	sem, err := strconv.ParseUint(semStr, 10, 32)
	if err != nil || sem == 0 {
		return helper.Error(c, 400, "invalid semester")
	}

	payment, err := cl.paymentService.GetDepartmentPaymentBySemester(uint(deptID), uint(sem))
	if err != nil || payment == nil {
		return helper.Error(c, 404, "payment configuration not found for this department and semester")
	}

	return helper.Success(c, "Department payment fetched successfully", dto.ToDepartmentPaymentResponseDTO(payment))
}

func (cl *PaymentController) GetDepartmentPaymentsController(c fiber.Ctx) error {
	deptIDStr := c.Params("departmentId")
	deptID, err := strconv.ParseUint(deptIDStr, 10, 32)
	if err != nil || deptID == 0 {
		return helper.Error(c, 400, "invalid department id")
	}

	payments, err := cl.paymentService.GetDepartmentPayments(uint(deptID))
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "Department payments fetched successfully", dto.ToDepartmentPaymentResponseListDTO(payments))
}

func (cl *PaymentController) CreateStudentPaymentController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "unauthorized user")
	}

	var body dto.CreateStudentPaymentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	payment, updatedStudent, err := cl.paymentService.MakeStudentPayment(userID, &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cant able to make payment for other student") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Payment processed successfully", dto.ToStudentPaymentResponseDTO(payment, updatedStudent))
}
