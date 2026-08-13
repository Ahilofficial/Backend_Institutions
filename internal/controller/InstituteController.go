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
	var institute model.Institutions
	if err := c.Bind().Body(&institute); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if institute.Name == "" {
		return helper.Error(c, 400, "name is required")
	}
	if institute.InstitutionCode == "" {
		return helper.Error(c, 400, "institution_code is required")
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

	institutes, total, err := cl.instituteService.GetInstituteServicePaginated(search, page, limit)
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

	institute, err := cl.instituteService.GetInstituteServiceById(userID, uint(id))
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Institute fetched successfully",
		dto.ToInstitutionResponseDTO(&institute),
	)
}

func (cl *InstituteController) GetDeletedInstitutesController(c fiber.Ctx) error {
	institutes, err := cl.instituteService.GetInstituteServiceDeleted()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"Deleted institutes fetched successfully",
		dto.ToInstitutionResponseListDTO(institutes),
	)
}

func (cl *InstituteController) GetActiveInstituteController(c fiber.Ctx) error {
	institute, err := cl.instituteService.GetActiveInstitute()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Active institution fetched successfully",
		dto.ToInstitutionResponseDTO(&institute),
	)
}

func (cl *InstituteController) GetInactiveInstituteController(c fiber.Ctx) error {
	institute, err := cl.instituteService.GetInactiveInstitute()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Inactive institution fetched successfully",
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
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 400, err.Error())
	}

	updated, err := cl.instituteService.GetInstituteServiceById(userID, uint(id))
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 500, err.Error())
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
	if err != nil {
		return helper.Error(c, 400, "invalid institute id")
	}

	if err := cl.instituteService.DeleteInstitutionService(userID, uint(id)); err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Institution deleted successfully",
		nil,
	)
}

func (cl *InstituteController) FetchAllInstitutesController(c fiber.Ctx) error {
	institutes, err := cl.instituteService.GetInstituteService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"All institutes fetched successfully",
		dto.ToInstitutionResponseListDTO(institutes),
	)
}
