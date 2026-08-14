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

type FacultyController struct {
	facultyService *services.FacultyService
}

func NewFacultyController(facultyService *services.FacultyService) *FacultyController {
	return &FacultyController{facultyService: facultyService}
}

func (cl *FacultyController) CreateFacultyController(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)

	var faculty model.Faculty

	if err := c.Bind().Body(&faculty); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if faculty.Name == "" {
		return helper.Error(c, 400, "name is required")
	}

	if faculty.DepartmentID == 0 {
		return helper.Error(c, 400, "department_id is required")
	}

	createdFaculty, err := cl.facultyService.CreateFacultyService(
		userID,
		&faculty,
	)
	if err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Faculty created successfully",
		dto.ToFacultyResponseDTO(&createdFaculty),
	)
}

func (cl *FacultyController) GetAllFacultiesController(c fiber.Ctx) error {
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

	faculties, total, err := cl.facultyService.GetFacultyServicePaginated(search, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

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

func (cl *FacultyController) GetFacultyByIDController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid faculty ID")
	}

	faculty, err := cl.facultyService.GetFacultyServiceById(
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
		"Faculty fetched successfully",
		dto.ToFacultyResponseDTO(faculty),
	)
}

func (cl *FacultyController) GetDeletedFacultiesController(c fiber.Ctx) error {
	faculties, err := cl.facultyService.GetFacultyServiceDeleted()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"Deleted faculties fetched successfully",
		dto.ToFacultyResponseListDTO(faculties),
	)
}

func (cl *FacultyController) GetActiveFacultyController(c fiber.Ctx) error {
	faculty, err := cl.facultyService.GetActiveFacultyService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Active faculty fetched successfully",
		dto.ToFacultyResponseDTO(&faculty),
	)
}

func (cl *FacultyController) GetInactiveFacultyController(c fiber.Ctx) error {
	faculty, err := cl.facultyService.GetInactiveFacultyService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Inactive faculty fetched successfully",
		dto.ToFacultyResponseDTO(&faculty),
	)
}

func (cl *FacultyController) UpdateFacultyController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid id")
	}

	var body dto.UpdateFacultyDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.facultyService.UpdateFacultyService(
		userID,
		uint(id),
		&body,
	); err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	updated, err := cl.facultyService.GetFacultyServiceById(
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
		"Faculty updated successfully",
		dto.ToFacultyResponseDTO(updated),
	)
}

func (cl *FacultyController) DeleteFacultyController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid faculty id")
	}

	if err := cl.facultyService.DeleteFacultyService(
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
		"Faculty deleted successfully",
		nil,
	)
}

func (cl *FacultyController) FetchAllFacultiesController(c fiber.Ctx) error {
	faculties, err := cl.facultyService.GetFacultyService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"All faculties fetched successfully",
		dto.ToFacultyResponseListDTO(faculties),
	)
}
