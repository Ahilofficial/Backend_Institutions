package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
)

// InstituteService provides business logic for institution creation, querying, updating, and access verification
type InstituteService struct {
	instituterepo *repository.InstitutionRepository
	userrepo      *repository.UserRepository
}

// NewInstituteService creates a new instance of InstituteService
func NewInstituteService(
	instituterepo *repository.InstitutionRepository,
	userrepo *repository.UserRepository,
) *InstituteService {
	return &InstituteService{
		instituterepo: instituterepo,
		userrepo:      userrepo,
	}
}

// GetInstitutionIDForUserService gets the institution ID mapped to a user
func (s *InstituteService) GetInstitutionIDForUserService(userID uint) uint {
	// 1. Resolve institution ID from repository
	value := s.instituterepo.GetInstitutionIDForUserRepo(userID)
	return value
}

// IsInstAdminService checks if the user is an institution admin
func (s *InstituteService) IsInstAdminService(userID uint) bool {
	// 1. Check institution admin membership in repository
	IsInstAdmin := s.instituterepo.IsInstAdminRepo(userID)
	return IsInstAdmin
}

// CreateInsituteService handles business logic to create and store a new institution
func (s *InstituteService) CreateInsituteService(institute *model.Institutions) (model.Institutions, error) {
	// 1. Persist institution record
	err := s.instituterepo.CreateInstitution(institute)
	if err != nil {
		return model.Institutions{}, err
	}

	// 2. Return created institution entity
	return *institute, nil
}

// GetInstituteServicePaginated retrieves a paginated list of institutions
func (s *InstituteService) GetInstituteServicePaginated(userID uint, page, limit int) ([]model.Institutions, int64, error) {
	// 1. Fetch paginated institutions from repository
	return s.instituterepo.FetchInstitutionPaginated(page, limit)
}

// GetInstituteServiceById retrieves an institution by primary key ID
func (s *InstituteService) GetInstituteServiceById(
	userID uint,
	id uint,
) (model.Institutions, error) {
	// 1. Fetch institution with preloaded hierarchy
	return s.instituterepo.FetchInstitutionById(id)
}

// DeleteInstitutionService handles soft deleting an institution
func (s *InstituteService) DeleteInstitutionService(id uint) error {
	// 1. Soft delete institution record in repository
	return s.instituterepo.DeleteInstitution(id)
}

// UpdateInstitutionService applies updates to an existing institution record
func (s *InstituteService) UpdateInstitutionService(
	userID uint,
	id uint,
	dto *dto.UpdateInstitutionDTO,
) error {
	// 1. Fetch existing institution record
	institute, err := s.instituterepo.FetchInstitutionById(id)
	if err != nil {
		return err
	}

	// 2. Update model fields from DTO
	institute.Name = dto.Name
	institute.InstitutionCode = dto.InstitutionCode
	institute.State = dto.State

	// 3. Persist updated institution
	return s.instituterepo.UpdateInstitution(&institute)
}

// GetInstitutionIDByUserID resolves institution ID mapped to a specific user
func (s *InstituteService) GetInstitutionIDByUserID(userID uint) (uint, error) {
	return s.instituterepo.GetInstitutionIDByUserID(userID)
}

