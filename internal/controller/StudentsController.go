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

type StudentController struct {
	studentService   *services.StudentService
	userService      *services.UserService
	instituteService *services.InstituteService
	facultyService   *services.FacultyService
}

func NewStudentController(studentService *services.StudentService, userService *services.UserService, instituteService *services.InstituteService, facultyService *services.FacultyService) *StudentController {
	return &StudentController{
		studentService:   studentService,
		userService:      userService,
		instituteService: instituteService,
		facultyService:   facultyService,
	}
}

func (cl *StudentController) CreateStudentControllers(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	var body dto.CreateStudentDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()
	course_duration, _ := cl.studentService.GetCourseDurationByFacultyID(body.FacultyID)
	if course_duration > 0 && body.Semester > course_duration*2 {
		return helper.Error(c, 400, "This particular semester does not contain for the particular department")
	}

	user, err := cl.userService.GetUserByID(userID)
	if err != nil || user == nil {
		return helper.Error(c, 404, "user not found")
	}

	if user.FacultyID != 0 && body.FacultyID == 0 {
		body.FacultyID = user.FacultyID
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	var targetUserID uint
	if user.StudentID == 0 {
		targetUserID = userID
	}

	student := model.Student{
		Name:        body.Name,
		Gender:      body.Gender,
		FacultyID:   body.FacultyID,
		UserID:      targetUserID,
		Hosteller:   body.Hosteller,
		Scholarship: body.Scholorship,
		MQ:          body.MQ,
		IsActive:    true,
	}

	createdStudent, err := cl.studentService.CreateStudentService(
		userID,
		&student,
		&body,
	)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if targetUserID != 0 {
		_ = cl.userService.UpdateStudentID(targetUserID, createdStudent.ID)
	}

	return helper.Success(
		c,
		"Student created successfully",
		dto.ToStudentResponseDTO(createdStudent),
	)
}
func (cl *StudentController) GetActiveStudentController(c fiber.Ctx) error {
	student, err := cl.studentService.GetActiveStudentService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Active student fetched successfully",
		dto.ToStudentResponseDTO(&student),
	)
}



func (cl *StudentController) GetStudentByIDControllers(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid student ID")
	}

	studentID := uint(id)

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(studentID)
	
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}
	student, err := cl.studentService.GetStudentServiceById(
		userID,
		studentID,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 404, err.Error())
	}
	if student == nil {
		return helper.Error(c, 404, "Student not found")
	}

	if !student.IsProfileVerified {
		return helper.Error(c, 400, "Faculty first need to verify your profile")
	}

	FacultyID, _ := cl.studentService.GetUserFacultyID(userID)
	if FacultyID > 0 {
		_, _ = cl.studentService.StudentVerification(studentID, FacultyID)
	}

	return helper.Success(
		c,
		"Student fetched successfully",
		dto.ToStudentResponseDTO(student),
	)
}

func (cl *StudentController) VerifyStudentController(
	c fiber.Ctx,
) error {

	userID, ok := c.Locals("user_id").(uint)

	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)

	if err != nil {
		return helper.Error(c, 400, "invalid student ID")
	}

	studentID := uint(id)

	err = cl.studentService.UpdateStudentVerified(
		userID,
		studentID,
	)

	if err != nil {
		return helper.Error(
			c,
			fiber.StatusForbidden,
			err.Error(),
		)
	}

	return helper.Success(
		c,
		"Student verified successfully",
		nil,
	)
}

func (cl *StudentController) UpdateStudentController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "Invalid student ID")
	}

	var body dto.UpdateStudentDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	if err := cl.studentService.UpdateStudentService(
		userID,
		uint(id),
		&body,
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	student, err := cl.studentService.GetStudentServiceById(userID, uint(id))
	if err != nil || student == nil {
		return helper.Success(
			c,
			"Student updated successfully",
			nil,
		)
	}

	return helper.Success(
		c,
		"Student updated successfully",
		dto.ToStudentResponseDTO(student),
	)
}

func (cl *StudentController) UpdateStudentSemesterController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, 400, "Invalid student ID")
	}

	var body dto.UpdateStudentSemesterDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	student, fee, err := cl.studentService.UpdateStudentSemesterService(userID, uint(id), &body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}
	

	return helper.Success(c, "Student semester details updated successfully", fiber.Map{
		"student": dto.ToStudentResponseDTO(student),
		"fee":     dto.ToFeesResponseDTO(fee),
	})
}

func (cl *StudentController) DeleteStudentControllers(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid student id")
	}

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.studentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	if err := cl.studentService.DeleteStudentService(
		userID,
		uint(id),
	); err != nil {

		if strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, err.Error())
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Student deleted successfully",
		nil,
	)
}



func (cl *StudentController) FetchAllStudentsPaginatedControllers(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

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

	students, total, err := cl.studentService.FetchAllStudentsPaginatedServicesScoped(userID, search, page, limit)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "cant able to access") || strings.Contains(strings.ToLower(err.Error()), "access denied") {
			return helper.Error(c, 403, "Cant able to access other institution")
		}
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

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

func (c *StudentController) GetLoggedInStudentController(ctx fiber.Ctx) error {
	userID, _ := ctx.Locals("user_id").(uint)

	student, err := c.studentService.GetLoggedInStudentProfile(userID)
	if err != nil {
		return helper.Error(ctx, fiber.StatusNotFound, err.Error())
	}

	return helper.Success(ctx, "LoggedIn student profile fetched successfully", dto.ToStudentResponseDTO(student))
}
