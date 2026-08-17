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
	institutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(
		faculty.DepartmentID,
	)
	if err != nil {
		return model.Faculty{}, err
	}

	if err := s.checkInstitutionAccess(
		userID,
		institutionID,
	); err != nil {
		return model.Faculty{}, err
	}

	if err := s.userRepo.ValidateUser(faculty.UserID); err != nil {
		return model.Faculty{}, err
	}

	existingType, err := s.userRepo.CheckUserExistingProfile(faculty.UserID)
	if err == nil && existingType != "" {
		return model.Faculty{}, errors.New("user is already registered as a " + existingType)
	}

	if err := s.facultyRepo.CreateFaculty(faculty); err != nil {
		return model.Faculty{}, err
	}

	if faculty.UserID != 0 {
		if err := s.userRepo.UpdateUserFacultyID(faculty.UserID, faculty.ID); err != nil {
			return model.Faculty{}, err
		}
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

func (s *FacultyService) GetFacultyServiceById(userID uint, id uint) (*model.Faculty, error) {
	isSuper, _ := s.userRepo.IsSuperAdmin(userID)
	instAdminID, _ := s.userRepo.GetUserInstitutionID(userID)

	if !isSuper && instAdminID == 0 {
		userFacultyID, err := s.userRepo.GetUserFacultyID(userID)
		if err != nil || userFacultyID == 0 {
			return nil, errors.New("access denied: faculty profile not created yet")
		}
		if userFacultyID != id {
			return nil, errors.New("access denied: you can only access your own faculty profile")
		}
	}

	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return nil, err
	}

	institutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(
		faculty.DepartmentID,
	)
	if err != nil {
		return nil, err
	}

	if err := s.checkInstitutionAccess(
		userID,
		institutionID,
	); err != nil {
		return nil, err
	}

	return &faculty, nil
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
	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return err
	}

	institutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(
		faculty.DepartmentID,
	)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(
		userID,
		institutionID,
	); err != nil {
		return err
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
	dto *dto.UpdateFacultyDTO,
) error {
	isSuper, _ := s.userRepo.IsSuperAdmin(userID)
	instAdminID, _ := s.userRepo.GetUserInstitutionID(userID)

	if !isSuper && instAdminID == 0 {
		userFacultyID, err := s.userRepo.GetUserFacultyID(userID)
		if err != nil || userFacultyID == 0 {
			return errors.New("access denied: faculty profile not created yet")
		}
		if userFacultyID != id {
			return errors.New("access denied: you can only update your own faculty profile")
		}
	}

	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return err
	}

	institutionID, err := s.departmentRepo.GetInstitutionByDepartmentID(
		faculty.DepartmentID,
	)
	if err != nil {
		return err
	}

	if err := s.checkInstitutionAccess(
		userID,
		institutionID,
	); err != nil {
		return err
	}

	if dto.Name != "" {
		faculty.Name = dto.Name
	}
	if dto.Gender != "" {
		faculty.Gender = dto.Gender
	}

	return s.facultyRepo.UpdateFacultyById(&faculty)
}
