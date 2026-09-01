package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/services"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type FeesController struct {
	feesService *services.FeesService
	
}

func NewFeesController(feesService *services.FeesService, ) *FeesController {
	return &FeesController{feesService: feesService, }
}

func (cl *FeesController) CreateFeesController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	var body dto.CreateFeesDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	fee, err := cl.feesService.CreateDepartmentFee(userID, &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Department fee template created successfully", dto.ToFeesResponseDTO(fee))
}

func (cl *FeesController) GetDepartmentFeeBySemesterController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	deptID, err := strconv.ParseUint(c.Params("departmentId"), 10, 32)
	if err != nil || deptID == 0 {
		return helper.Error(c, 400, "invalid department ID")
	}

	semester, err := strconv.ParseUint(c.Params("semester"), 10, 32)
	if err != nil || semester == 0 {
		return helper.Error(c, 400, "invalid semester")
	}

	fee, err := cl.feesService.GetDepartmentFeeBySemester(userID, uint(deptID), uint(semester))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(c, "Department fee fetched successfully", dto.ToFeesResponseDTO(fee))
}

func (cl *FeesController) GetDepartmentFeesController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	deptID, err := strconv.ParseUint(c.Params("departmentId"), 10, 32)
	if err != nil || deptID == 0 {
		return helper.Error(c, 400, "invalid department ID")
	}

	fees, err := cl.feesService.GetDepartmentFees(userID, uint(deptID))
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Department fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

func (cl *FeesController) GetAllFeesController(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)
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

	fees, total, err := cl.feesService.GetFeesServicePaginated(userID, search, page, limit)
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
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid fees id")
	}

	fee, err := cl.feesService.GetFeesByIDService(userID, uint(id))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(c, "Fees fetched successfully", dto.ToFeesResponseDTO(&fee))
}

func (cl *FeesController) UpdateFeesController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
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

	updated, err := cl.feesService.UpdateDepartmentFee(userID, uint(id), &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Fees updated successfully", dto.ToFeesResponseDTO(updated))
}

func (cl *FeesController) DeleteFeesController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid fees id")
	}

	if err := cl.feesService.DeleteDepartmentFee(userID, uint(id)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Fees deleted successfully", nil)
}

func (cl *FeesController) FetchAllFeesController(c fiber.Ctx) error {
	fees, err := cl.feesService.GetFeesService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "All fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

func (cl *FeesController) FetchFeesByStudentID(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid student id")
	}

	fees, err := cl.feesService.FetchFeesByStudentID(userID, uint(id))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 404, "Fees not found")
	}

	return helper.Success(c, "Fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

func (cl *FeesController) GetMyFeesController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	fees, err := cl.feesService.GetMyFees(userID)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(c, "Student fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

func (cl *FeesController) CreatePayment(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}
	

	var req dto.CreatePaymentDTO
	if err := c.Bind().Body(&req); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	req.Sanitize()
	if err := req.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	payment, err := cl.feesService.CreatePayment(userID, &req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}
	loggined_student_id:=cl.feesService.GetLogginedStudentID(userID)
	if loggined_student_id!=req.StudentID{
		return helper.Error(c, 403, "cant pay fees for other student")
	}
	

	return helper.Success(c, "Payment created successfully", payment)
}

func (cl *FeesController) GetPaymentByIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid payment id")
	}

	payment, err := cl.feesService.GetPaymentByID(userID, uint(id))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(c, "Payment fetched successfully", payment)
}

func (cl *FeesController) GetPaymentByFeeIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	feeIDStr := c.Params("fee_id")
	feeID, err := strconv.ParseUint(feeIDStr, 10, 32)
	if err != nil || feeID == 0 {
		return helper.Error(c, 400, "invalid fee id")
	}

	payments, err := cl.feesService.GetPaymentByFeeID(userID, uint(feeID))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(c, "Payments fetched successfully", payments)
}
