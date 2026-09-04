package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
)

// FacultyService provides business logic operations for faculty entities
type FacultyService struct {
	facultyRepo    *repository.FacultyRepository
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
	instituteRepo  *repository.InstitutionRepository
}

// NewFacultyService creates a new instance of FacultyService with required repositories
func NewFacultyService(
	facultyRepo *repository.FacultyRepository,
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
	instituteRepo *repository.InstitutionRepository,
) *FacultyService {
	return &FacultyService{
		facultyRepo:    facultyRepo,
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
		instituteRepo:  instituteRepo,
	}
}

// GetNonPaidStudentsForFacultyService fetches students assigned to faculty who have pending fee payments
func (s *FacultyService) GetNonPaidStudentsForFacultyService(
	userID uint,
) ([]model.Student, error) {
	// 1. Resolve faculty ID for the logged-in user
	facultyID, err := s.GetFacultyIDForUserService(userID)
	if err != nil || facultyID == 0 {
		return nil, errors.New("faculty profile not found for logged in user")
	}

	// 2. Fetch non-paid/pending students assigned to this faculty
	students, err := s.facultyRepo.GetNonPaidStudentsForFaculty(facultyID)
	if err != nil {
		return nil, err
	}

	return students, nil
}

// GetFacultyIDForUserService resolves the faculty ID linked to a user account
func (s *FacultyService) GetFacultyIDForUserService(
	userID uint,
) (uint, error) {
	return s.userRepo.GetUserFacultyID(userID)
}

// GetPaidStudentsForFacultyService fetches students assigned to faculty who have completed fee payments
func (s *FacultyService) GetPaidStudentsForFacultyService(
	userID uint,
) ([]model.Student, error) {
	// 1. Resolve faculty ID for the logged-in user
	facultyID, err := s.GetFacultyIDForUserService(userID)
	if err != nil || facultyID == 0 {
		return nil, errors.New("faculty profile not found for logged in user")
	}

	// 2. Fetch fully paid students assigned to this faculty
	students, err := s.facultyRepo.GetPaidStudentsForFaculty(facultyID)
	if err != nil {
		return nil, err
	}

	return students, nil
}

// CreateFacultyService validates requirements and registers a new faculty profile
func (s *FacultyService) CreateFacultyService(
	userID uint,
	body *dto.CreateFacultyDTO,
) (model.Faculty, error) {

	// 1. Check department existence and validity
	department, err := s.departmentRepo.FetchDepartmentById(body.DepartmentID)
	if err != nil || department.ID == 0 {
		return model.Faculty{}, errors.New("department not found")
	}

	// 2. Check whether user already has Student or Faculty profile
	profileExists, message := s.userRepo.CheckUserExistingProfileFaculty(userID)
	if profileExists {
		return model.Faculty{}, errors.New(message)
	}

	// 3. Get department's institution ID
	departmentInstitutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(body.DepartmentID)
	if err != nil || departmentInstitutionID == 0 {
		return model.Faculty{}, errors.New("institution not found for this department")
	}

	// 4. Get logged-in user's institution ID
	loggedInUserInstitutionID := s.facultyRepo.LoginnedUserInstitutionIDRepo(userID)

	// 5. Check institution access authorization
	isInstAdmin := s.instituteRepo.IsInstAdminRepo(userID)
	
	if isInstAdmin{
		if loggedInUserInstitutionID!=departmentInstitutionID{
			return model.Faculty{},errors.New("you can create faculty only for your institution")
		}
	}

	// 6. Assemble faculty model
	faculty := model.Faculty{
		Name:         body.Name,
		Gender:       body.Gender,
		JoiningDate:  body.JoiningDate,
		DepartmentID: body.DepartmentID,
		UserID:       userID,
		IsActive:     true,
	}

	// 7. Save faculty record to database
	if err := s.facultyRepo.CreateFaculty(&faculty); err != nil {
		return model.Faculty{}, err
	}

	return faculty, nil
}

// CheckFacultyAccess verifies whether the authenticated user has rights to view or update the target faculty
func (s *FacultyService) CheckFacultyAccess(userID uint, targetFacultyID uint) error {
	// 1. If user is a faculty member, they can only view/update their own profile
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		if userFacultyID != targetFacultyID {
			return errors.New("access denied: you can only access your own faculty profile")
		}
		return nil
	}

	// 2. If user is an institution admin, verify the target faculty belongs to their institution
	if s.instituteRepo.IsInstAdminRepo(userID) {
		adminInstID := s.instituteRepo.GetInstitutionIDForUserRepo(userID)
		facultyInstID := s.facultyRepo.GetInstitutionIDForUserRepo(targetFacultyID)
		if adminInstID == 0 || adminInstID != facultyInstID {
			return errors.New("Cant able to access other institution")
		}
		return nil
	}

	return nil
}

