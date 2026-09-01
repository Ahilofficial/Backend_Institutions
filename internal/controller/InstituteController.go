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

type InstituteController struct {
	instituteService *services.InstituteService
}

func NewInstituteController(instituteService *services.InstituteService) *InstituteController {
	return &InstituteController{instituteService: instituteService}
}

func (cl *InstituteController) CreateInstituteController(c fiber.Ctx) error {
	var body dto.CreateInstitutionDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	institute := model.Institutions{
		Name:            body.Name,
		InstitutionCode: body.InstitutionCode,
		State:           body.State,
		IsActive:        true,
	}

	createdInstitute, err := cl.instituteService.CreateInsituteService(&institute)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Institution created successfully",
		dto.ToInstitutionResponseDTO(&createdInstitute),
	)
}

func (cl *InstituteController) GetAllInstitutesController(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)

	
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

	institutes, total, err := cl.instituteService.GetInstituteServicePaginated(userID, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return helper.Success(
		c,
		"Institutes fetched successfully",
		fiber.Map{
			"items":       dto.ToInstitutionResponseListDTO(institutes),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}

func (cl *InstituteController) GetInstituteByIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idstr := c.Params("id")
	id, err := strconv.ParseUint(idstr, 10, 32)

	if err != nil {
		return helper.Error(c, 400, "Invalid institute ID")
	}
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_user_institution_id := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && ( checking_user_institution_id != uint(id)) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}
	institute, err := cl.instituteService.GetInstituteServiceById(userID, uint(id))
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Institute fetched successfully",
		dto.ToInstitutionResponseDTO(&institute),
	)
}



func (cl *InstituteController) UpdateInstituteController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid id")
	}
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_user_institution_id := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && ( checking_user_institution_id != uint(id)) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}
	var body dto.UpdateInstitutionDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.instituteService.UpdateInstitutionService(
		userID,
		uint(id),
		&body,
	); err != nil {
			return helper.Error(c, 403, err.Error())

	}

	updated, err := cl.instituteService.GetInstituteServiceById(userID, uint(id))
	if err != nil {
			return helper.Error(c, 403, err.Error())
	}

	return helper.Success(
		c,
		"Institution updated successfully",
		dto.ToInstitutionResponseDTO(&updated),
	)
}

func (cl *InstituteController) DeleteInstituteController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_user_institution_id := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && (checking_user_institution_id != uint(id)) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}
	if err != nil {
		return helper.Error(c, 400, "invalid institute id")
	}

	if err := cl.instituteService.DeleteInstitutionService( uint(id)); err != nil {
		
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Institution deleted successfully",
		nil,
	)
}
