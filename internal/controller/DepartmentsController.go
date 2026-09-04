package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/model"
	"backend_institutions/internal/services"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// DepartmentController handles HTTP requests for department operations
type DepartmentController struct {
	departmentService *services.DepartmentService
	instituteService  *services.InstituteService
	facultyService    *services.FacultyService
}

// NewDepartmentController initializes a new DepartmentController instance
func NewDepartmentController(departmentService *services.DepartmentService, instituteService *services.InstituteService, facultyService *services.FacultyService) *DepartmentController {
	return &DepartmentController{
		departmentService: departmentService,
		instituteService:  instituteService,
		facultyService:    facultyService,
	}
}

// CreateDepartmentController creates a new academic department within an institution
func (cl *DepartmentController) CreateDepartmentController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Bind and validate request body
	var body dto.CreateDepartmentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Construct department model
	department := model.Department{
		DepartmentName: body.DepartmentName,
		InstitutionID:  body.InstitutionID,
		IsActive:       true,
	}

	// 4. Persist department via service (service validates institution access)
	createdDept, err := cl.departmentService.AddDepartmentService(
		userID,
		&department,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return response with department DTO
	return helper.Success(
		c,
		"Department created successfully",
		dto.ToDepartmentResponseDTO(&createdDept),
	)
}

// GetAllDepartmentsController retrieves paginated list of departments
func (cl *DepartmentController) GetAllDepartmentsController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, _ := c.Locals("user_id").(uint)

	// 2. Parse pagination query parameters
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err := strconv.ParseUint(pageStr, 10, 64)
	limit, err := strconv.ParseUint(limitStr, 10, 64)

	// 3. Query paginated departments (scoped to institution admin if applicable)
	departments, total, err := cl.departmentService.GetDepartmentServicePaginated(userID, int(page), int(limit))
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 4. Calculate total page count
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return paginated response
	return helper.Success(
		c,
		"Departments fetched successfully",
		fiber.Map{
			"items":       dto.ToDepartmentResponseListDTO(departments),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}

// GetDepartmentByIDController retrieves a single department with access verification
func (cl *DepartmentController) GetDepartmentByIDController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse department ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid department id")
	}

	// 3. Fetch department from service (service validates access boundaries)
	department, err := cl.departmentService.GetDepartmentByIDService(
		userID,
		uint(id),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 404, err.Error())
	}

	// 4. Return success response with department DTO
	return helper.Success(
		c,
		"Department fetched successfully",
		dto.ToDepartmentResponseDTO(&department),
	)
}

// UpdateDepartmentController updates department details
func (cl *DepartmentController) UpdateDepartmentController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse department ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid department id")
	}

	// 3. Bind and validate request body
	var body dto.UpdateDepartmentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update department in service (service validates access boundaries)
	if err := cl.departmentService.UpdateDepartmentService(
		userID,
		uint(id),
		&body,
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 5. Fetch updated record
	updated, err := cl.departmentService.GetDepartmentByIDService(
		userID,
		uint(id),
	)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 6. Return success response
	return helper.Success(
		c,
		"Department updated successfully",
		dto.ToDepartmentResponseDTO(&updated),
	)
}

// DeleteDepartmentController deletes (soft deletes) a department
func (cl *DepartmentController) DeleteDepartmentController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse department ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid department id")
	}

	// 3. Delete department via service (service validates access boundaries)
	if err := cl.departmentService.DeleteDepartment(
		userID,
		uint(id),
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success response
	return helper.Success(
		c,
		"Department deleted successfully",
		nil,
	)
}
