package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
)

// DepartmentService encapsulates business logic for department management, access control, and fees
type DepartmentService struct {
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
}

// NewDepartmentService initializes a new DepartmentService instance
func NewDepartmentService(
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *DepartmentService {
	return &DepartmentService{
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

// AddDepartmentService creates a new department record
func (s *DepartmentService) AddDepartmentService(
	userID uint,
	department *model.Department,
) (model.Department, error) {
	// 1. Insert department record in repository
	if err := s.departmentRepo.CreateDepartment(department); err != nil {
		return model.Department{}, err
	}

	// 2. Return created department
	return *department, nil
}

// GetDepartmentService fetches all active departments
func (s *DepartmentService) GetDepartmentService() ([]model.Department, error) {
	return s.departmentRepo.FetchDepartment()
}

// GetDepartmentServicePaginated retrieves paginated department records
func (s *DepartmentService) GetDepartmentServicePaginated(
	page int,
	limit int,
) ([]model.Department, int64, error) {
	// 1. Fetch paginated departments from repository
	return s.departmentRepo.FetchDepartmentPaginated(
		page,
		limit,
	)
}

// GetDepartmentByIDService retrieves department by primary key ID
func (s *DepartmentService) GetDepartmentByIDService(
	userID uint,
	id uint,
) (model.Department, error) {
	// 1. Fetch department details with relations
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return model.Department{}, err
	}

	return department, nil
}

// DeleteDepartment handles soft deletion of a department
func (s *DepartmentService) DeleteDepartment(
	userID uint,
	id uint,
) error {
	// 1. Soft delete department in repository
	return s.departmentRepo.DeleteDepartment(id)
}

// UpdateDepartmentService updates department name
func (s *DepartmentService) UpdateDepartmentService(
	userID uint,
	id uint,
	req *dto.UpdateDepartmentDTO,
) error {
	// 1. Fetch existing department record
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return err
	}

	// 2. Apply name update
	department.DepartmentName = req.DepartmentName

	// 3. Persist update
	return s.departmentRepo.UpdateDepartmentById(&department)
}

// GetDepartmentFeeService retrieves department fee amount
func (s *DepartmentService) GetDepartmentFeeService(departmentID uint) (float64, error) {
	return s.departmentRepo.GetDepartmentFee(departmentID)
}

// GetInstitutionIDForUserService resolves the institution ID for a department ID
func (s *DepartmentService) GetInstitutionIDForUserService(id uint) uint {
	user_inst_id := s.departmentRepo.GetInstitutionIDForUserRepo(id)
	return user_inst_id
}
