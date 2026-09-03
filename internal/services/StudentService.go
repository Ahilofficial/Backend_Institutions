package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
	"fmt"
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
	if createstudent.FacultyID == 0 {
		userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
		if userFacultyID > 0 {
			createstudent.FacultyID = userFacultyID
		} else {
			return nil, errors.New("faculty id is required")
		}
	}

	courseDuration, _ := s.studentRepo.GetCourseDurationByFacultyIDRepo(createstudent.FacultyID)
	if courseDuration > 0 && createstudent.Semester > courseDuration*2 {
		return nil, errors.New("This particular semester does not contain for the particular department")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.FacultyID != 0 {
		return nil, errors.New("user is already registered as a faculty")
	}
	if user.StudentID != 0 {
		return nil, errors.New("student profile already exists for this user")
	}

	faculty, err := s.facultyRepo.FetchFacultyById(createstudent.FacultyID)
	if err != nil || faculty.ID == 0 {
		return nil, errors.New("faculty not found")
	}

	targetSemester := createstudent.Semester
	if targetSemester == 0 {
		targetSemester = 1
	}

	if createstudent.MQ && createstudent.Scholorship {
		return nil, errors.New("management quota student cannot have scholarship")
	}

	deptFee, err := s.feesRepo.GetDepartmentFeeBySemester(faculty.DepartmentID, targetSemester)
	if err != nil || deptFee == nil {
		return nil, fmt.Errorf("fee template not configured for department ID %d and semester %d; please configure department fees first", faculty.DepartmentID, targetSemester)
	}

	collegeAmount := deptFee.CollegeAmount
	hostelAmount := deptFee.HostelAmount
	baseFee := deptFee.TotalAmount
	if baseFee <= 0 {
		baseFee = collegeAmount + hostelAmount
	}

	var finalTuitionFee float64
	var finalHostelFee float64

	if createstudent.MQ {
		finalTuitionFee = collegeAmount + (collegeAmount * 0.50)
		if createstudent.Hosteller {
			finalHostelFee = hostelAmount + (hostelAmount * 0.25)
		} else {
			finalHostelFee = 0
		}
	} else if createstudent.Scholorship {
		finalTuitionFee = collegeAmount - (collegeAmount * 0.25)
		if createstudent.Hosteller {
			finalHostelFee = hostelAmount
		} else {
			finalHostelFee = 0
		}
	} else {
		finalTuitionFee = collegeAmount
		if createstudent.Hosteller {
			finalHostelFee = hostelAmount
		} else {
			finalHostelFee = 0
		}
	}

	finalFeeAmount := finalTuitionFee + finalHostelFee

	student := model.Student{
		Name:         createstudent.Name,
		Gender:       createstudent.Gender,
		FacultyID:    createstudent.FacultyID,
		DepartmentID: faculty.DepartmentID,
		UserID:       userID,
		Semester:     targetSemester,
		Hosteller:    createstudent.Hosteller,
		Scholarship:  createstudent.Scholorship,
		MQ:           createstudent.MQ,
		BaseAmount:   baseFee,
		FeeAmount:    finalFeeAmount,
		IsActive:     true,
		Pending:      true,
	}

	initialFee := model.Fees{
		DepartmentID:  student.DepartmentID,
		Semester:      student.Semester,
		TotalAmount:   student.FeeAmount,
		PendingAmount: student.FeeAmount,
		TotalPaid:     0,
		IsActive:      true,
	}

	if err := s.studentRepo.CreateStudentWithFeeTx(&student, &initialFee); err != nil {
		return nil, err
	}

	_ = s.userRepo.UpdateUserStudentID(userID, student.ID)

	student.Fees = append(student.Fees, initialFee)
	return &student, nil
}


func (s *StudentService) StudentVerification(studentID uint, facultyID uint) (model.StudentVerificationAccess, error) {
	stu_verify := model.StudentVerificationAccess{
		StudentID: studentID,
		FacultyID: facultyID,
	}
	err := s.studentRepo.StudentVerificationRepo(stu_verify)
	if err != nil {
		return model.StudentVerificationAccess{}, err
	}

	return stu_verify, nil
}

