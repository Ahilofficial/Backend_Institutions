package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
	"fmt"
)

// FeesService provides business logic operations for fee structures and payment processing
type FeesService struct {
	feesRepo       *repository.FeesRepository
	studentRepo    *repository.StudentRepository
	userRepo       *repository.UserRepository
	departmentRepo *repository.DepartmentRepository
}

// NewFeesService initializes a new instance of FeesService with required repositories
func NewFeesService(
	feesRepo *repository.FeesRepository,
	studentRepo *repository.StudentRepository,
	userRepo *repository.UserRepository,
	departmentRepo *repository.DepartmentRepository,
) *FeesService {
	return &FeesService{
		feesRepo:       feesRepo,
		studentRepo:    studentRepo,
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
	}
}

// checkDepartmentAccess verifies whether a user is authorized to manage or view department fees
func (s *FeesService) checkDepartmentAccess(userID uint, departmentID uint) error {
	// 1. Verify user authentication
	if userID == 0 {
		return errors.New("unauthorized user")
	}

	// 2. Resolve institution associated with department
	instID, err := s.departmentRepo.GetInstitutionByDepartmentID(departmentID)
	if err != nil {
		return err
	}

	// 3. Verify user's access rights to this institution
	hasAccess, err := s.userRepo.HasInstitutionAccess(userID, instID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("access denied: department does not belong to your institution")
	}

	return nil
}

// CreateDepartmentFee configures a base fee template for a department and semester
func (s *FeesService) CreateDepartmentFee(
	userID uint,
	req *dto.CreateFeesDTO,
) (*model.Fees, error) {
	// 1. Check department authorization
	if err := s.checkDepartmentAccess(userID, req.DepartmentID); err != nil {
		return nil, err
	}

	// 2. Set default semester if omitted
	if req.Semester == 0 {
		req.Semester = 1
	}

	// 3. Check for existing fee configuration for this department/semester
	existing, _ := s.feesRepo.GetDepartmentFeeBySemester(req.DepartmentID, req.Semester)
	if existing != nil && existing.ID > 0 {
		return nil, fmt.Errorf("fee template already configured for department %d and semester %d", req.DepartmentID, req.Semester)
	}

	// 4. Calculate total amount
	totalAmount := req.HostelAmount + req.CollegeAmount

	// 5. Assemble department fee template model
	fee := model.Fees{
		DepartmentID:  req.DepartmentID,
		Semester:      req.Semester,
		StudentID:     nil,
		CollegeAmount: req.CollegeAmount,
		HostelAmount:  req.HostelAmount,
		TotalAmount:   totalAmount,
		PendingAmount: totalAmount,
		PaymentMode:   req.PaymentMode,
		IsActive:      true,
	}

	// 6. Save fee template in repository
	if err := s.feesRepo.CreateDepartmentFee(&fee); err != nil {
		return nil, err
	}

	return &fee, nil
}

// GetDepartmentFeeBySemester retrieves department fee configuration for a specific semester
func (s *FeesService) GetDepartmentFeeBySemester(
	userID uint,
	departmentID uint,
	semester uint,
) (*model.Fees, error) {
	// 1. Check department authorization
	if err := s.checkDepartmentAccess(userID, departmentID); err != nil {
		return nil, err
	}

	// 2. Fetch fee record by department and semester
	return s.feesRepo.GetDepartmentFeeBySemester(departmentID, semester)
}

// GetDepartmentFees retrieves all fee templates for a department
func (s *FeesService) GetDepartmentFees(
	userID uint,
	departmentID uint,
) ([]model.Fees, error) {
	// 1. Check department authorization
	if err := s.checkDepartmentAccess(userID, departmentID); err != nil {
		return nil, err
	}

	// 2. Fetch all department fees
	return s.feesRepo.GetDepartmentFees(departmentID)
}

// UpdateDepartmentFee updates amounts for a department fee template
func (s *FeesService) UpdateDepartmentFee(
	userID uint,
	id uint,
	req *dto.UpdateFeesDTO,
) (*model.Fees, error) {
	// 1. Fetch existing fee template
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil || fee.ID == 0 {
		return nil, errors.New("fee record not found")
	}

	// 2. Verify department authorization
	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		return nil, err
	}

	// 3. Update amounts
	if req.CollegeAmount > 0 {
		fee.CollegeAmount = req.CollegeAmount
	}
	if req.HostelAmount > 0 {
		fee.HostelAmount = req.HostelAmount
	}
	if req.TotalAmount > 0 {
		fee.TotalAmount = req.TotalAmount
	} else if req.CollegeAmount > 0 || req.HostelAmount > 0 {
		fee.TotalAmount = fee.CollegeAmount + fee.HostelAmount
	}

	// 4. Save updated fee configuration
	if err := s.feesRepo.UpdateDepartmentFee(&fee); err != nil {
		return nil, err
	}

	return &fee, nil
}

