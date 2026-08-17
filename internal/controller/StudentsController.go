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
	studentService *services.StudentService
}

func NewStudentController(studentService *services.StudentService) *StudentController {
	return &StudentController{studentService: studentService}
}

func (cl *StudentController) CreateStudentControllers(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)

	var student model.Student

	if err := c.Bind().Body(&student); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if student.Name == "" {
		return helper.Error(c, 400, "name is required")
	}

	userFacultyID, _ := cl.studentService.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		if student.FacultyID == 0 {
			student.FacultyID = userFacultyID
		}
		student.UserID = 0
	} else {
		userExistingType, _ := cl.studentService.GetUserExistingProfile(userID)
		if userExistingType == "student" || userExistingType == "" {
			student.UserID = userID
		} else {
			student.UserID = 0
		}
	}

	if student.FacultyID == 0 {
		return helper.Error(c, 400, "faculty_id is required")
	}

	createdStudent, err := cl.studentService.CreateStudentService(
		userID,
		&student,
	)
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(c, "Student created successfully", dto.ToStudentResponseDTO(createdStudent))
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

func (cl *StudentController) GetInactiveStudentController(c fiber.Ctx) error {
	student, err := cl.studentService.GetInactiveStudentService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Inactive student fetched successfully",
		dto.ToStudentResponseDTO(&student),
	)
}

func (cl *StudentController) GetStudentByIDControllers(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid student id")
	}

	student, err := cl.studentService.GetStudentServiceById(
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
		"Student fetched successfully",
		dto.ToStudentResponseDTO(student),
	)
}

func (cl *StudentController) UpdateStudentControllers(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid student id")
	}

	var body dto.UpdateStudentDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.studentService.UpdateStudentService(
		userID,
		uint(id),
		&body,
	); err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	updated, err := cl.studentService.GetStudentServiceById(
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
		"Student updated successfully",
		dto.ToStudentResponseDTO(updated),
	)
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

	if err := cl.studentService.DeleteStudentService(
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
		"Student deleted successfully",
		nil,
	)
}

func (cl *StudentController) FetchAllStudentsControllers(c fiber.Ctx) error {
	students, err := cl.studentService.FetchAllStudentsServices()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"All students fetched successfully",
		dto.ToStudentResponseListDTO(students),
	)
}

func (cl *StudentController) FetchAllStudentsPaginatedControllers(c fiber.Ctx) error {
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

	students, total, err := cl.studentService.FetchAllStudentsPaginatedServices(search, page, limit)
	if err != nil {
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

func (c *StudentController) FetchStudentsByPaymentMonth(ctx fiber.Ctx) error {
	userID, _ := ctx.Locals("user_id").(uint)
	month := strings.TrimSpace(ctx.Query("month"))

	if month == "" {
		return helper.Error(ctx, fiber.StatusBadRequest, "month query parameter is required")
	}

	students, err := c.studentService.FetchStudentsByPaymentMonthService(userID, month)
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(ctx, "Paid students for month fetched successfully", dto.ToStudentResponseListDTO(students))
}

func (c *StudentController) FetchStudentsNotPaidByMonth(ctx fiber.Ctx) error {
	userID, _ := ctx.Locals("user_id").(uint)
	month := strings.TrimSpace(ctx.Query("month"))

	if month == "" {
		return helper.Error(ctx, fiber.StatusBadRequest, "month query parameter is required")
	}

	students, err := c.studentService.FetchStudentsNotPaidByMonthService(userID, month)
	if err != nil {
		return helper.Error(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return helper.Success(ctx, "Not paid students for month fetched successfully", dto.ToStudentResponseListDTO(students))
}

func (c *StudentController) FetchFacultyPaidStudents(ctx fiber.Ctx) error {
	userID, _ := ctx.Locals("user_id").(uint)
	month := strings.TrimSpace(ctx.Query("month"))
	facultyID, _ := strconv.ParseUint(ctx.Query("faculty_id"), 10, 32)

	students, err := c.studentService.FetchFacultyPaidStudentsService(userID, uint(facultyID), month)
	if err != nil {
		return helper.Error(ctx, fiber.StatusBadRequest, err.Error())
	}

	return helper.Success(ctx, "Faculty paid students fetched successfully", dto.ToStudentResponseListDTO(students))
}

func (c *StudentController) FetchFacultyUnpaidStudents(ctx fiber.Ctx) error {
	userID, _ := ctx.Locals("user_id").(uint)
	month := strings.TrimSpace(ctx.Query("month"))
	facultyID, _ := strconv.ParseUint(ctx.Query("faculty_id"), 10, 32)

	students, err := c.studentService.FetchFacultyUnpaidStudentsService(userID, uint(facultyID), month)
	if err != nil {
		return helper.Error(ctx, fiber.StatusBadRequest, err.Error())
	}

	return helper.Success(ctx, "Faculty unpaid students fetched successfully", dto.ToStudentResponseListDTO(students))
}

func (c *StudentController) GetLoggedInStudentController(ctx fiber.Ctx) error {
	userID, _ := ctx.Locals("user_id").(uint)

	student, err := c.studentService.GetLoggedInStudentProfile(userID)
	if err != nil {
		return helper.Error(ctx, fiber.StatusNotFound, err.Error())
	}

	return helper.Success(ctx, "LoggedIn student profile fetched successfully", dto.ToStudentResponseDTO(student))
}
