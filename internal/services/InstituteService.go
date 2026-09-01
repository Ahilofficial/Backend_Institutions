package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
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

func (s *InstituteService) GetInstitutionIDForUserService(userID uint) uint {
	value := s.instituterepo.GetInstitutionIDForUserRepo(userID)
	return value

}
func (s *InstituteService) IsInstAdminService(userID uint) bool {
	IsInstAdmin := s.instituterepo.IsInstAdminRepo(userID)
	return IsInstAdmin
}

func (s *InstituteService) CreateInsituteService(institute *model.Institutions) (model.Institutions, error) {
	err := s.instituterepo.CreateInstitution(institute)
	if err != nil {
		return model.Institutions{}, err
	}

	return *institute, nil
}



func (s *InstituteService) GetInstituteServicePaginated(userID uint, page, limit int) ([]model.Institutions, int64, error) {

	return s.instituterepo.FetchInstitutionPaginated(page, limit)
}

func (s *InstituteService) GetInstituteServiceById(
	userID uint,
	id uint,
) (model.Institutions, error) {

	return s.instituterepo.FetchInstitutionById(id)
}

func (s *InstituteService) DeleteInstitutionService(id uint) error {
	
	return s.instituterepo.DeleteInstitution(id)
}

func (s *InstituteService) UpdateInstitutionService(
	userID uint,
	id uint,
	dto *dto.UpdateInstitutionDTO,
) error {

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