// DeleteDepartmentFee soft deletes a department fee template
func (s *FeesService) DeleteDepartmentFee(
	userID uint,
	id uint,
) error {
	// 1. Fetch fee record to verify existence
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil || fee.ID == 0 {
		return errors.New("fee record not found")
	}

	// 2. Check department authorization
	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		return err
	}

	// 3. Delete fee template
	return s.feesRepo.DeleteDepartmentFee(id)
}

// GetFeesService retrieves all fee records
func (s *FeesService) GetFeesService() ([]model.Fees, error) {
	return s.feesRepo.FetchFees()
}

// GetFeesServicePaginated retrieves paginated fee records with optional search
func (s *FeesService) GetFeesServicePaginated(
	userID uint,
	search string,
	page int,
	limit int,
) ([]model.Fees, int64, error) {
	return s.feesRepo.FetchFeesPaginated(search, page, limit)
}

// GetFeesByIDService fetches a single fee record by ID and enforces access checks
func (s *FeesService) GetFeesByIDService(userID uint, id uint) (model.Fees, error) {
	// 1. Fetch fee record
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return model.Fees{}, err
	}

	// 2. Verify department or student ownership access
	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		if userStudentID > 0 && fee.StudentID != nil && *fee.StudentID == userStudentID {
			return fee, nil
		}
		return model.Fees{}, err
	}

	return fee, nil
}

// GetLogginedStudentID returns the student ID for the authenticated user
func (s *FeesService) GetLogginedStudentID(userID uint) uint {
	studentID, _ := s.userRepo.GetUserStudentID(userID)
	return studentID
}

// FetchFeesByStudentID retrieves all semester fee records for a specific student
func (s *FeesService) FetchFeesByStudentID(userID uint, studentID uint) ([]model.Fees, error) {
	// 1. Fetch student record
	student, err := s.studentRepo.FetchStudentById(studentID)
	if err != nil || student.ID == 0 {
		return nil, errors.New("student not found")
	}

	// 2. Verify authorization
	if err := s.checkDepartmentAccess(userID, student.DepartmentID); err != nil {
		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		if userStudentID == 0 || userStudentID != studentID {
			return nil, errors.New("access denied")
		}
	}

	// 3. Fetch student fees
	return s.feesRepo.FetchFeesByStudentID(studentID)
}

// GetMyFees retrieves fee history and invoices for the logged-in student
func (s *FeesService) GetMyFees(userID uint) ([]model.Fees, error) {
	// 1. Resolve student ID from logged-in user
	studentID, err := s.userRepo.GetUserStudentID(userID)
	if err != nil || studentID == 0 {
		return nil, errors.New("user is not registered as a student")
	}

	// 2. Fetch fees for this student
	return s.feesRepo.FetchFeesByStudentID(studentID)
}

