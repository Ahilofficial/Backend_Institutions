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
	feesRepo       *repository.FeesRepository
}

func NewStudentService(
	studentRepo *repository.StudentRepository,
	facultyRepo *repository.FacultyRepository,
	userRepo *repository.UserRepository,
	departmentRepo *repository.DepartmentRepository,
	feesRepo *repository.FeesRepository,
) *StudentService {
	return &StudentService{
		studentRepo:    studentRepo,
		facultyRepo:    facultyRepo,
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
		feesRepo:       feesRepo,
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

	if createstudent.MQ == true && createstudent.Scholorship == true {
		return &model.Student{}, errors.New("Student enrolled in Management Quota will not have scholorship")
	}

	StudentID, _ := s.userRepo.GetUserStudentID(userID)

	fetchStudentDepartment, _ := s.studentRepo.GetStudentDepartment(userID)
	if createstudent.Semester > (fetchStudentDepartment.CourseDuration*2) || createstudent.Semester <= 0 {
		return &model.Student{}, errors.New("Semester ID is not present in the institution")
	}
	student_fees, _ := s.feesRepo.FetchFeesByStudentID(StudentID)
	baseamount := student_fees.CollegeAmount + student_fees.HostelAmount
	altered_hostelAmount := student_fees.HostelAmount
	altered_CollegeAmount := student_fees.CollegeAmount
	fee_Amount := altered_CollegeAmount + altered_hostelAmount

	if createstudent.MQ == true {
		altered_hostelAmount = altered_hostelAmount + (altered_hostelAmount * 0.25)
		altered_CollegeAmount = altered_CollegeAmount + (altered_CollegeAmount * 0.50)
	}
	if createstudent.Scholorship == true {
		altered_CollegeAmount = altered_CollegeAmount - (altered_CollegeAmount * 0.25)
		altered_hostelAmount = altered_hostelAmount - (altered_hostelAmount * 0.25)
	}
	fee_Amount = altered_CollegeAmount + altered_hostelAmount
	student := model.Student{
		Name:         createstudent.Name,
		Gender:       createstudent.Gender,
		IsActive:     true,
		Hosteller:    createstudent.Hosteller,
		MQ:           createstudent.MQ,
		Scholarship:  createstudent.Scholorship,
		BaseAmount:   baseamount,
		FeeAmount:    fee_Amount,
		Semester:     createstudent.Semester,
		Pending:      true,
		FacultyID:    createstudent.FacultyID,
		DepartmentID: fetchStudentDepartment.ID,
	}
	_=s.studentRepo.CreateStudent(&student)
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
	userID uint,
	id uint,
) (*model.Student, error) {
	student, err := s.studentRepo.FetchStudentById(id)
	if err != nil {
		return nil, err
	}

	// 1. If user is a faculty member, verify student belongs to this faculty
	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	if userFacultyID > 0 && student.FacultyID != userFacultyID {
		return nil, errors.New("access denied: student does not belong to your faculty")
	}

	// 2. If user is a student, verify student is accessing their own profile
	userStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if userStudentID > 0 && userStudentID != id {
		return nil, errors.New("access denied: you can only view your own profile")
	}

	return &student, nil
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
	if userFacultyID > 0 && student.FacultyID != userFacultyID {
		return errors.New("access denied: student does not belong to your faculty")
	}

	return s.studentRepo.DeleteStudent(id)
}


func(s *StudentService)UpdateStudentSemesterControllerService(userID uint, id uint,dto *dto.UpdateSemesterDTO)(model.Student,error){

	fetchStudentDepartment, _ := s.studentRepo.GetStudentDepartment(userID)
	if dto.Semester > (fetchStudentDepartment.CourseDuration*2) || dto.Semester <= 0 {
		return nil, errors.New("Semester ID is not present in the institution")
	}
	fetch_students,_:=s.studentRepo.FetchStudentById(id)
	student_dept_id:=fetch_students.DepartmentID
	student_sem:=dto.Semester
	student:=model.Student{
		
	}
	

	cl,_:=s.studentRepo.UpdateStudentSemesterControllerRepo(student)(model.Student)
}

