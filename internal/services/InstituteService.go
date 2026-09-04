package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
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

// CheckInstitutionAccess verifies whether an institution admin is authorized to access the given institution
func (s *InstituteService) CheckInstitutionAccess(userID uint, targetInstitutionID uint) error {
	if s.IsInstAdminService(userID) {
		userInstitutionID := s.GetInstitutionIDForUserService(userID)
		if userInstitutionID != targetInstitutionID {
			return errors.New("Cant able to access other institution")
		}
	}
	return nil
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

// GetInstituteServicePaginated retrieves a paginated list of institutions scoped to user's access
func (s *InstituteService) GetInstituteServicePaginated(userID uint, page, limit int) ([]model.Institutions, int64, error) {
	// 1. If institution admin, return only their own institution
	if s.IsInstAdminService(userID) {
		instID := s.GetInstitutionIDForUserService(userID)
		if instID > 0 {
			inst, err := s.instituterepo.FetchInstitutionById(instID)
			if err != nil {
				return nil, 0, err
			}
			return []model.Institutions{inst}, 1, nil
		}
		return []model.Institutions{}, 0, nil
	}

	// 2. Fetch paginated institutions from repository
	return s.instituterepo.FetchInstitutionPaginated(page, limit)
}

// GetInstituteServiceById retrieves an institution by primary key ID after verifying access
func (s *InstituteService) GetInstituteServiceById(
	userID uint,
	id uint,
) (model.Institutions, error) {
	// 1. Verify institution admin access boundaries
	if err := s.CheckInstitutionAccess(userID, id); err != nil {
		return model.Institutions{}, err
	}

	// 2. Fetch institution with preloaded hierarchy
	return s.instituterepo.FetchInstitutionById(id)
}

// DeleteInstitutionService handles soft deleting an institution after verifying access
func (s *InstituteService) DeleteInstitutionService(userID uint, id uint) error {
	// 1. Verify institution admin access boundaries
	if err := s.CheckInstitutionAccess(userID, id); err != nil {
		return err
	}

	// 2. Soft delete institution record in repository
	return s.instituterepo.DeleteInstitution(id)
}

// UpdateInstitutionService applies updates to an existing institution record after verifying access
func (s *InstituteService) UpdateInstitutionService(
	userID uint,
	id uint,
	dto *dto.UpdateInstitutionDTO,
) error {
	// 1. Verify institution admin access boundaries
	if err := s.CheckInstitutionAccess(userID, id); err != nil {
		return err
	}

	// 2. Fetch existing institution record
	institute, err := s.instituterepo.FetchInstitutionById(id)
	if err != nil {
		return err
	}

	// 3. Update model fields from DTO
	institute.Name = dto.Name
	institute.InstitutionCode = dto.InstitutionCode
	institute.State = dto.State

	// 4. Persist updated institution
	return s.instituterepo.UpdateInstitution(&institute)
}

// GetInstitutionIDByUserID resolves institution ID mapped to a specific user
func (s *InstituteService) GetInstitutionIDByUserID(userID uint) (uint, error) {
	return s.instituterepo.GetInstitutionIDByUserID(userID)
}

