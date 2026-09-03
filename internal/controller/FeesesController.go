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

// FeesController handles HTTP requests for fee templates and payment operations
type FeesController struct {
	feesService    *services.FeesService
	studentService *services.StudentService
}

// NewFeesController instantiates a new FeesController
func NewFeesController(feesService *services.FeesService, studentService *services.StudentService) *FeesController {
	return &FeesController{
		feesService:    feesService,
		studentService: studentService,
	}
}

// CreateFeesController handles creating a new fee template for a department and semester
func (cl *FeesController) CreateFeesController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	// 2. Parse request JSON body
	var body dto.CreateFeesDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 3. Sanitize and validate request fields
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Create department fee template via service
	fee, err := cl.feesService.CreateDepartmentFee(userID, &body)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success response
	return helper.Success(c, "Department fee template created successfully", dto.ToFeesResponseDTO(fee))
}

// GetDepartmentFeeBySemesterController fetches fee configuration for a department and semester
func (cl *FeesController) GetDepartmentFeeBySemesterController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	// 2. Parse department ID parameter
	deptID, err := strconv.ParseUint(c.Params("departmentId"), 10, 32)
	if err != nil || deptID == 0 {
		return helper.Error(c, 400, "invalid department ID")
	}

	// 3. Parse semester parameter
	semester, err := strconv.ParseUint(c.Params("semester"), 10, 32)
	if err != nil || semester == 0 {
		return helper.Error(c, 400, "invalid semester")
	}

	// 4. Fetch fee record from service
	fee, err := cl.feesService.GetDepartmentFeeBySemester(userID, uint(deptID), uint(semester))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	// 5. Return response
	return helper.Success(c, "Department fee fetched successfully", dto.ToFeesResponseDTO(fee))
}

// GetDepartmentFeesController fetches all fee templates for a department
func (cl *FeesController) GetDepartmentFeesController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	// 2. Parse department ID parameter
	deptID, err := strconv.ParseUint(c.Params("departmentId"), 10, 32)
	if err != nil || deptID == 0 {
		return helper.Error(c, 400, "invalid department ID")
	}

	// 3. Fetch fees list from service
	fees, err := cl.feesService.GetDepartmentFees(userID, uint(deptID))
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return list response
	return helper.Success(c, "Department fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

// GetAllFeesController retrieves paginated fee records
func (cl *FeesController) GetAllFeesController(c fiber.Ctx) error {
	// 1. Parse pagination and search query params
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

	// 2. Fetch paginated fees from service
	fees, total, err := cl.feesService.GetFeesServicePaginated(userID, search, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Compute total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 4. Return formatted response map
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

// GetFeesByIDController fetches a specific fee record by ID
func (cl *FeesController) GetFeesByIDController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse fee ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid fees id")
	}

	// 3. Fetch fee details via service
	fee, err := cl.feesService.GetFeesByIDService(userID, uint(id))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 404, err.Error())
	}

	// 4. Return response
	return helper.Success(c, "Fees fetched successfully", dto.ToFeesResponseDTO(&fee))
}

// UpdateFeesController updates amounts for a department fee template
func (cl *FeesController) UpdateFeesController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse fee ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid fees id")
	}

	// 3. Parse and validate update request body
	var body dto.UpdateFeesDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update fee record via service
	updated, err := cl.feesService.UpdateDepartmentFee(userID, uint(id), &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return updated fee response
	return helper.Success(c, "Fees updated successfully", dto.ToFeesResponseDTO(updated))
}

// DeleteFeesController deletes a department fee template
func (cl *FeesController) DeleteFeesController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse fee ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid fees id")
	}

	// 3. Delete fee template via service
	if err := cl.feesService.DeleteDepartmentFee(userID, uint(id)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success response
	return helper.Success(c, "Fees deleted successfully", nil)
}

// FetchAllFeesController retrieves all fees across the application
func (cl *FeesController) FetchAllFeesController(c fiber.Ctx) error {
	fees, err := cl.feesService.GetFeesService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(c, "All fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

// FetchFeesByStudentID retrieves all semester fee records for a student
func (cl *FeesController) FetchFeesByStudentID(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse student ID parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid student id")
	}

	// 3. Fetch fees for this student via service
	fees, err := cl.feesService.FetchFeesByStudentID(userID, uint(id))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 404, "Fees not found")
	}

	// 4. Return list response
	return helper.Success(c, "Fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

// GetMyFeesController retrieves fee history for the logged-in student
func (cl *FeesController) GetMyFeesController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Fetch fees for authenticated student
	fees, err := cl.feesService.GetMyFees(userID)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	// 3. Return response
	return helper.Success(c, "Student fees fetched successfully", dto.ToFeesResponseListDTO(fees))
}

// CreatePayment processes a student fee payment
func (cl *FeesController) CreatePayment(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse request JSON body
	var req dto.CreatePaymentDTO
	if err := c.Bind().Body(&req); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	// 3. Sanitize and validate payment request data
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Ensure students can only make payments for their own profile
	loggined_student_id, _ := cl.studentService.GetUserStudentIDService(userID)
	if req.StudentID != loggined_student_id {
		return helper.Error(c, 403, "Cant able to make payment for other student")
	}

	// 5. Process payment transaction via service
	payment, err := cl.feesService.CreatePayment(userID, &req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") || strings.Contains(strings.ToLower(err.Error()), "cant pay fees for other student") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 6. Return payment transaction receipt
	return helper.Success(c, "Payment created successfully", payment)
}

// GetPaymentByIDController retrieves payment transaction receipt by ID
func (cl *FeesController) GetPaymentByIDController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse payment ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid payment id")
	}

	// 3. Fetch payment receipt via service
	payment, err := cl.feesService.GetPaymentByID(userID, uint(id))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	// 4. Return receipt response
	return helper.Success(c, "Payment fetched successfully", payment)
}

// GetPaymentByFeeIDController fetches all payments made towards a specific fee record
func (cl *FeesController) GetPaymentByFeeIDController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse fee ID parameter
	feeIDStr := c.Params("fee_id")
	feeID, err := strconv.ParseUint(feeIDStr, 10, 32)
	if err != nil || feeID == 0 {
		return helper.Error(c, 400, "invalid fee id")
	}

	// 3. Fetch payment history from service
	payments, err := cl.feesService.GetPaymentByFeeID(userID, uint(feeID))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	// 4. Return payment list
	return helper.Success(c, "Payments fetched successfully", payments)
}
