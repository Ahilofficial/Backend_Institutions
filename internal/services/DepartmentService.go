package services

import (
	"errors"

	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"

	// "github.com/gofiber/fiber/v3"
)

type DepartmentService struct {
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
}

func NewDepartmentService(
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *DepartmentService {
	return &DepartmentService{
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

func (s *DepartmentService) checkInstitutionAccess(
	userID uint,
	institutionID uint,
) error {

	hasAccess, err := s.userRepo.HasInstitutionAccess(
		userID,
		institutionID,
	)
	if err != nil {
		return err
	}

	if !hasAccess {
		return errors.New(
			"user does not have access to this institution",
		)
	}

	return nil
}

func (s *DepartmentService) AddDepartmentService(
	userID uint,
	department *model.Department,
	
) (model.Department, error) {

	if err := s.checkInstitutionAccess(
		userID,
		department.InstitutionID,
	); err != nil {
		return model.Department{}, err
	}
	
	if err := s.departmentRepo.CreateDepartment(department); err != nil {
		return model.Department{}, err
	}

	return *department, nil
}

func (s *DepartmentService) GetDepartmentService() ([]model.Department, error) {
	return s.departmentRepo.FetchDepartment()
}

func (s *DepartmentService) GetDepartmentServicePaginated(
	search string,
	page int,
	limit int,
) ([]model.Department, int64, error) {

	return s.departmentRepo.FetchDepartmentPaginated(
		search,
		page,
		limit,
	)
}

func (s *DepartmentService) GetDepartmentByIDService(
	userID uint,
	id uint,
) (model.Department, error) {
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return model.Department{}, err
	}

	if err := s.checkInstitutionAccess(userID, department.InstitutionID); err != nil {
		return model.Department{}, errors.New("access denied: department does not belong to your institution")
	}

	return department, nil
}

func (s *DepartmentService) GetDepartmentServiceDeleted() ([]model.Department, error) {
	return s.departmentRepo.FetchDepartmentDeleted()
}

func (s *DepartmentService) DeleteDepartment(
	userID uint,
	id uint,
) error {
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(userID, department.InstitutionID); err != nil {
		return errors.New("access denied: department does not belong to your institution")
	}

	return s.departmentRepo.DeleteDepartment(id)
}

func (s *DepartmentService) GetActiveDepartmentService() (model.Department, error) {
	return s.departmentRepo.GetActiveDepartment()
}

func (s *DepartmentService) GetInactiveDepartmentService() (model.Department, error) {
	return s.departmentRepo.GetInactiveDepartment()
}

func (s *DepartmentService) UpdateDepartmentService(
	userID uint,
	id uint,
	req *dto.UpdateDepartmentDTO,
) error {
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(userID, department.InstitutionID); err != nil {
		return errors.New("access denied: department does not belong to your institution")
	}

	department.DepartmentName = req.DepartmentName
	// department.FeeAmount = req.FeeAmount
	return s.departmentRepo.UpdateDepartmentById(&department)
}

func (s *DepartmentService) GetDepartmentFeeService(departmentID uint) (float64, error) {
	return s.departmentRepo.GetDepartmentFee(departmentID)
}

func(s *DepartmentService)GetInstitutionIDForUserService(id uint)(uint){
	user_inst_id:=s.departmentRepo.GetInstitutionIDForUserRepo(id)
	return user_inst_id
}