package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
)

type PrincipalService struct {
	principalRepo  *repository.PrincipalRepository
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
}

func NewPrincipalService(
	principalRepo *repository.PrincipalRepository,
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *PrincipalService {
	return &PrincipalService{
		principalRepo:  principalRepo,
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

func (s *PrincipalService) checkInstitutionAccess(
	userID uint,
	institutionID uint,
) error {
	hasAccess, err := s.userRepo.HasInstitutionAccess(userID, institutionID)
	if err != nil {
		return err
	}

	if !hasAccess {
		return errors.New("access denied")
	}

	return nil
}

func (s *PrincipalService) CreatePrincipalService(userID uint, principal *model.Principal) (model.Principal, error) {
	if principal.InstitutionID == 0 {
		return model.Principal{}, errors.New("institution_id is required")
	}

	if err := s.checkInstitutionAccess(
		userID,
		principal.InstitutionID,
	); err != nil {
		return model.Principal{}, err
	}

	if err := s.userRepo.ValidateUser(principal.UserID); err != nil {
		return model.Principal{}, err
	}

	existingType, err := s.userRepo.CheckUserExistingProfile(principal.UserID)
	if err == nil && existingType != "" {
		return model.Principal{}, errors.New("user is already registered as a " + existingType)
	}

	if err := s.principalRepo.CreatePrincipal(principal); err != nil {
		return model.Principal{}, err
	}

	if principal.UserID != 0 {
		if err := s.userRepo.UpdateUserPrincipalID(principal.UserID, principal.ID); err != nil {
			return model.Principal{}, err
		}
	}

	return *principal, nil
}

func (s *PrincipalService) GetPrincipalService() ([]model.Principal, error) {
	return s.principalRepo.FetchPrincipal()
}

func (s *PrincipalService) GetPrincipalServicePaginated(
	search string,
	page int,
	limit int,
) ([]model.Principal, int64, error) {
	return s.principalRepo.FetchPrincipalPaginated(
		search,
		page,
		limit,
	)
}

func (s *PrincipalService) GetPrincipalServiceById(userID uint, id uint) (*model.Principal, error) {
	user, err := s.userRepo.FindByID(userID)
	if err == nil && user.PrincipalID > 0 {
		if user.PrincipalID != id {
			return nil, errors.New("access denied")
		}
	}

	principal, err := s.principalRepo.FetchPrincipalById(id)
	if err != nil {
		return nil, err
	}

	if err := s.checkInstitutionAccess(
		userID,
		principal.InstitutionID,
	); err != nil {
		return nil, err
	}

	return &principal, nil
}

func (s *PrincipalService) GetPrincipalServiceDeleted() ([]model.Principal, error) {
	return s.principalRepo.FetchPrincipalDeleted()
}

func (s *PrincipalService) DeletePrincipalService(
	userID uint,
	id uint,
) error {
	principal, err := s.principalRepo.FetchPrincipalById(id)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(
		userID,
		principal.InstitutionID,
	); err != nil {
		return err
	}

	return s.principalRepo.DeletePrincipal(id)
}

func (s *PrincipalService) GetActivePrincipalService() (model.Principal, error) {
	return s.principalRepo.GetActivePrincipal()
}

func (s *PrincipalService) GetInactivePrincipalService() (model.Principal, error) {
	return s.principalRepo.GetInactivePrincipal()
}

func (s *PrincipalService) UpdatePrincipalService(
	userID uint,
	id uint,
	dto *dto.UpdatePrincipalDTO,
) error {

	principal, err := s.principalRepo.FetchPrincipalById(id)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(
		userID,
		principal.InstitutionID,
	); err != nil {
		return err
	}

	if dto.Name != "" {
		principal.Name = dto.Name
	}
	if dto.Gender != "" {
		principal.Gender = dto.Gender
	}

	return s.principalRepo.UpdatePrincipalById(&principal)
}
