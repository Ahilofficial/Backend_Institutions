package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
)

type StudentService struct {
	studentRepo *repository.StudentRepository
	facultyRepo *repository.FacultyRepository
	userRepo    *repository.UserRepository
}

func NewStudentService(
	studentRepo *repository.StudentRepository,
	facultyRepo *repository.FacultyRepository,
	userRepo    *repository.UserRepository,
) *StudentService {
	return &StudentService{
		studentRepo: studentRepo,
		facultyRepo: facultyRepo,
		userRepo:    userRepo,
	}
}

func (s *StudentService) checkInstitutionAccess(
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

func (s *StudentService) GetUserFacultyID(userID uint) (uint, error) {
	return s.userRepo.GetUserFacultyID(userID)
}

func (s *StudentService) GetUserExistingProfile(userID uint) (string, error) {
	return s.userRepo.CheckUserExistingProfile(userID)
}

func (s *StudentService) CreateStudentService(
	userID uint,
	student *model.Student,
) (*model.Student, error) {

	institutionID, err := s.facultyRepo.GetInstitutionByFacultyID(
		student.FacultyID,
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

	if student.UserID > 0 {
		if err := s.userRepo.ValidateUser(student.UserID); err != nil {
			return nil, err
		}

		existingType, err := s.userRepo.CheckUserExistingProfile(student.UserID)
		if err == nil && existingType != "" {
			return nil, errors.New("user is already registered as a " + existingType)
		}
	}

	if err := s.studentRepo.CreateStudent(student); err != nil {
		return nil, err
	}

	if student.UserID != 0 {
		if err := s.userRepo.UpdateUserStudentID(student.UserID, student.ID); err != nil {
			return nil, err
		}
	}

	return student, nil
}

func (s *StudentService) FetchAllStudentsServices() ([]model.Student, error) {
	return s.studentRepo.FetchStudent()
}

func (s *StudentService) FetchAllStudentsPaginatedServices(
	search string,
	page int,
	limit int,
) ([]model.Student, int64, error) {
	return s.studentRepo.FetchStudentPaginated(
		search,
		page,
		limit,
	)
}

func (s *StudentService) GetStudentServiceById(
	userID uint,
	id uint,
) (*model.Student, error) {
	isSuper, _ := s.userRepo.IsSuperAdmin(userID)
	instAdminID, _ := s.userRepo.GetUserInstitutionID(userID)

	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return nil, err
	}

	if !isSuper && instAdminID == 0 {
		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)

		if userStudentID > 0 {
			if userStudentID != id {
				return nil, errors.New("access denied: you can only access your own student profile")
			}
		} else if userFacultyID > 0 {
			if student.FacultyID != userFacultyID {
				return nil, errors.New("access denied: student does not belong to your faculty")
			}
		} else {
			return nil, errors.New("access denied: profile not created yet")
		}
	}

	institutionID, err := s.facultyRepo.GetInstitutionByFacultyID(
		student.FacultyID,
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

	return &student, nil
}

func (s *StudentService) GetLoggedInStudentProfile(userID uint) (*model.Student, error) {
	studentID, err := s.userRepo.GetUserStudentID(userID)
	if err != nil || studentID == 0 {
		return nil, errors.New("student profile not created yet for logged in user")
	}
	return s.GetStudentServiceById(userID, studentID)
}

func (s *StudentService) FetchStudentsByPaymentMonthService(
	userID uint,
	month string,
) ([]model.Student, error) {
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	instID, _ := s.userRepo.GetUserInstitutionID(userID)
	return s.studentRepo.FetchPaidStudentsByMonth(instID, facultyID, month)
}

func (s *StudentService) FetchStudentsNotPaidByMonthService(
	userID uint,
	month string,
) ([]model.Student, error) {
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	instID, _ := s.userRepo.GetUserInstitutionID(userID)
	return s.studentRepo.FetchNotPaidStudentsByMonth(instID, facultyID, month)
}

func (s *StudentService) FetchPaidStudentsService(userID uint, month string) ([]model.Student, error) {
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	instID, _ := s.userRepo.GetUserInstitutionID(userID)
	return s.studentRepo.FetchPaidStudentsByMonth(instID, facultyID, month)
}

func (s *StudentService) FetchNotPaidStudentsService(userID uint, month string) ([]model.Student, error) {
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	instID, _ := s.userRepo.GetUserInstitutionID(userID)
	return s.studentRepo.FetchNotPaidStudentsByMonth(instID, facultyID, month)
}

func (s *StudentService) FetchFacultyPaidStudentsService(userID uint, facultyIDParam uint, month string) ([]model.Student, error) {
	var targetFacultyID uint
	if facultyIDParam > 0 {
		targetFacultyID = facultyIDParam
	} else {
		targetFacultyID, _ = s.userRepo.GetUserFacultyID(userID)
	}

	instID, _ := s.userRepo.GetUserInstitutionID(userID)
	return s.studentRepo.FetchPaidStudentsByMonth(instID, targetFacultyID, month)
}

func (s *StudentService) FetchFacultyUnpaidStudentsService(userID uint, facultyIDParam uint, month string) ([]model.Student, error) {
	var targetFacultyID uint
	if facultyIDParam > 0 {
		targetFacultyID = facultyIDParam
	} else {
		targetFacultyID, _ = s.userRepo.GetUserFacultyID(userID)
	}

	instID, _ := s.userRepo.GetUserInstitutionID(userID)
	return s.studentRepo.FetchNotPaidStudentsByMonth(instID, targetFacultyID, month)
}

func (s *StudentService) GetActiveStudentService() (model.Student, error) {
	return s.studentRepo.GetActiveStudent()
}

func (s *StudentService) GetInactiveStudentService() (model.Student, error) {
	return s.studentRepo.GetInactiveStudent()
}

func (s *StudentService) UpdateStudentService(
	userID uint,
	id uint,
	dto *dto.UpdateStudentDTO,
) error {

	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return err
	}

	institutionID, err := s.facultyRepo.GetInstitutionByFacultyID(
		student.FacultyID,
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
		student.Name = dto.Name
	}
	if dto.Gender != "" {
		student.Gender = dto.Gender
	}

	return s.studentRepo.UpdateStudentById(&student)
}

func (s *StudentService) DeleteStudentService(
	userID uint,
	id uint,
) error {

	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return err
	}

	institutionID, err := s.facultyRepo.GetInstitutionByFacultyID(
		student.FacultyID,
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

	return s.studentRepo.DeleteStudent(id)
}