// CreatePayment processes a student fee payment, recalculates balance, and updates status
func (s *FeesService) CreatePayment(
	userID uint,
	dto *dto.CreatePaymentDTO,
) (model.Payment, error) {
	// 1. Resolve target student
	var targetStudentID uint
	if dto.StudentID > 0 {
		targetStudentID = dto.StudentID
	} else {
		targetStudentID, _ = s.userRepo.GetUserStudentID(userID)
	}

	var student model.Student
	if targetStudentID > 0 {
		student, _ = s.studentRepo.FetchStudentById(targetStudentID)
	}

	// 2. Resolve or locate fee record
	var fee model.Fees
	if dto.FeeID > 0 {
		f, err := s.feesRepo.FetchFeesById(dto.FeeID)
		if err == nil && f.ID > 0 {
			fee = f
		}
	}

	// 3. Fallback: auto-create fee record if missing for semester
	if fee.ID == 0 && targetStudentID > 0 {
		semester := dto.Semester
		if semester == 0 && student.Semester > 0 {
			semester = student.Semester
		}
		if semester == 0 {
			semester = 1
		}

		existingFee, _ := s.feesRepo.FetchFeeByStudentAndSemester(targetStudentID, semester)
		if existingFee != nil && existingFee.ID > 0 {
			fee = *existingFee
		} else {
			feeAmount := student.FeeAmount
			if feeAmount <= 0 {
				feeAmount = dto.AmountPaid
			}
			fee = model.Fees{
				DepartmentID:  student.DepartmentID,
				Semester:      semester,
				StudentID:     &targetStudentID,
				TotalAmount:   feeAmount,
				PendingAmount: feeAmount,
				TotalPaid:     0,
				IsActive:      true,
			}
			if err := s.feesRepo.CreateFees(&fee); err != nil {
				return model.Payment{}, fmt.Errorf("failed to create fee record: %w", err)
			}
		}
	}

	if fee.ID == 0 {
		return model.Payment{}, errors.New("fee record not found")
	}

	if fee.StudentID != nil && *fee.StudentID > 0 {
		targetStudentID = *fee.StudentID
		if student.ID == 0 {
			student, _ = s.studentRepo.FetchStudentById(targetStudentID)
		}
	}

	// 4. Verify department authorization
	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		if userStudentID == 0 || targetStudentID != userStudentID {
			return model.Payment{}, errors.New("access denied: cannot make payment for this fee record")
		}
	}

	// 5. Prevent double payment if fees already fully settled
	if (fee.PendingAmount == 0 && fee.TotalPaid > 0) || (targetStudentID > 0 && !student.Pending && fee.PendingAmount == 0 && fee.TotalPaid > 0) {
		return model.Payment{}, errors.New("you have already paid total amount")
	}

	// 6. Validate exact payment amount requirement
	expectedAmount := fee.PendingAmount
	if expectedAmount <= 0 && fee.TotalPaid == 0 {
		expectedAmount = fee.TotalAmount
	}
	if expectedAmount <= 0 && student.FeeAmount > 0 {
		expectedAmount = student.FeeAmount
	}

	if dto.AmountPaid != expectedAmount {
		return model.Payment{}, errors.New("need to pay exact amount")
	}
	userStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if dto.StudentID != userStudentID {
		return model.Payment{}, errors.New("cant pay fees for other student")
	}

	// 7. Create payment transaction record
	var studentIDPtr *uint
	if targetStudentID > 0 {
		studentIDPtr = &targetStudentID
	}
	payment := model.Payment{
		AmountPaid:  dto.AmountPaid,
		PaymentMode: dto.PaymentMode,
		FeeID:       fee.ID,
		StudentID:   studentIDPtr,
	}

	if err := s.feesRepo.CreatePayment(&payment); err != nil {
		return model.Payment{}, fmt.Errorf("failed to create payment: %w", err)
	}

	// 8. Recalculate fee balances and latest payment mode
	if err := s.feesRepo.RecalculateFeeTotals(fee.ID, dto.PaymentMode); err != nil {
		return model.Payment{}, fmt.Errorf("payment created but failed to recalculate totals: %w", err)
	}

	// 9. Update student payment ledger and mark student pending status as cleared
	if targetStudentID > 0 {
		semester := dto.Semester
		if semester == 0 {
			semester = fee.Semester
		}
		if semester == 0 && student.Semester > 0 {
			semester = student.Semester
		}
		if semester == 0 {
			semester = 1
		}

		studentPayment := model.StudentPayment{
			StudentID:   targetStudentID,
			PaymentID:   payment.ID,
			Semester:    semester,
			TotalAmount: dto.AmountPaid,
			Status:      "paid",
		}
		_ = s.studentRepo.UpsertStudentPayment(&studentPayment)
		_ = s.studentRepo.UpdateStudentPendingStatus(targetStudentID, false)
	}

	return payment, nil
}

// GetPaymentByID retrieves payment receipt by payment ID
func (s *FeesService) GetPaymentByID(userID uint, paymentID uint) (model.Payment, error) {
	payment, err := s.feesRepo.FetchPaymentByID(paymentID)
	if err != nil {
		return model.Payment{}, err
	}
	return payment, nil
}

// GetPaymentByFeeID retrieves all payment receipts for a fee ID
func (s *FeesService) GetPaymentByFeeID(userID uint, feeID uint) ([]model.Payment, error) {
	return s.feesRepo.FetchPaymentByFeeID(feeID)
}

// DeleteFees deletes fee record after verifying access
func (s *FeesService) DeleteFees(userID uint, id uint) error {
	// 1. Fetch fee record
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return err
	}

	// 2. Check department authorization
	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		return err
	}

	// 3. Delete fee record
	return s.feesRepo.DeleteFees(id)
}
