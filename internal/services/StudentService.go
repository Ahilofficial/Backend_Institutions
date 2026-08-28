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



func (s *StudentService) GetUserFacultyID(userID uint) (uint, error) {
	return s.userRepo.GetUserFacultyID(userID)
}

func (s *StudentService) GetUserExistingProfile(userID uint) (string, error) {
	return s.userRepo.CheckUserExistingProfile(userID)
}

func (s *StudentService) CreateStudentService(
	userID uint,
	student *model.Student,
	createstudent *dto.CreateStudentDTO,
) (*model.Student, error) {
	faculty, err := s.facultyRepo.FetchFacultyById(student.FacultyID)
	if err != nil || faculty.ID == 0 {
		return nil, errors.New("faculty not found")
	}

	targetUserID := userID
	if student.UserID > 0 {
		targetUserID = student.UserID
	}

	if targetUserID > 0 {
		if err := s.userRepo.ValidateUser(targetUserID); err != nil {
			return nil, err
		}

		existingType, err := s.userRepo.CheckUserExistingProfile(targetUserID)
		if err == nil && existingType != "" {
			return nil, errors.New("user is already registered as a " + existingType)
		}
	}
	

	student.UserID = targetUserID
	if err := s.studentRepo.CreateStudent(student); err != nil {
		return nil, err
	}

	
	if createstudent.Rank<=100{
		student.Scholorship=true
	}
	if createstudent.Scholorship==true && createstudent.Rank<100{
		student.FeeAmount = student.FeeAmount - (student.FeeAmount * 25 / 100)
	}
	if createstudent.Scholorship==true && createstudent.Rank<50{
		student.FeeAmount = student.FeeAmount - (student.FeeAmount * 25 / 100)
	}
	if createstudent.Rank>100{
		student.MQ=true
	}
	if student.MQ==true{
		student.FeeAmount=student.FeeAmount+(student.FeeAmount*50/100)
	}
	var model model.Payment
	err:=c.Bind().Body(&model);err!=nil{
		return nil, nil
	}
	return student, nil
	}



	


func (s *StudentService) FetchAllStudentsServices() ([]model.Student, error) {
	return s.studentRepo.FetchStudent()
}

func (s *StudentService) FetchAllStudentsServicesScoped(userID uint) ([]model.Student, error) {
	instID, err := s.resolveInstitutionScope(userID, 0)
	if err != nil {
		return nil, err
	}
	if instID > 0 {
		return s.studentRepo.FetchStudentByInstitution(instID)
	}
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

func (s *StudentService) FetchAllStudentsPaginatedServicesScoped(
	userID uint,
	search string,
	page int,
	limit int,
) ([]model.Student, int64, error) {
	instID, err := s.resolveInstitutionScope(userID, 0)
	if err != nil {
		return nil, 0, err
	}
	if instID > 0 {
		return s.studentRepo.FetchStudentPaginatedWithInstitution(search, page, limit, instID)
	}
	return s.studentRepo.FetchStudentPaginated(search, page, limit)
}

func (s *StudentService) GetStudentServiceById(
	userID uint,
	id uint,
) (*model.Student, error) {

	// Get logged-in user's role ID
	roleID, err := s.userRepo.GetUserRoleID(userID)
	if err != nil {
		return nil, err
	}

	// Get requested student
	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return nil, err
	}

	// Check whether this role is a SuperAdmin role
	isSuper, err := s.userRepo.IsSuperAdminByRoleID(roleID)
	if err != nil {
		return nil, err
	}

	if isSuper {
		return &student, nil
	}

	// Check whether this role is assigned as an Institution Admin
	institutionID, err := s.userRepo.GetInstitutionAdminID(userID)
	if err != nil {
		return nil, err
	}

	if institutionID != 0 {

		studentInstitutionID, err :=
			s.facultyRepo.GetInstitutionByFacultyID(student.FacultyID)
		if err != nil {
			return nil, err
		}

		if institutionID != studentInstitutionID {
			return nil, errors.New(
				"access denied: student does not belong to your institution",
			)
		}

		return &student, nil
	}

	
	userStudentID, err := s.userRepo.GetUserStudentID(userID)
	if err != nil {
		return nil, err
	}

	if userStudentID != 0 {

		if userStudentID != id {
			return nil, errors.New(
				"access denied: you can only access your own student profile",
			)
		}

		return &student, nil
	}

	// Check whether logged-in user has a Faculty profile
	userFacultyID, err := s.userRepo.GetUserFacultyID(userID)
	if err != nil {
		return nil, err
	}

	if userFacultyID != 0 {

		if student.FacultyID != userFacultyID {
			return nil, errors.New(
				"access denied: student does not belong to your faculty",
			)
		}

		return &student, nil
	}

	return nil, errors.New("access denied: profile not created yet")
}

