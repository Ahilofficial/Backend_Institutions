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
	instituteRepo   *repository.InstitutionRepository
}

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
		instituteRepo:   instituteRepo,
	}
}

func (s *FacultyService) GetNonPaidStudentsForFacultyService(
	userID uint,
) ([]model.Student, error) {

	facultyID, err := s.GetFacultyIDForUserService(userID)
	if err != nil {
		return nil, err
	}

	students, err := s.facultyRepo.GetNonPaidStudentsForFaculty(facultyID)
	if err != nil {
		return nil, err
	}

	return students, nil
}

func (s *FacultyService) GetFacultyIDForUserService(
	userID uint,
) (uint, error) {
	return s.userRepo.GetUserFacultyID(userID)
}

func (s *FacultyService) GetPaidStudentsForFacultyService(
	userID uint,
) ([]model.Student, error) {

	facultyID, err := s.GetFacultyIDForUserService(userID)
	if err != nil {
		return nil, err
	}

	students, err := s.facultyRepo.GetPaidStudentsForFaculty(facultyID)
	if err != nil {
		return nil, err
	}

	return students, nil
}




func (s *FacultyService) CreateFacultyService(
	userID uint,
	body *dto.CreateFacultyDTO,
) (model.Faculty, error) {
	department, err := s.departmentRepo.FetchDepartmentById(body.DepartmentID)
	if err != nil || department.ID == 0 {
		return model.Faculty{}, errors.New("department not found")
	}

	existingType := s.userRepo.CheckUserExistingProfileFaculty(userID)
	if !existingType {
		return model.Faculty{}, errors.New("cannot create faculty profile: user already has a student profile")
	}

	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 {
		return model.Faculty{}, errors.New("faculty profile already exists for this user")
	}

	isInstAdmin, assignedInstID, _ := s.userRepo.IsInstitutionAdmin(userID)
	if isInstAdmin {
		deptInstID, _ := s.departmentRepo.GetInstitutionByDepartmentID(body.DepartmentID)
		if assignedInstID > 0 && deptInstID != assignedInstID {
			return model.Faculty{}, errors.New("Cant able to access other institution")
		}
	}

	faculty := model.Faculty{
		Name:         body.Name,
		Gender:       body.Gender,
		JoiningDate:  body.JoiningDate,
		DepartmentID: body.DepartmentID,
		UserID:       userID,
		IsActive:     true,
	}

	if err := s.facultyRepo.CreateFaculty(&faculty); err != nil {
		return model.Faculty{}, err
	}

	if faculty.UserID > 0 && faculty.ID > 0 {
		_ = s.userRepo.UpdateUserFacultyID(faculty.UserID, faculty.ID)
	}

	return faculty, nil
}

func (s *FacultyService) GetFacultyService() ([]model.Faculty, error) {
	return s.facultyRepo.FetchFaculty()
}

func (s *FacultyService) GetFacultyServicePaginated(
	
	page int,
	limit int,
) ([]model.Faculty, int64, error) {
	return s.facultyRepo.FetchFacultyPaginated(
	
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


func (s *FacultyService) DeleteFacultyService(
	userID uint,
	id uint,
) error {
	faculty, err := s.facultyRepo.FetchFacultyById(id)
	if err != nil {
		return err
	}
	
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if faculty.ID!=userFacultyID{
		return errors.New("you cant able to delete other faculty")
	}


	return s.facultyRepo.DeleteFaculty(id)
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
		logginedUserInstitutionID := s.facultyRepo.LoginnedUserInstitutionIDRepo(userID)
		if logginedUserInstitutionID == 0 || facultyInstitutionID != logginedUserInstitutionID {
			return errors.New("access denied: you can only update your own faculty profile")
		}
	}

	faculty.Name = req.Name
	faculty.Gender = req.Gender
	return s.facultyRepo.UpdateFacultyById(&faculty)
}
