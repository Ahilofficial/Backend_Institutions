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

// StudentController handles HTTP endpoints for student registration, updates, and queries
type StudentController struct {
	studentService   *services.StudentService
	userService      *services.UserService
	instituteService *services.InstituteService
	facultyService   *services.FacultyService
}

// NewStudentController instantiates a new StudentController with required dependencies
func NewStudentController(studentService *services.StudentService, userService *services.UserService, instituteService *services.InstituteService, facultyService *services.FacultyService) *StudentController {
	return &StudentController{
		studentService:   studentService,
		userService:      userService,
		instituteService: instituteService,
		facultyService:   facultyService,
	}
}

// CreateStudentControllers handles creation and fee initialization for a student
func (cl *StudentController) CreateStudentControllers(c fiber.Ctx) error {
	// 1. Extract authenticated user ID from context
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	// 2. Parse request JSON payload
	var body dto.CreateStudentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 3. Sanitize and validate request data
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Delegate student creation and fee calculation to service
	createdStudent, err := cl.studentService.CreateStudentService(userID, &body)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success response with created student
	return helper.Success(
		c,
		"Student created successfully",
		dto.ToStudentResponseDTO(createdStudent),
	)
}

// GetStudentByIDControllers handles single student lookup with authorization checks
func (cl *StudentController) GetStudentByIDControllers(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse path parameter ID
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "invalid student ID")
	}
	studentID := uint(id)

	// 3. Check student access: students can only access own profile
	logginedUserStudentID, _ := cl.studentService.GetUserStudentIDService(userID)
	if logginedUserStudentID != studentID {
		return helper.Error(c, 403, "Cant able to access other student")
	}

	// 4. Institution admin verification: check student belongs to admin's institution
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(studentID)
	if is_inst_admin && (checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	// 5. Fetch student details from service
	student, err := cl.studentService.GetStudentServiceById(
		userID,
		studentID,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}
	if student == nil {
		return helper.Error(c, 404, "Student not found")
	}

	// 6. Return response
	return helper.Success(
		c,
		"Student fetched successfully",
		dto.ToStudentResponseDTO(student),
	)
}

// UpdateStudentController handles updating student details (name, gender)
func (cl *StudentController) UpdateStudentController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse path parameter ID
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid student ID")
	}

	// 3. Parse and validate update request body
	var body dto.UpdateStudentDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Check institution admin access scope
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	// 5. Update student details via service
	student, err := cl.studentService.UpdateStudentService(
		userID,
		uint(id),
		&body,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 6. Return updated student response
	return helper.Success(
		c,
		"Student updated successfully",
		dto.ToStudentResponseDTO(student),
	)
}

// UpdateStudentSemesterController handles updating student semester and recalculating fee quotas
func (cl *StudentController) UpdateStudentSemesterController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse path parameter ID
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid student ID")
	}

	// 3. Parse and validate semester update body
	var body dto.UpdateStudentSemesterDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Check institution admin scope
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	// 5. Perform update and fee adjustment via service
	student, fee, err := cl.studentService.UpdateStudentSemesterService(userID, uint(id), &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 6. Return updated details and fee snapshot
	return helper.Success(c, "Student semester details updated successfully", fiber.Map{
		"student": dto.ToStudentResponseDTO(student),
		"fee":     dto.ToFeesResponseDTO(fee),
	})
}

// DeleteStudentControllers handles soft deletion of a student profile
func (cl *StudentController) DeleteStudentControllers(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse path parameter ID
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid student id")
	}

	// 3. Institution admin scoping check
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	// 4. Delete student record via service
	if err := cl.studentService.DeleteStudentService(
		userID,
		uint(id),
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success confirmation
	return helper.Success(
		c,
		"Student deleted successfully",
		nil,
	)
}

// FetchAllStudentsPaginatedControllers handles paginated student listing filtered by user's institution
func (cl *StudentController) FetchAllStudentsPaginatedControllers(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse query parameters
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

	// 3. Fetch scoped paginated students from service
	students, total, err := cl.studentService.FetchAllStudentsPaginatedServicesScoped(userID, search, page, limit)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cant able to access") || strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, "Cant able to access other institution")
		}
		return helper.Error(c, 500, err.Error())
	}

	// 4. Calculate total page count
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return paginated result map
	return helper.Success(
		c,
		"Students fetched successfully",
		fiber.Map{
			"items":       dto.ToStudentResponseListDTO(students),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}

// GetLoggedInStudentController retrieves the student profile of the authenticated user
func (c *StudentController) GetLoggedInStudentController(ctx fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, _ := ctx.Locals("user_id").(uint)

	// 2. Fetch logged in student profile from service
	student, err := c.studentService.GetLoggedInStudentProfile(userID)
	if err != nil {
		return helper.Error(ctx, fiber.StatusNotFound, err.Error())
	}

	// 3. Return response DTO
	return helper.Success(ctx, "LoggedIn student profile fetched successfully", dto.ToStudentResponseDTO(student))
}
