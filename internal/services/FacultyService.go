package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
)

type FacultyService struct {
	facultyRepo    *repository.FacultyRepository
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
}

func NewFacultyService(
	facultyRepo *repository.FacultyRepository,
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *FacultyService {
	return &FacultyService{
		facultyRepo:    facultyRepo,
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

func (s *FacultyService) checkInstitutionAccess(
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

func (s *FacultyService) CreateFacultyService(userID uint, faculty *model.Faculty) (model.Faculty, error) {
	department, err := s.departmentRepo.FetchDepartmentById(faculty.DepartmentID)
	if err != nil || department.ID == 0 {
		return model.Faculty{}, errors.New("department not found")
	}

	targetUserID := userID
	if faculty.UserID > 0 {
		targetUserID = faculty.UserID
	}

	if targetUserID > 0 {
		if err := s.userRepo.ValidateUser(targetUserID); err != nil {
			return model.Faculty{}, err
		}

		existingType, err := s.userRepo.CheckUserExistingProfile(targetUserID)
		if err == nil && existingType != "" {
			return model.Faculty{}, errors.New("user is already registered as a " + existingType)
		}
	}

	faculty.UserID = targetUserID
	if err := s.facultyRepo.CreateFaculty(faculty); err != nil {
		return model.Faculty{}, err
	}

	if faculty.UserID != 0 {
		_ = s.userRepo.UpdateUserFacultyID(faculty.UserID, faculty.ID)
		_ = s.userRepo.AssignRoleByName(faculty.UserID, "faculty")
	}

	return *faculty, nil
}

func (s *FacultyService) GetFacultyService() ([]model.Faculty, error) {
	return s.facultyRepo.FetchFaculty()
}

func (s *FacultyService) GetFacultyServicePaginated(
	search string,
	page int,
	limit int,
) ([]model.Faculty, int64, error) {
	return s.facultyRepo.FetchFacultyPaginated(
		search,
		page,
		limit,
	)
}

func (s *FacultyService) GetFacultyServiceById(
	userID uint,
	id uint,
) (*model.Faculty, error) {
	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return nil, err
	}

	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		if userFacultyID != id {
			return nil, errors.New("access denied: you can only access your own faculty profile")
		}
		return &faculty, nil
	}

	// 2. For Super Admin, Institution Admin, or Principal: check institution access
	facultyInstitutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(faculty.DepartmentID)
	if err != nil {
		return nil, err
	}

	if err := s.checkInstitutionAccess(userID, facultyInstitutionID); err != nil {
		return nil, errors.New("access denied: faculty does not belong to your institution")
	}

	return &faculty, nil
}

func (s *FacultyService) GetInstitutionIDForUserRepo(facultyID uint) uint {
	return s.facultyRepo.GetInstitutionIDForUserRepo(facultyID)
}

func (s *FacultyService) GetInstitutionByDepartmentID(deptID uint) (uint, error) {
	return s.departmentRepo.GetInstitutionByDepartmentID(deptID)
}

func (s *FacultyService) LoginnedUserInstitutionIDService(userID uint) uint {
	logginedUserInstitutionID := s.facultyRepo.LoginnedUserInstitutionIDRepo(userID)
	return logginedUserInstitutionID
}
func (s *FacultyService) GetLoggedInFacultyProfile(userID uint) (*model.Faculty, error) {
	facultyID, err := s.userRepo.GetUserFacultyID(userID)
	if err != nil || facultyID == 0 {
		return nil, errors.New("faculty profile not created yet for logged in user")
	}
	return s.GetFacultyServiceById(userID, facultyID)
}

func (s *FacultyService) GetLoggedInFacultyStudents(userID uint) ([]model.Student, error) {
	facultyID, err := s.userRepo.GetUserFacultyID(userID)
	if err != nil || facultyID == 0 {
		return nil, errors.New("faculty profile not created yet for logged in user")
	}
	return s.facultyRepo.FetchStudentsByFacultyID(facultyID)
}

func (s *FacultyService) GetFacultyServiceDeleted() ([]model.Faculty, error) {
	return s.facultyRepo.FetchFacultyDeleted()
}

func (s *FacultyService) DeleteFacultyService(
	userID uint,
	id uint,
) error {
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		return errors.New("access denied: faculty cannot delete faculty profiles")
	}

	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return err
	}

	facultyInstitutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(faculty.DepartmentID)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(userID, facultyInstitutionID); err != nil {
		return errors.New("access denied: faculty does not belong to your institution")
	}

	return s.facultyRepo.DeleteFaculty(id)
}

func (s *FacultyService) GetActiveFacultyService() (model.Faculty, error) {
	return s.facultyRepo.GetActiveFaculty()
}

func (s *FacultyService) GetInactiveFacultyService() (model.Faculty, error) {
	return s.facultyRepo.GetInactiveFaculty()
}

func (s *FacultyService) UpdateFacultyService(
	userID uint,
	id uint,
	req *dto.UpdateFacultyDTO,
) error {
	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return err
	}

	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		if userFacultyID != id {
			return errors.New("access denied: you can only update your own faculty profile")
		}
	} else {
		facultyInstitutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(faculty.DepartmentID)
		if err != nil {
			return err
		}
		if err := s.checkInstitutionAccess(userID, facultyInstitutionID); err != nil {
			return errors.New("access denied: cannot update this faculty profile")
		}
	}

	faculty.Name = req.Name
	faculty.Gender = req.Gender
	return s.facultyRepo.UpdateFacultyById(&faculty)
}


