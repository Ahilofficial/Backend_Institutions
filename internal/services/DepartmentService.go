package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
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


func (s *DepartmentService) AddDepartmentService(
	userID uint,
	department *model.Department,

) (model.Department, error) {

	

	if err := s.departmentRepo.CreateDepartment(department); err != nil {
		return model.Department{}, err
	}

	return *department, nil
}

func (s *DepartmentService) GetDepartmentService() ([]model.Department, error) {
	return s.departmentRepo.FetchDepartment()
}

func (s *DepartmentService) GetDepartmentServicePaginated(
	
	page int,
	limit int,
) ([]model.Department, int64, error) {

	return s.departmentRepo.FetchDepartmentPaginated(
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

	
	return department, nil
}

func (s *DepartmentService) DeleteDepartment(
	userID uint,
	id uint,
) error {
	

	

	return s.departmentRepo.DeleteDepartment(id)
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

	

	department.DepartmentName = req.DepartmentName

	return s.departmentRepo.UpdateDepartmentById(&department)
}

func (s *DepartmentService) GetDepartmentFeeService(departmentID uint) (float64, error) {
	return s.departmentRepo.GetDepartmentFee(departmentID)
}

func (s *DepartmentService) GetInstitutionIDForUserService(id uint) uint {
	user_inst_id := s.departmentRepo.GetInstitutionIDForUserRepo(id)
	return user_inst_id
}
