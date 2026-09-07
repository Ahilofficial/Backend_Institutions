package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"

)

type StudentService struct {
	studentRepo    *repository.StudentRepository
	facultyRepo    *repository.FacultyRepository
	userRepo       *repository.UserRepository
	departmentRepo *repository.DepartmentRepository
}

func NewStudentService(
	studentRepo *repository.StudentRepository,
	facultyRepo *repository.FacultyRepository,
	userRepo *repository.UserRepository,
	departmentRepo *repository.DepartmentRepository,
) *StudentService {
	return &StudentService{
		studentRepo:    studentRepo,
		facultyRepo:    facultyRepo,
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
	}
}

func (s *StudentService) GetUserFacultyID(userID uint) (uint, error) {
	return s.userRepo.GetUserFacultyID(userID)
}

func (s *StudentService) GetCourseDurationByFacultyID(facultyID uint) (uint, error) {
	return s.studentRepo.GetCourseDurationByFacultyIDRepo(facultyID)
}

func (s *StudentService) CreateStudentService(
	userID uint,
	createstudent *dto.CreateStudentDTO,
) (*model.Student, error) {

	if createstudent.MQ && createstudent.Scholorship {
		return nil, errors.New(
			"Student enrolled in Management Quota will not have scholarship",
		)
	}

	// Get faculty
	faculty, err := s.facultyRepo.FetchFacultyById(
		createstudent.FacultyID,
	)

	if err != nil || faculty.ID == 0 {
		return nil, errors.New("faculty not found")
	}

	// Get department from faculty
	dept, err := s.departmentRepo.FetchDepartmentById(
		faculty.DepartmentID,
	)

	if err != nil || dept.ID == 0 {
		return nil, errors.New("department not found")
	}

	// Validate semester
	if createstudent.Semester <= 0 ||
		createstudent.Semester > dept.CourseDuration*2 {

		return nil, errors.New(
			"semester exceeds course duration",
		)
	}

	// Check whether student profile already exists for this user (upsert if exists)
	existingStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if existingStudentID > 0 {
		existingStudent, err := s.studentRepo.FetchStudentById(existingStudentID)
		if err == nil && existingStudent.ID > 0 {
			existingStudent.Name = createstudent.Name
			existingStudent.Gender = createstudent.Gender
			existingStudent.Hosteller = createstudent.Hosteller
			existingStudent.MQ = createstudent.MQ
			existingStudent.Scholarship = createstudent.Scholorship
			existingStudent.Semester = createstudent.Semester
			existingStudent.FacultyID = createstudent.FacultyID
			existingStudent.DepartmentID = dept.ID

			if err := s.studentRepo.UpdateStudentById(&existingStudent); err != nil {
				return nil, err
			}
			return &existingStudent, nil
		}
	}

	// Create new student profile
	student := model.Student{
		Name:         createstudent.Name,
		Gender:       createstudent.Gender,
		IsActive:     true,
		Hosteller:    createstudent.Hosteller,
		MQ:           createstudent.MQ,
		Scholarship:  createstudent.Scholorship,
		Semester:     createstudent.Semester,
		UserID:       userID,
		FacultyID:    createstudent.FacultyID,
		DepartmentID: dept.ID,
	}

	// Save student
	if err := s.studentRepo.CreateStudent(&student); err != nil {
		return nil, err
	}

	// Link student to user
	if err := s.userRepo.UpdateUserStudentID(
		userID,
		student.ID,
	); err != nil {
		return nil, err
	}

	return &student, nil
}

// 



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
	
	
	return s.studentRepo.FetchStudentPaginated(search, page, limit)
}

func (s *StudentService) GetUserStudentIDService(userID uint) (uint, error) {
	return s.userRepo.GetUserStudentID(userID)
}


func (s *StudentService) GetStudentServiceById(
	// userID uint,
	id uint,
) (*model.Student, error) {

	student, err := s.studentRepo.FetchStudentById(id)
	return &student,err
	
}

// func (s *StudentService) GetActiveStudentService() (model.Student, error) {
// 	return s.studentRepo.GetActiveStudent()
// }

func (s *StudentService) UpdateStudentService(
	userID uint,
	id uint,
	req *dto.UpdateStudentDTO,
) (*model.Student, error) {
	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return nil, err
	}

	userStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if userStudentID > 0 && userStudentID != id {
		return nil, errors.New("access denied: you can only update your own profile")
	}

	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 && student.FacultyID != userFacultyID {
		return nil, errors.New("access denied: student does not belong to your faculty")
	}
	

	student.Name = req.Name
	student.Gender = req.Gender
	student.Semester=req.Semester

	if err := s.studentRepo.UpdateStudentById(&student); err != nil {
		return nil, err
	}
	return &student, nil
}

func (s *StudentService) GetInstitutionIDForUserService(studentID uint) uint {
	students_institution_id := s.studentRepo.GetInstitutionIDForUserRepo(studentID)
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

	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 && student.FacultyID == userFacultyID {
		return s.studentRepo.DeleteStudent(id)
	}
	return s.studentRepo.DeleteStudent(id)
}


func (s *StudentService) UpdateStudentSemesterControllerService(userID uint, id uint, dto *dto.UpdateSemesterDTO) (*model.Student, error) {
	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil || student.ID == 0 {
		return nil, errors.New("student not found")
	}

	dept, err := s.departmentRepo.FetchDepartmentById(student.DepartmentID)
	if err != nil || dept.ID == 0 {
		return nil, errors.New("department not found")
	}

	if dto.Semester > (dept.CourseDuration*2) || dto.Semester <= 0 {
		return nil, errors.New("semester exceeds course duration")
	}

	student.Semester = dto.Semester

	if err := s.studentRepo.UpdateStudentById(&student); err != nil {
		return nil, err
	}

	return &student, nil
}

