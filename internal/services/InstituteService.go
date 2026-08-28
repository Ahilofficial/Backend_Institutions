package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
)

type InstituteService struct {
	instituterepo *repository.InstitutionRepository
	userrepo      *repository.UserRepository
}

func NewInstituteService(
	instituterepo *repository.InstitutionRepository,
	userrepo *repository.UserRepository,
) *InstituteService {
	return &InstituteService{
		instituterepo: instituterepo,
		userrepo:      userrepo,
	}
}

func(s *InstituteService)GetInstitutionIDForUserService(userID uint)(uint){
	value:=s.instituterepo.GetInstitutionIDForUserRepo(userID)
	return value

}
func(s *InstituteService)IsInstAdminService(userID uint)(bool){
	IsInstAdmin:=s.instituterepo.IsInstAdminRepo(userID)
	return  IsInstAdmin
}


func (s *InstituteService) CreateInsituteService(institute *model.Institutions) (model.Institutions, error) {
	err := s.instituterepo.CreateInstitution(institute)
	if err != nil {
		return model.Institutions{}, err
	}

	return *institute, nil
}

func (s *InstituteService) GetInstituteService(userID uint) ([]model.Institutions, error) {
	isInstAdmin, assignedInstID, err := s.userrepo.IsInstitutionAdmin(userID)
	if err == nil && isInstAdmin {
		if assignedInstID == 0 {
			return []model.Institutions{}, nil
		}
		inst, err := s.instituterepo.FetchInstitutionById(assignedInstID)
		if err != nil {
			return []model.Institutions{}, err
		}
		return []model.Institutions{inst}, nil
	}

	return s.instituterepo.FetchInstitution()
}

func (s *InstituteService) GetInstituteServicePaginated(userID uint, search string, page, limit int) ([]model.Institutions, int64, error) {
	isInstAdmin, assignedInstID, err := s.userrepo.IsInstitutionAdmin(userID)
	if err == nil && isInstAdmin {
		if assignedInstID == 0 {
			return []model.Institutions{}, 0, nil
		}
		inst, err := s.instituterepo.FetchInstitutionById(assignedInstID)
		if err != nil {
			return []model.Institutions{}, 0, err
		}
		return []model.Institutions{inst}, 1, nil
	}

	return s.instituterepo.FetchInstitutionPaginated(search, page, limit)
}

func (s *InstituteService) GetInstituteServiceById(
	userID uint,
	id uint,
) (model.Institutions, error) {

	// 1. Check if user is an Institution Admin (via user_roles or institution_admins table)
	isInstAdmin, assignedInstID, err := s.userrepo.IsInstitutionAdmin(userID)
	if err == nil && isInstAdmin {
		if assignedInstID == 0 || assignedInstID != id {
			return model.Institutions{}, errors.New("access denied: you can only access your assigned institution")
		}
		return s.instituterepo.FetchInstitutionById(id)
	}


	return s.instituterepo.FetchInstitutionById(id)
}

func (s *InstituteService) GetInstituteServiceDeleted() ([]model.Institutions, error) {
	return s.instituterepo.FetchInstitutionDeleted()
}

func (s *InstituteService) DeleteInstitutionService(userID uint, id uint) error {
	isSuper, err := s.userrepo.IsSuperAdmin(userID)
	if err != nil || !isSuper {
		return errors.New("access denied: only super admin can delete an institution")
	}

	return s.instituterepo.DeleteInstitution(id)
}

func (s *InstituteService) GetActiveInstitute() (model.Institutions, error) {
	return s.instituterepo.GetActiveInstitute()
}

func (s *InstituteService) GetInactiveInstitute() (model.Institutions, error) {
	return s.instituterepo.GetInactiveInstitute()
}


func (s *InstituteService) UpdateInstitutionService(
	userID uint,
	id uint,
	dto *dto.UpdateInstitutionDTO,
) error {
	isInstAdmin, assignedInstID, err := s.userrepo.IsInstitutionAdmin(userID)
	if err == nil && isInstAdmin {
		if assignedInstID == 0 || assignedInstID != id {
			return errors.New("access denied: you can only update your assigned institution")
		}
	} else {
		isSuper, err := s.userrepo.IsSuperAdmin(userID)
		if err != nil || !isSuper {
			return errors.New("access denied: only super admin or institution admin can update institution")
		}
	}

	institute, err := s.instituterepo.FetchInstitutionById(id)
	if err != nil {
		return err
	}

	// Update fields
	institute.Name = dto.Name
	institute.InstitutionCode = dto.InstitutionCode
	institute.State = dto.State

	return s.instituterepo.UpdateInstitution(&institute)
}

func (s *InstituteService) GetInstitutionIDByUserID(userID uint) (uint, error) {
	return s.instituterepo.GetInstitutionIDByUserID(userID)
}
