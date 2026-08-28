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

	targetUserID := userID
	if principal.UserID > 0 {
		targetUserID = principal.UserID
	}

	if principal.UserID != 0 && principal.UserID != userID {
		if err := s.checkInstitutionAccess(userID, principal.InstitutionID); err != nil {
			return model.Principal{}, err
		}
	}

	if targetUserID > 0 {
		if err := s.userRepo.ValidateUser(targetUserID); err != nil {
			return model.Principal{}, err
		}

		existingType, err := s.userRepo.CheckUserExistingProfile(targetUserID)
		if err == nil && existingType != "" {
			return model.Principal{}, errors.New("user is already registered as a " + existingType)
		}
	}

	principal.UserID = targetUserID
	if err := s.principalRepo.CreatePrincipal(principal); err != nil {
		return model.Principal{}, err
	}

	if principal.UserID != 0 {
		_ = s.userRepo.UpdateUserPrincipalID(principal.UserID, principal.ID)
		_ = s.userRepo.AssignRoleByName(principal.UserID, "principal")
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

func (s *PrincipalService) GetPrincipalServiceById(
	userID uint,
	id uint,
) (*model.Principal, error) {
	principal, err := s.principalRepo.FetchPrincipalById(id)
	if err != nil {
		return nil, err
	}

	userPrincipalID, _ := s.userRepo.GetUserPrincipalID(userID)
	if userPrincipalID > 0 {
		if userPrincipalID != id {
			return nil, errors.New("access denied: you can only access your own principal profile")
		}
		return &principal, nil
	}

	if err := s.checkInstitutionAccess(userID, principal.InstitutionID); err != nil {
		return nil, errors.New("access denied: principal does not belong to your institution")
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
	userPrincipalID, _ := s.userRepo.GetUserPrincipalID(userID)
	if userPrincipalID > 0 {
		return errors.New("access denied: principal cannot delete principal profiles")
	}

	principal, err := s.principalRepo.FetchPrincipalById(id)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(userID, principal.InstitutionID); err != nil {
		return errors.New("access denied: principal does not belong to your institution")
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
	req *dto.UpdatePrincipalDTO,
) error {
	principal, err := s.principalRepo.FetchPrincipalById(id)
	if err != nil {
		return err
	}

	userPrincipalID, _ := s.userRepo.GetUserPrincipalID(userID)
	if userPrincipalID > 0 {
		if userPrincipalID != id {
			return errors.New("access denied: you can only update your own principal profile")
		}
	} else {
		if err := s.checkInstitutionAccess(userID, principal.InstitutionID); err != nil {
			return errors.New("access denied: cannot update this principal profile")
		}
	}

	principal.Name = req.Name
	principal.Gender = req.Gender
	return s.principalRepo.UpdatePrincipalById(&principal)
}