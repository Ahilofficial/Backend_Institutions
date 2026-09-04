package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
)

// DepartmentService encapsulates business logic for department management, access control, and fees
type DepartmentService struct {
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
	instituteRepo  *repository.InstitutionRepository
}

// NewDepartmentService initializes a new DepartmentService instance
func NewDepartmentService(
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
	instituteRepo *repository.InstitutionRepository,
) *DepartmentService {
	return &DepartmentService{
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
		instituteRepo:  instituteRepo,
	}
}

// IsInstAdminService checks if the user is an institution admin
func (s *DepartmentService) IsInstAdminService(userID uint) bool {
	return s.instituteRepo.IsInstAdminRepo(userID)
}

// CheckInstitutionAccess verifies whether an institution admin can manage a specific institution
func (s *DepartmentService) CheckInstitutionAccess(userID uint, targetInstitutionID uint) error {
	if s.IsInstAdminService(userID) {
		userInstitutionID := s.instituteRepo.GetInstitutionIDForUserRepo(userID)
		if userInstitutionID == 0 || userInstitutionID != targetInstitutionID {
			return errors.New("Cant able to access other institution")
		}
	}
	return nil
}

// CheckDepartmentAccess verifies whether an institution admin has access to the target department
func (s *DepartmentService) CheckDepartmentAccess(userID uint, departmentID uint) error {
	if s.IsInstAdminService(userID) {
		userInstitutionID := s.instituteRepo.GetInstitutionIDForUserRepo(userID)
		deptInstitutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(departmentID)
		if err != nil || userInstitutionID == 0 || userInstitutionID != deptInstitutionID {
			return errors.New("Cant able to access other institution")
		}
	}
	return nil
}

// AddDepartmentService creates a new department record after validating access
func (s *DepartmentService) AddDepartmentService(
	userID uint,
	department *model.Department,
) (model.Department, error) {
	// 1. Verify institution access
	if err := s.CheckInstitutionAccess(userID, department.InstitutionID); err != nil {
		return model.Department{}, err
	}

	// 2. Insert department record in repository
	if err := s.departmentRepo.CreateDepartment(department); err != nil {
		return model.Department{}, err
	}

	// 3. Return created department
	return *department, nil
}

// GetDepartmentService fetches all active departments
func (s *DepartmentService) GetDepartmentService() ([]model.Department, error) {
	return s.departmentRepo.FetchDepartment()
}

// GetDepartmentServicePaginated retrieves paginated department records scoped to user's access
func (s *DepartmentService) GetDepartmentServicePaginated(
	userID uint,
	page int,
	limit int,
) ([]model.Department, int64, error) {
	// 1. If institution admin, return only departments from their assigned institution
	if s.IsInstAdminService(userID) {
		instID := s.instituteRepo.GetInstitutionIDForUserRepo(userID)
		if instID > 0 {
			return s.departmentRepo.FetchDepartmentPaginatedByInstitution(instID, page, limit)
		}
		return []model.Department{}, 0, nil
	}

	// 2. Otherwise fetch all paginated departments
	return s.departmentRepo.FetchDepartmentPaginated(
		page,
		limit,
	)
}

// GetDepartmentByIDService retrieves department by primary key ID after validating access
func (s *DepartmentService) GetDepartmentByIDService(
	userID uint,
	id uint,
) (model.Department, error) {
	// 1. Verify department access
	if err := s.CheckDepartmentAccess(userID, id); err != nil {
		return model.Department{}, err
	}

	// 2. Fetch department details with relations
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return model.Department{}, err
	}

	return department, nil
}

// DeleteDepartment handles soft deletion of a department after validating access
func (s *DepartmentService) DeleteDepartment(
	userID uint,
	id uint,
) error {
	// 1. Verify department access
	if err := s.CheckDepartmentAccess(userID, id); err != nil {
		return err
	}

	// 2. Soft delete department in repository
	return s.departmentRepo.DeleteDepartment(id)
}

// UpdateDepartmentService updates department name after validating access
func (s *DepartmentService) UpdateDepartmentService(
	userID uint,
	id uint,
	req *dto.UpdateDepartmentDTO,
) error {
	// 1. Verify department access
	if err := s.CheckDepartmentAccess(userID, id); err != nil {
		return err
	}

	// 2. Fetch existing department record
	department, err := s.departmentRepo.FetchDepartmentById(id)
	if err != nil {
		return err
	}

	// 3. Apply name update
	department.DepartmentName = req.DepartmentName

	// 4. Persist update
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
