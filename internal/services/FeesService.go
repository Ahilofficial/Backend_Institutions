package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
	"fmt"
)

type FeesService struct {
	feesRepo       *repository.FeesRepository
	studentRepo    *repository.StudentRepository
	userRepo       *repository.UserRepository
	departmentRepo *repository.DepartmentRepository
}

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

func (s *FeesService) checkDepartmentAccess(userID uint, departmentID uint) error {
	if userID == 0 {
		return errors.New("unauthorized user")
	}

	instID, err := s.departmentRepo.GetInstitutionByDepartmentID(departmentID)
	if err != nil {
		return err
	}

	hasAccess, err := s.userRepo.HasInstitutionAccess(userID, instID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return errors.New("access denied: department does not belong to your institution")
	}

	return nil
}

func (s *FeesService) CreateDepartmentFee(
	userID uint,
	req *dto.CreateFeesDTO,
) (*model.Fees, error) {
	if err := s.checkDepartmentAccess(userID, req.DepartmentID); err != nil {
		return nil, err
	}

	if req.Semester == 0 {
		req.Semester = 1
	}

	existing, _ := s.feesRepo.GetDepartmentFeeBySemester(req.DepartmentID, req.Semester)
	if existing != nil && existing.ID > 0 {
		return nil, fmt.Errorf("fee template already configured for department %d and semester %d", req.DepartmentID, req.Semester)
	}

	totalAmount := req.Amount
	if totalAmount == 0 && (req.CollegeAmount > 0 || req.HostelAmount > 0) {
		totalAmount = req.CollegeAmount + req.HostelAmount
	}

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

	if err := s.feesRepo.CreateDepartmentFee(&fee); err != nil {
		return nil, err
	}

	return &fee, nil
}

func (s *FeesService) GetDepartmentFeeBySemester(
	userID uint,
	departmentID uint,
	semester uint,
) (*model.Fees, error) {
	if err := s.checkDepartmentAccess(userID, departmentID); err != nil {
		return nil, err
	}

	return s.feesRepo.GetDepartmentFeeBySemester(departmentID, semester)
}

func (s *FeesService) GetDepartmentFees(
	userID uint,
	departmentID uint,
) ([]model.Fees, error) {
	if err := s.checkDepartmentAccess(userID, departmentID); err != nil {
		return nil, err
	}

	return s.feesRepo.GetDepartmentFees(departmentID)
}

func (s *FeesService) UpdateDepartmentFee(
	userID uint,
	id uint,
	req *dto.UpdateFeesDTO,
) (*model.Fees, error) {
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil || fee.ID == 0 {
		return nil, errors.New("fee record not found")
	}

	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		return nil, err
	}

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

	if err := s.feesRepo.UpdateDepartmentFee(&fee); err != nil {
		return nil, err
	}

	return &fee, nil
}

func (s *FeesService) DeleteDepartmentFee(
	userID uint,
	id uint,
) error {
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil || fee.ID == 0 {
		return errors.New("fee record not found")
	}

	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		return err
	}

	return s.feesRepo.DeleteDepartmentFee(id)
}

func (s *FeesService) GetFeesService() ([]model.Fees, error) {
	return s.feesRepo.FetchFees()
}

func (s *FeesService) GetFeesServicePaginated(
	userID uint,
	search string,
	page int,
	limit int,
) ([]model.Fees, int64, error) {
	return s.feesRepo.FetchFeesPaginated(search, page, limit)
}

func (s *FeesService) GetFeesByIDService(userID uint, id uint) (model.Fees, error) {
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return model.Fees{}, err
	}

	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {

		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		if userStudentID > 0 && fee.StudentID != nil && *fee.StudentID == userStudentID {
			return fee, nil
		}
		return model.Fees{}, err
	}

	return fee, nil
}
func(s *FeesService) GetLogginedStudentID(userID uint) uint {
	studentID, _ := s.userRepo.GetUserStudentID(userID)
	return studentID
}

func (s *FeesService) FetchFeesByStudentID(userID uint, studentID uint) ([]model.Fees, error) {
	student, err := s.studentRepo.FetchStudentById(studentID)
	if err != nil || student.ID == 0 {
		return nil, errors.New("student not found")
	}

	if err := s.checkDepartmentAccess(userID, student.DepartmentID); err != nil {
		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		if userStudentID == 0 || userStudentID != studentID {
			return nil, errors.New("access denied")
		}
	}

	return s.feesRepo.FetchFeesByStudentID(studentID)
}

func (s *FeesService) GetMyFees(userID uint) ([]model.Fees, error) {
	studentID, err := s.userRepo.GetUserStudentID(userID)
	if err != nil || studentID == 0 {
		return nil, errors.New("user is not registered as a student")
	}
	return s.feesRepo.FetchFeesByStudentID(studentID)
}

func (s *FeesService) CreatePayment(
	userID uint,
	dto *dto.CreatePaymentDTO,
) (model.Payment, error) {
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

	var fee model.Fees
	if dto.FeeID > 0 {
		f, err := s.feesRepo.FetchFeesById(dto.FeeID)
		if err == nil && f.ID > 0 {
			fee = f
		}
	}

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

	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		userStudentID, _ := s.userRepo.GetUserStudentID(userID)
		if userStudentID == 0 || targetStudentID != userStudentID {
			return model.Payment{}, errors.New("access denied: cannot make payment for this fee record")
		}
	}

	if (fee.PendingAmount == 0 && fee.TotalPaid > 0) || (targetStudentID > 0 && !student.Pending && fee.PendingAmount == 0 && fee.TotalPaid > 0) {
		return model.Payment{}, errors.New("you have already paid total amount")
	}

	expectedAmount := fee.PendingAmount
	if expectedAmount <= 0 && fee.TotalPaid == 0 {
		expectedAmount = fee.TotalAmount
	}
	if expectedAmount <= 0 && student.FeeAmount > 0 {
		expectedAmount = student.FeeAmount
	}
	userStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if userStudentID!=dto.StudentID{
		return model.Payment{}, errors.New("cant pay fees for other student")
	}

	if (dto.AmountPaid != expectedAmount) {
		return model.Payment{}, errors.New("need to pay exact amount")
	}
	
	
	

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

	if err := s.feesRepo.RecalculateFeeTotals(fee.ID, dto.PaymentMode); err != nil {
		return model.Payment{}, fmt.Errorf("payment created but failed to recalculate totals: %w", err)
	}

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

func (s *FeesService) GetPaymentByID(userID uint, paymentID uint) (model.Payment, error) {
	payment, err := s.feesRepo.FetchPaymentByID(paymentID)
	if err != nil {
		return model.Payment{}, err
	}
	return payment, nil
}

func (s *FeesService) GetPaymentByFeeID(userID uint, feeID uint) ([]model.Payment, error) {
	return s.feesRepo.FetchPaymentByFeeID(feeID)
}

func (s *FeesService) DeleteFees(userID uint, id uint) error {
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return err
	}

	if err := s.checkDepartmentAccess(userID, fee.DepartmentID); err != nil {
		return err
	}

	return s.feesRepo.DeleteFees(id)
}