func (s *StudentService) GetLoggedInStudentProfile(userID uint) (*model.Student, error) {
	studentID, err := s.userRepo.GetUserStudentID(userID)
	if err != nil || studentID == 0 {
		return nil, errors.New("student profile not created yet for logged in user")
	}
	return s.GetStudentServiceById(userID, studentID)
}

func (s *StudentService) resolveInstitutionScope(userID uint, requestedInstID uint) (uint, error) {
	isInstAdmin, assignedInstID, _ := s.userRepo.IsInstitutionAdmin(userID)
	if isInstAdmin {
		if assignedInstID == 0 {
			return 0, errors.New("cant able to access other institution")
		}
		if requestedInstID > 0 && requestedInstID != assignedInstID {
			return 0, errors.New("cant able to access other institution")
		}
		return assignedInstID, nil
	}

	isSuper, _ := s.userRepo.IsSuperAdmin(userID)
	if isSuper {
		return requestedInstID, nil
	}

	userInstID, _ := s.userRepo.GetUserInstitutionID(userID)
	if userInstID > 0 {
		if requestedInstID > 0 && requestedInstID != userInstID {
			return 0, errors.New("cant able to access other institution")
		}
		return userInstID, nil
	}

	return 0, nil
}

func (s *StudentService) FetchStudentsByPaymentMonthService(
	userID uint,
	month string,
) ([]model.Student, error) {
	return s.FetchPaidStudentsService(userID, month)
}

func (s *StudentService) FetchStudentsNotPaidByMonthService(
	userID uint,
	month string,
) ([]model.Student, error) {
	return s.FetchNotPaidStudentsService(userID, month)
}

func (s *StudentService) FetchPaidStudentsService(userID uint, month string) ([]model.Student, error) {
	instID, err := s.resolveInstitutionScope(userID, 0)
	if err != nil {
		return nil, err
	}
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	return s.studentRepo.FetchPaidStudentsByMonth(instID, facultyID, month)
}

func (s *StudentService) FetchNotPaidStudentsService(userID uint, month string) ([]model.Student, error) {
	instID, err := s.resolveInstitutionScope(userID, 0)
	if err != nil {
		return nil, err
	}
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	return s.studentRepo.FetchNotPaidStudentsByMonth(instID, facultyID, month)
}

func (s *StudentService) FetchAllStudentsMonthOverviewService(
	userID uint,
	requestedInstID uint,
	month string,
) (dto.MonthlyStudentsOverviewDTO, error) {
	instID, err := s.resolveInstitutionScope(userID, requestedInstID)
	if err != nil {
		return dto.MonthlyStudentsOverviewDTO{}, err
	}
	facultyID, _ := s.userRepo.GetUserFacultyID(userID)
	return s.studentRepo.FetchAllStudentsMonthOverview(instID, facultyID, month)
}

func (s *StudentService) FetchFacultyPaidStudentsService(userID uint, facultyIDParam uint, month string) ([]model.Student, error) {
	var targetFacultyID uint
	if facultyIDParam > 0 {
		targetFacultyID = facultyIDParam
	} else {
		targetFacultyID, _ = s.userRepo.GetUserFacultyID(userID)
	}

	instID, err := s.resolveInstitutionScope(userID, 0)
	if err != nil {
		return nil, err
	}
	return s.studentRepo.FetchPaidStudentsByMonth(instID, targetFacultyID, month)
}

func (s *StudentService) FetchFacultyUnpaidStudentsService(userID uint, facultyIDParam uint, month string) ([]model.Student, error) {
	var targetFacultyID uint
	if facultyIDParam > 0 {
		targetFacultyID = facultyIDParam
	} else {
		targetFacultyID, _ = s.userRepo.GetUserFacultyID(userID)
	}

	instID, err := s.resolveInstitutionScope(userID, 0)
	if err != nil {
		return nil, err
	}
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
	req *dto.UpdateStudentDTO,
) error {
	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return err
	}

	// 1. If student updating their own record
	userStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if userStudentID > 0 && userStudentID == id {
		student.Name = req.Name
		student.Gender = req.Gender
		return s.studentRepo.UpdateStudentById(&student)
	}

	// 2. If faculty mentor updating their student
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 && student.FacultyID == userFacultyID {
		student.Name = req.Name
		student.Gender = req.Gender
		return s.studentRepo.UpdateStudentById(&student)
	}

	

	

	student.Name = req.Name
	student.Gender = req.Gender
	return s.studentRepo.UpdateStudentById(&student)
}

func (s *StudentService)GetInstitutionIDForUserService(studentID uint)uint{
	students_institution_id:=s.studentRepo.GetInstitutionIDForUserRepo(studentID)
	return students_institution_id
}

func (s *StudentService) DeleteStudentService(
	userID uint,
	id uint,
) error {
	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return err
	}

	// 1. Faculty mentor deleting their student
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 && student.FacultyID == userFacultyID {
		return s.studentRepo.DeleteStudent(id)
	}
	return s.studentRepo.DeleteStudent(id)
}