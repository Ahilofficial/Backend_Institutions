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

type PrincipalControllers struct {
	principalService *services.PrincipalService
}

func NewPrincipalControllers(principalService *services.PrincipalService) *PrincipalControllers {
	return &PrincipalControllers{principalService: principalService}
}

func (cl *PrincipalControllers) CreatePrincipalController(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)

	var principal model.Principal

	if err := c.Bind().Body(&principal); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	if principal.Name == "" {
		return helper.Error(c, 400, "name is required")
	}

	if principal.InstitutionID == 0 {
		return helper.Error(c, 400, "institution_id is required")
	}

	createdPrincipal, err := cl.principalService.CreatePrincipalService(
		userID,
		&principal,
	)
	if err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Principal created successfully",
		dto.ToPrincipalResponseDTO(&createdPrincipal),
	)
}

func (cl *PrincipalControllers) GetAllPrincipalsController(c fiber.Ctx) error {
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

	principals, total, err := cl.principalService.GetPrincipalServicePaginated(search, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return helper.Success(
		c,
		"Principals fetched successfully",
		fiber.Map{
			"items":       dto.ToPrincipalResponseListDTO(principals),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}

func (cl *PrincipalControllers) GetPrincipalByIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid principal ID")
	}

	principal, err := cl.principalService.GetPrincipalServiceById(
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
		"Principal fetched successfully",
		dto.ToPrincipalResponseDTO(principal),
	)
}

func (cl *PrincipalControllers) GetDeletedPrincipalsController(c fiber.Ctx) error {
	principals, err := cl.principalService.GetPrincipalServiceDeleted()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"Deleted principals fetched successfully",
		dto.ToPrincipalResponseListDTO(principals),
	)
}

func (cl *PrincipalControllers) GetActivePrincipalController(c fiber.Ctx) error {
	principal, err := cl.principalService.GetActivePrincipalService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Active principal fetched successfully",
		dto.ToPrincipalResponseDTO(&principal),
	)
}

func (cl *PrincipalControllers) GetInactivePrincipalController(c fiber.Ctx) error {
	principal, err := cl.principalService.GetInactivePrincipalService()
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Inactive principal fetched successfully",
		dto.ToPrincipalResponseDTO(&principal),
	)
}

func (cl *PrincipalControllers) UpdatePrincipalController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid id")
	}

	var body dto.UpdatePrincipalDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if err := cl.principalService.UpdatePrincipalService(
		userID,
		uint(id),
		&body,
	); err != nil {
		if err.Error() == "access denied" {
			return helper.Error(c, 403, "Access denied")
		}

		return helper.Error(c, 400, err.Error())
	}

	updated, err := cl.principalService.GetPrincipalServiceById(
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
		"Principal updated successfully",
		dto.ToPrincipalResponseDTO(updated),
	)
}

func (cl *PrincipalControllers) DeletePrincipalController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid principal id")
	}

	if err := cl.principalService.DeletePrincipalService(
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
		"Principal deleted successfully",
		nil,
	)
}

func (cl *PrincipalControllers) FetchAllPrincipalsController(c fiber.Ctx) error {
	principals, err := cl.principalService.GetPrincipalService()
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"All principals fetched successfully",
		dto.ToPrincipalResponseListDTO(principals),
	)
}