func (s *StudentService) UpdateStudentVerified(
	userID uint,
	studentID uint,
) error {

	facultyID, err := s.userRepo.GetUserFacultyID(userID)

	if err != nil {
		return errors.New("faculty not found")
	}

	var access model.StudentVerificationAccess

	err = s.studentRepo.GetStudentVerificationAccess(
		studentID,
		facultyID,
		&access,
	)

	if err != nil {
		return errors.New(
			"you must view the student before verification",
		)
	}

	err = s.studentRepo.UpdateStudentVerified(studentID)

	if err != nil {
		return err
	}

	return nil
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

func(s *StudentService) GetUserStudentIDService(userID uint) (uint, error) {
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
	if err == nil && userStudentID > 0 {
		if userStudentID != id {
			return nil, errors.New("access denied: you can only access your own student profile")
		}
		if !student.IsProfileVerified {
			return nil, errors.New("Faculty first need to verify your profile")
		}
		return &student, nil
	}

	userFacultyID, err := s.userRepo.GetUserFacultyID(userID)
	if err == nil && userFacultyID > 0 {
		if student.FacultyID != userFacultyID {
			return nil, errors.New("access denied: student does not belong to your faculty")
		}
		_, _ = s.StudentVerification(id, userFacultyID)
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



func (s *StudentService) GetActiveStudentService() (model.Student, error) {
	return s.studentRepo.GetActiveStudent()
}



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

func (s *StudentService) PromoteStudentService(
	userID uint,
	studentID uint,
) (*model.Student, *model.Fees, error) {
	student, err := s.studentRepo.FetchStudentById(studentID)
	if err != nil || student.ID == 0 {
		return nil, nil, errors.New("student not found")
	}

	if !student.IsActive {
		return nil, nil, errors.New("cannot promote an inactive student")
	}

	dept, err := s.departmentRepo.FetchDepartmentById(student.DepartmentID)
	if err == nil && dept.CourseDuration > 0 && student.Semester >= dept.CourseDuration {
		return nil, nil, fmt.Errorf("student has already completed the maximum course duration (%d semesters)", dept.CourseDuration)
	}

	nextSemester := student.Semester + 1

	existingFee, _ := s.feesRepo.FetchFeeByStudentAndSemester(student.ID, nextSemester)
	if existingFee != nil && existingFee.ID > 0 {
		return nil, nil, fmt.Errorf("fee record for semester %d already exists for this student", nextSemester)
	}

	deptFee, err := s.feesRepo.GetDepartmentFeeBySemester(student.DepartmentID, nextSemester)
	if err != nil || deptFee == nil {
		return nil, nil, fmt.Errorf("fee template not found for department ID %d and semester %d; please configure department fees first", student.DepartmentID, nextSemester)
	}

	collegeAmount := deptFee.CollegeAmount
	hostelAmount := deptFee.HostelAmount
	baseFee := deptFee.TotalAmount
	if baseFee <= 0 {
		baseFee = collegeAmount + hostelAmount
	}

	var finalTuitionFee float64
	var finalHostelFee float64

	if student.MQ {
		finalTuitionFee = collegeAmount + (collegeAmount * 0.50)
		if student.Hosteller {
			finalHostelFee = hostelAmount + (hostelAmount * 0.25)
		} else {
			finalHostelFee = 0
		}
	} else if student.Scholarship {
		finalTuitionFee = collegeAmount - (collegeAmount * 0.25)
		if student.Hosteller {
			finalHostelFee = hostelAmount
		} else {
			finalHostelFee = 0
		}
	} else {
		finalTuitionFee = collegeAmount
		if student.Hosteller {
			finalHostelFee = hostelAmount
		} else {
			finalHostelFee = 0
		}
	}

	newFeeAmount := finalTuitionFee + finalHostelFee

	newFee := model.Fees{
		DepartmentID:  student.DepartmentID,
		Semester:      nextSemester,
		TotalAmount:   newFeeAmount,
		PendingAmount: newFeeAmount,
		TotalPaid:     0,
		IsActive:      true,
	}


	student.Semester = nextSemester
	student.BaseAmount = baseFee
	student.FeeAmount = newFeeAmount

	return &student, &newFee, nil
}

func (s *StudentService) UpdateStudentSemesterService(
	userID uint,
	studentID uint,
	req *dto.UpdateStudentSemesterDTO,
) (*model.Student, *model.Fees, error) {
	student, err := s.studentRepo.FetchStudentById(studentID)
	if err != nil || student.ID == 0 {
		return nil, nil, errors.New("student not found")
	}

	if !student.IsActive {
		return nil, nil, errors.New("cannot update an inactive student")
	}

	userStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if userStudentID > 0 && userStudentID != studentID {
		return nil, nil, errors.New("access denied: you can only update your own semester details")
	}

	userFacultyID, _ := s.userRepo.GetUserFacultyID(userID)
	isInstAdmin, assignedInstID, _ := s.userRepo.IsInstitutionAdmin(userID)
	isSuper, _ := s.userRepo.IsSuperAdmin(userID)

	if userStudentID == 0 && !isSuper {
		if isInstAdmin {
			studentInstID, _ := s.studentRepo.GetInstitutionIDByStudent(studentID)
			if assignedInstID == 0 || studentInstID != assignedInstID {
				return nil, nil, errors.New("access denied: student does not belong to your institution")
			}
		} else if userFacultyID > 0 {
			if student.FacultyID != userFacultyID {
				return nil, nil, errors.New("access denied: student does not belong to your faculty")
			}
		} else if student.UserID != userID {
			return nil, nil, errors.New("access denied: unauthorized to update this student")
		}
	}

	dept, err := s.departmentRepo.FetchDepartmentById(student.DepartmentID)
	if err == nil && dept.CourseDuration > 0 && req.Semester > dept.CourseDuration {
		return nil, nil, fmt.Errorf("semester %d exceeds maximum course duration (%d semesters)", req.Semester, dept.CourseDuration)
	}

	if req.Hosteller != nil {
		student.Hosteller = *req.Hosteller
	}
	if req.Scholarship != nil {
		student.Scholarship = *req.Scholarship
	}
	if req.MQ != nil {
		student.MQ = *req.MQ
	}

	targetSemester := req.Semester
	if targetSemester == 0 {
		targetSemester = student.Semester
	}

	deptFee, err := s.feesRepo.GetDepartmentFeeBySemester(student.DepartmentID, targetSemester)
	if err != nil || deptFee == nil {
		return nil, nil, fmt.Errorf("fee template not found for department ID %d and semester %d; please configure department fees first", student.DepartmentID, targetSemester)
	}

	collegeAmount := deptFee.CollegeAmount
	hostelAmount := deptFee.HostelAmount
	baseFee := deptFee.TotalAmount
	if baseFee <= 0 {
		baseFee = collegeAmount + hostelAmount
	}

	var finalTuitionFee float64
	var finalHostelFee float64

	if student.MQ {
		finalTuitionFee = collegeAmount + (collegeAmount * 0.50)
		if student.Hosteller {
			finalHostelFee = hostelAmount + (hostelAmount * 0.25)
		} else {
			finalHostelFee = 0
		}
	} else if student.Scholarship {
		finalTuitionFee = collegeAmount - (collegeAmount * 0.25)
		if student.Hosteller {
			finalHostelFee = hostelAmount
		} else {
			finalHostelFee = 0
		}
	} else {
		finalTuitionFee = collegeAmount
		if student.Hosteller {
			finalHostelFee = hostelAmount
		} else {
			finalHostelFee = 0
		}
	}

	newFeeAmount := finalTuitionFee + finalHostelFee

	existingFee, _ := s.feesRepo.FetchFeeByStudentAndSemester(student.ID, targetSemester)
	var feeToSave model.Fees
	if existingFee != nil && existingFee.ID > 0 {
		feeToSave = *existingFee
		feeToSave.TotalAmount = newFeeAmount
		feeToSave.PendingAmount = newFeeAmount - feeToSave.TotalPaid
		if feeToSave.PendingAmount < 0 {
			feeToSave.PendingAmount = 0
		}
	} else {
		feeToSave = model.Fees{
			DepartmentID:  student.DepartmentID,
			Semester:      targetSemester,
			StudentID:     &student.ID,
			TotalAmount:   newFeeAmount,
			PendingAmount: newFeeAmount,
			TotalPaid:     0,
			IsActive:      true,
		}
	}

	if err := s.studentRepo.UpdateStudentSemesterTx(student.ID, targetSemester, student.Hosteller, student.Scholarship, student.MQ, baseFee, newFeeAmount, &feeToSave); err != nil {
		return nil, nil, err
	}

	student.Semester = targetSemester
	student.BaseAmount = baseFee
	student.FeeAmount = newFeeAmount
	student.Pending = feeToSave.PendingAmount > 0

	return &student, &feeToSave, nil
}
