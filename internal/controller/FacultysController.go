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

// FacultyController handles HTTP requests for faculty management
type FacultyController struct {
	facultyService   *services.FacultyService
	userService      *services.UserService
	instituteService *services.InstituteService
}

// NewFacultyController instantiates a new FacultyController with required services
func NewFacultyController(
	facultyService *services.FacultyService,
	userService *services.UserService,
	instituteService *services.InstituteService,
) *FacultyController {
	return &FacultyController{
		facultyService:   facultyService,
		userService:      userService,
		instituteService: instituteService,
	}
}

// GetFacultyByIDController handles fetching a single faculty by ID with RBAC and access checks
func (cl *FacultyController) GetFacultyByIDController(c fiber.Ctx) error {
	// 1. Extract and validate authenticated user ID from context
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse and validate path parameter ID
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid faculty ID")
	}
	facultyID := uint(id)

	// 3. Fetch faculty details via service (service validates access boundaries)
	faculty, err := cl.facultyService.GetFacultyServiceById(
		userID,
		facultyID,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") || strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 404, err.Error())
	}

	// 4. Return response DTO
	return helper.Success(
		c,
		"Faculty fetched successfully",
		dto.ToFacultyResponseDTO(faculty),
	)
}

// CreateFacultyController handles registration of a new faculty profile
func (cl *FacultyController) CreateFacultyController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID from context
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	// 2. Bind request body JSON
	var body dto.CreateFacultyDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(
			c,
			400,
			"invalid request body: "+err.Error(),
		)
	}

	// 3. Sanitize and validate input fields
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Delegate profile creation to service
	createdFaculty, err := cl.facultyService.CreateFacultyService(userID, &body)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success response with created entity
	return helper.Success(
		c,
		"Faculty created successfully",
		dto.ToFacultyResponseDTO(&createdFaculty),
	)
}

// GetAllFacultiesController handles paginated retrieval of all faculty records
func (cl *FacultyController) GetAllFacultiesController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, _ := c.Locals("user_id").(uint)

	// 2. Parse pagination query parameters
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 3. Fetch paginated faculty data from service (scoped to user's permissions)
	faculties, total, err := cl.facultyService.GetFacultyServicePaginated(userID, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 4. Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return formatted response map
	return helper.Success(
		c,
		"Faculties fetched successfully",
		fiber.Map{
			"items":       dto.ToFacultyResponseListDTO(faculties),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}

// GetLoggedInFacultyController fetches the faculty profile of the authenticated user
func (cl *FacultyController) GetLoggedInFacultyController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, _ := c.Locals("user_id").(uint)

	// 2. Fetch profile from service
	faculty, err := cl.facultyService.GetLoggedInFacultyProfile(userID)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	// 3. Return response
	return helper.Success(
		c,
		"LoggedIn faculty profile fetched successfully",
		dto.ToFacultyResponseDTO(faculty),
	)
}

// GetLoggedInFacultyStudentsController fetches all students assigned to the authenticated faculty
func (cl *FacultyController) GetLoggedInFacultyStudentsController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID := c.Locals("user_id").(uint)

	// 2. Fetch students from service
	students, err := cl.facultyService.GetLoggedInFacultyStudents(userID)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	// 3. Return list response
	return helper.Success(
		c,
		"LoggedIn faculty students fetched successfully",
		dto.ToStudentResponseListDTO(students),
	)
}

// UpdateFacultyController handles updating faculty profile details
func (cl *FacultyController) UpdateFacultyController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse path parameter ID
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid faculty ID")
	}
	facultyID := uint(id)

	// 3. Parse and validate update request body
	var body dto.UpdateFacultyDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update faculty profile via service (service validates access boundaries)
	err = cl.facultyService.UpdateFacultyService(
		userID,
		facultyID,
		&body,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") || strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success response
	return helper.Success(
		c,
		"Faculty updated successfully",
		nil,
	)
}

// GetPaidStudentsForFacultyController fetches all fully paid students assigned to the logged-in faculty
func (cl *FacultyController) GetPaidStudentsForFacultyController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	user_id := c.Locals("user_id").(uint)

	// 2. Fetch paid students list via service
	paid_students, err := cl.facultyService.GetPaidStudentsForFacultyService(user_id)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Return response DTO list
	return helper.Success(c, "Paid students fetched successfully", dto.ToStudentResponseListDTO(paid_students))
}

// GetNonPaidStudentsForFacultyController fetches non-paid students assigned to the logged-in faculty
func (cl *FacultyController) GetNonPaidStudentsForFacultyController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	user_id := c.Locals("user_id").(uint)

	// 2. Fetch pending students list via service
	non_paid_students, err := cl.facultyService.GetNonPaidStudentsForFacultyService(user_id)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 3. Return response DTO list
	return helper.Success(c, "Non-paid students fetched successfully", dto.ToStudentResponseListDTO(non_paid_students))
}

// DeleteFacultyController handles deletion of faculty profiles
func (cl *FacultyController) DeleteFacultyController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse path parameter ID
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid faculty id")
	}
	facultyID := uint(id)

	// 3. Delete faculty record via service (service validates access boundaries)
	if err := cl.facultyService.DeleteFacultyService(
		userID,
		facultyID,
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") || strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success confirmation
	return helper.Success(
		c,
		"Faculty deleted successfully",
		nil,
	)
}
