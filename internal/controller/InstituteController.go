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

// InstituteController handles HTTP requests for institution management
type InstituteController struct {
	instituteService *services.InstituteService
}

// NewInstituteController initializes a new InstituteController
func NewInstituteController(instituteService *services.InstituteService) *InstituteController {
	return &InstituteController{instituteService: instituteService}
}

// CreateInstituteController handles creating a new institution
func (cl *InstituteController) CreateInstituteController(c fiber.Ctx) error {
	// 1. Bin
	// d incoming JSON request body
	var body dto.CreateInstitutionDTO
	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body: "+err.Error())
	}

	// 2. Sanitize and validate input payload
	body.Sanitize()
	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 3. Construct institution model
	institute := model.Institutions{
		Name:            body.Name,
		InstitutionCode: body.InstitutionCode,
		State:           body.State,
		IsActive:        true,
	}

	// 4. Call service to persist institution
	createdInstitute, err := cl.instituteService.CreateInsituteService(&institute)
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 5. Return success response with institution DTO
	return helper.Success(
		c,
		"Institution created successfully",
		dto.ToInstitutionResponseDTO(&createdInstitute),
	)
}

// GetAllInstitutesController handles paginated retrieval of institutions
func (cl *InstituteController) GetAllInstitutesController(c fiber.Ctx) error {
	// 1. Extract logged-in user ID from context
	userID, _ := c.Locals("user_id").(uint)

	// 2. Parse pagination query parameters
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

	// 3. Fetch paginated institutions from service
	institutes, total, err := cl.instituteService.GetInstituteServicePaginated(userID, page, limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	// 4. Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return paginated response
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

// GetInstituteByIDController retrieves a single institution by ID with access control checks
func (cl *InstituteController) GetInstituteByIDController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse institution ID from URL parameter
	idstr := c.Params("id")
	id, err := strconv.ParseUint(idstr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "Invalid institute ID")
	}

	// 3. Fetch institution record from service (service validates access boundaries)
	institute, err := cl.instituteService.GetInstituteServiceById(userID, uint(id))
	if err != nil {
		
			return helper.Error(c, 403, err.Error())
		}
	
	

	// 4. Return institution response
	return helper.Success(
		c,
		"Institute fetched successfully",
		dto.ToInstitutionResponseDTO(&institute),
	)
}

// UpdateInstituteController updates an existing institution record
func (cl *InstituteController) UpdateInstituteController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse institution ID parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid id")
	}

	// 3. Bind and validate request body
	var body dto.UpdateInstitutionDTO
	body.Sanitize()

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "invalid request body")
	}

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 4. Update institution via service (service validates access boundaries)
	if err := cl.instituteService.UpdateInstitutionService(
		userID,
		uint(id),
		&body,
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 5. Fetch updated institution details
	updated, err := cl.instituteService.GetInstituteServiceById(userID, uint(id))
	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	// 6. Return success response with updated institution DTO
	return helper.Success(
		c,
		"Institution updated successfully",
		dto.ToInstitutionResponseDTO(&updated),
	)
}

// DeleteInstituteController deletes (soft deletes) an institution
func (cl *InstituteController) DeleteInstituteController(c fiber.Ctx) error {
	// 1. Extract authenticated user ID
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	// 2. Parse institution ID parameter
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid institute id")
	}

	// 3. Delete institution via service (service validates access boundaries)
	if err := cl.instituteService.DeleteInstitutionService(userID, uint(id)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "access other institution") {
			return helper.Error(c, 403, err.Error())
		}
		return helper.Error(c, 400, err.Error())
	}

	// 4. Return success response
	return helper.Success(
		c,
		"Institution deleted successfully",
		nil,
	)
}
