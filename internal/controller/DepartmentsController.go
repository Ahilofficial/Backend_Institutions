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

type DepartmentController struct {
	departmentService *services.DepartmentService
	instituteService *services.InstituteService
	facultyService *services.FacultyService
}

func NewDepartmentController(departmentService *services.DepartmentService,instituteService *services.InstituteService , facultyService *services.FacultyService) *DepartmentController {
	return &DepartmentController{
		departmentService: departmentService,
		instituteService: instituteService,
		facultyService: facultyService,
	}
}

func (cl *DepartmentController) GetActiveDepartmentController(c fiber.Ctx) error {
	department, err := cl.departmentService.GetActiveDepartmentService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Active department fetched successfully",
		dto.ToDepartmentResponseDTO(&department),
	)
}

func (cl *DepartmentController) GetInactiveDepartmentController(c fiber.Ctx) error {
	department, err := cl.departmentService.GetInactiveDepartmentService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Inactive department fetched successfully",
		dto.ToDepartmentResponseDTO(&department),
	)
}

func (cl *DepartmentController) CreateDepartmentController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	var body dto.CreateDepartmentDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_user_institution_id := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && (checking_user_institution_id == 0 || checking_user_institution_id != body.InstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	department := model.Department{
		DepartmentName: body.DepartmentName,
		InstitutionID:  body.InstitutionID,
		IsActive:       true,
	}

	createdDept, err := cl.departmentService.AddDepartmentService(
		userID,
		&department,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access") || strings.Contains(strings.ToLower(err.Error()), "denied") {
			return helper.Error(c, 403, "Cant able to access other institution")
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Department created successfully",
		dto.ToDepartmentResponseDTO(&createdDept),
	)
}

func (cl *DepartmentController) GetAllDepartmentsController(c fiber.Ctx) error {
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

	departments, total, err := cl.departmentService.GetDepartmentServicePaginated(search, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

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

func (cl *DepartmentController) GetDepartmentByIDController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)

	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid department id")
	}

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.departmentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	department, err := cl.departmentService.GetDepartmentByIDService(
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
		"Department fetched successfully",
		dto.ToDepartmentResponseDTO(&department),
	)
}

func (cl *DepartmentController) GetDeletedDepartmentsController(c fiber.Ctx) error {
	departments, err := cl.departmentService.GetDepartmentServiceDeleted()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"Deleted departments fetched successfully",
		dto.ToDepartmentResponseListDTO(departments),
	)
}

func (cl *DepartmentController) UpdateDepartmentController(c fiber.Ctx) error {


	userID, ok := c.Locals("user_id").(uint)


	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid department id")
	}
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.departmentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	var body dto.UpdateDepartmentDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.departmentService.UpdateDepartmentService(
		userID,
		uint(id),
		&body,
	); err != nil {

		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	updated, err := cl.departmentService.GetDepartmentByIDService(
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
		"Department updated successfully",
		dto.ToDepartmentResponseDTO(&updated),
	)
}
func (cl *DepartmentController) DeleteDepartmentController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid department id")
	}
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	checking_user_institution_id := cl.departmentService.GetInstitutionIDForUserService(uint(id))
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_user_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	if err := cl.departmentService.DeleteDepartment(
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
		"Department deleted successfully",
		nil,
	)
}

func (cl *DepartmentController) FetchAllDepartmentsController(c fiber.Ctx) error {
	departments, err := cl.departmentService.GetDepartmentService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"All departments fetched successfully",
		dto.ToDepartmentResponseListDTO(departments),
	)
}