// GetFacultyService fetches all active faculty records
func (s *FacultyService) GetFacultyService() ([]model.Faculty, error) {
	return s.facultyRepo.FetchFaculty()
}

// GetFacultyServicePaginated fetches faculty records with pagination support scoped to user role
func (s *FacultyService) GetFacultyServicePaginated(
	userID uint,
	page int,
	limit int,
) ([]model.Faculty, int64, error) {
	// 1. If user is a faculty member, return only their own profile
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		fac, err := s.facultyRepo.FetchFacultyById(userFacultyID)
		if err != nil {
			return nil, 0, err
		}
		return []model.Faculty{fac}, 1, nil
	}

	// 2. If user is an institution admin, return only faculties belonging to their institution
	if s.instituteRepo.IsInstAdminRepo(userID) {
		adminInstID := s.instituteRepo.GetInstitutionIDForUserRepo(userID)
		if adminInstID > 0 {
			return s.facultyRepo.FetchFacultyPaginatedByInstitution(adminInstID, page, limit)
		}
		return []model.Faculty{}, 0, nil
	}

	// 3. Otherwise return all paginated faculties
	return s.facultyRepo.FetchFacultyPaginated(
		page,
		limit,
	)
}

// GetFacultyServiceById fetches a single faculty record by ID after verifying access
func (s *FacultyService) GetFacultyServiceById(
	userID uint,
	id uint,
) (*model.Faculty, error) {
	// 1. Verify access permissions
	if err := s.CheckFacultyAccess(userID, id); err != nil {
		return nil, err
	}

	// 2. Fetch faculty from repository
	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return nil, err
	}

	return &faculty, nil
}

// GetInstitutionIDForUserRepo retrieves institution ID associated with a faculty ID
func (s *FacultyService) GetInstitutionIDForUserRepo(facultyID uint) uint {
	return s.facultyRepo.GetInstitutionIDForUserRepo(facultyID)
}

// GetInstitutionByDepartmentID retrieves institution ID associated with a department ID
func (s *FacultyService) GetInstitutionByDepartmentID(deptID uint) (uint, error) {
	return s.departmentRepo.GetInstitutionByDepartmentID(deptID)
}

// LoginnedUserInstitutionIDService retrieves institution ID associated with the logged-in user
func (s *FacultyService) LoginnedUserInstitutionIDService(userID uint) uint {
	logginedUserInstitutionID := s.facultyRepo.LoginnedUserInstitutionIDRepo(userID)
	return logginedUserInstitutionID
}

// GetLoggedInFacultyProfile fetches the faculty profile belonging to the authenticated user
func (s *FacultyService) GetLoggedInFacultyProfile(userID uint) (*model.Faculty, error) {
	// 1. Retrieve faculty ID linked to user
	facultyID, err := s.userRepo.GetUserFacultyID(userID)
	if err != nil || facultyID == 0 {
		return nil, errors.New("faculty profile not created yet for logged in user")
	}

	// 2. Fetch full faculty profile details
	return s.GetFacultyServiceById(userID, facultyID)
}

// GetLoggedInFacultyStudents fetches all students assigned to the logged-in faculty member
func (s *FacultyService) GetLoggedInFacultyStudents(userID uint) ([]model.Student, error) {
	// 1. Retrieve faculty ID linked to user
	facultyID, err := s.userRepo.GetUserFacultyID(userID)
	if err != nil || facultyID == 0 {
		return nil, errors.New("faculty profile not created yet for logged in user")
	}

	// 2. Fetch students assigned to this faculty ID
	return s.facultyRepo.FetchStudentsByFacultyID(facultyID)
}

// DeleteFacultyService handles soft deletion of a faculty record after verifying access
func (s *FacultyService) DeleteFacultyService(
	userID uint,
	id uint,
) error {
	// 1. Verify access authorization
	if err := s.CheckFacultyAccess(userID, id); err != nil {
		return err
	}

	// 2. Delete faculty record in repository
	return s.facultyRepo.DeleteFaculty(id)
}

// UpdateFacultyService validates authorization and updates faculty profile fields
func (s *FacultyService) UpdateFacultyService(
	userID uint,
	id uint,
	req *dto.UpdateFacultyDTO,
) error {
	// 1. Verify access authorization
	if err := s.CheckFacultyAccess(userID, id); err != nil {
		return err
	}

	// 2. Fetch existing faculty record
	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return err
	}

	// 3. Update faculty profile fields
	faculty.Name = req.Name
	faculty.Gender = req.Gender
	return s.facultyRepo.UpdateFacultyById(&faculty)
}
