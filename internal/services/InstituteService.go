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

func (s *InstituteService) checkInstitutionAccess(userID uint, institutionID uint) error {
	hasAccess, err := s.userrepo.HasInstitutionAccess(userID, institutionID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("access denied")
	}
	return nil
}

func (s *InstituteService) CreateInsituteService(institute *model.Institutions) (model.Institutions, error) {
	err := s.instituterepo.CreateInstitution(institute)
	if err != nil {
		return model.Institutions{}, err
	}

	return *institute, nil
}

func (s *InstituteService) GetInstituteService() ([]model.Institutions, error) {
	return s.instituterepo.FetchInstitution()
}

func (s *InstituteService) GetInstituteServicePaginated(search string, page, limit int) ([]model.Institutions, int64, error) {
	return s.instituterepo.FetchInstitutionPaginated(search, page, limit)
}

func (s *InstituteService) GetInstituteServiceById(userID uint, id uint) (model.Institutions, error) {
	if err := s.checkInstitutionAccess(userID, id); err != nil {
		return model.Institutions{}, err
	}

	return s.instituterepo.FetchInstitutionById(id)
}

func (s *InstituteService) GetInstituteServiceDeleted() ([]model.Institutions, error) {
	return s.instituterepo.FetchInstitutionDeleted()
}

func (s *InstituteService) DeleteInstitutionService(userID uint, id uint) error {
	if err := s.checkInstitutionAccess(userID, id); err != nil {
		return err
	}

	return s.instituterepo.DeleteInstitution(id)
}

func (s *InstituteService) GetActiveInstitute() (model.Institutions, error) {
	return s.instituterepo.GetActiveInstitute()
}

func (s *InstituteService) GetInactiveInstitute() (model.Institutions, error) {
	return s.instituterepo.GetInactiveInstitute()
}

func (s *InstituteService) UpdateInstitutionService(userID uint, id uint, dto *dto.UpdateInstitutionDTO) error {
	if err := s.checkInstitutionAccess(userID, id); err != nil {
		return err
	}

	institute, err := s.instituterepo.FetchInstitutionById(id)
	if err != nil {
		return err
	}

	institute.Name = dto.Name
	institute.InstitutionCode = dto.InstitutionCode
	institute.State = dto.State

	return s.instituterepo.UpdateInstitution(&institute)
}

func (s *InstituteService) GetInstitutionIDByUserID(userID uint) (uint, error) {
	return s.instituterepo.GetInstitutionIDByUserID(userID)
}
