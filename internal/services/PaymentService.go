package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
	"fmt"
)

type PaymentService struct {
	paymentRepo    *repository.PaymentRepository
	instituteRepo  *repository.InstitutionRepository
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
}

func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	instituteRepo *repository.InstitutionRepository,
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo:    paymentRepo,
		instituteRepo:  instituteRepo,
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

func (s *PaymentService) checkInstitutionAdminAccess(userID uint, departmentID uint) error {
	if userID == 0 {
		return errors.New("unauthorized user")
	}

	isInstAdmin := s.instituteRepo.IsInstAdminRepo(userID)
	if !isInstAdmin {
		return errors.New("only institution admin can perform this action")
	}

	adminInstID := s.instituteRepo.GetInstitutionIDForUserRepo(userID)
	if adminInstID == 0 {
		return errors.New("admin is not assigned to any institution")
	}

	deptInstID, err := s.departmentRepo.GetInstitutionByDepartmentID(departmentID)
	if err != nil || deptInstID == 0 {
		return errors.New("department not found or has invalid institution")
	}

	if adminInstID != deptInstID {
		return errors.New("cant able to manage payment for other institution department")
	}

	return nil
}

func (s *PaymentService) CreateDepartmentPayment(
	userID uint,
	req *dto.CreateDepartmentPaymentDTO,
) (*model.DepartmentPayment, error) {

	if err := s.checkInstitutionAdminAccess(userID, req.DepartmentID); err != nil {
		return nil, err
	}

	existing, _ := s.paymentRepo.GetDepartmentPaymentBySemester(req.DepartmentID, req.Semester)
	if existing != nil && existing.ID > 0 {
		return nil, fmt.Errorf("payment already configured for department %d and semester %d", req.DepartmentID, req.Semester)
	}

	totalAmount := req.CollegeAmount + req.HostelAmount

	payment := model.DepartmentPayment{
		DepartmentID:  req.DepartmentID,
		Semester:      req.Semester,
		CollegeAmount: req.CollegeAmount,
		HostelAmount:  req.HostelAmount,
		TotalAmount:   totalAmount,
	}

	if err := s.paymentRepo.CreateDepartmentPayment(&payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (s *PaymentService) UpdateDepartmentPayment(
	userID uint,
	id uint,
	req *dto.UpdateDepartmentPaymentDTO,
) (*model.DepartmentPayment, error) {
	payment, err := s.paymentRepo.GetDepartmentPaymentByID(id)
	if err != nil || payment == nil || payment.ID == 0 {
		return nil, errors.New("payment configuration not found")
	}

	if err := s.checkInstitutionAdminAccess(userID, payment.DepartmentID); err != nil {
		return nil, err
	}

	payment.CollegeAmount = req.CollegeAmount
	payment.HostelAmount = req.HostelAmount
	payment.TotalAmount = payment.CollegeAmount + payment.HostelAmount

	if err := s.paymentRepo.UpdateDepartmentPayment(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) DeleteDepartmentPayment(userID uint, id uint) error {
	payment, err := s.paymentRepo.GetDepartmentPaymentByID(id)
	if err != nil || payment == nil || payment.ID == 0 {
		return errors.New("payment configuration not found")
	}

	if err := s.checkInstitutionAdminAccess(userID, payment.DepartmentID); err != nil {
		return err
	}

	return s.paymentRepo.DeleteDepartmentPayment(id)
}

func (s *PaymentService) GetDepartmentPaymentBySemester(deptID uint, semester uint) (*model.DepartmentPayment, error) {
	return s.paymentRepo.GetDepartmentPaymentBySemester(deptID, semester)
}

func (s *PaymentService) GetDepartmentPayments(deptID uint) ([]model.DepartmentPayment, error) {
	return s.paymentRepo.GetDepartmentPaymentsByDeptID(deptID)
}

func (s *PaymentService) MakeStudentPayment(
	userID uint,
	req *dto.CreateStudentPaymentDTO,

) (*model.StudentPayment, *model.Student, error) {

	loggedInStudentID, _ := s.userRepo.GetUserStudentID(userID)
	if req.StudentID == 0 && loggedInStudentID > 0 {
		req.StudentID = loggedInStudentID
	}
	if loggedInStudentID > 0 && req.StudentID != loggedInStudentID {
		return nil, nil, errors.New("cant able to make payment for other student")
	}

	student, err := s.paymentRepo.GetStudentByID(req.StudentID)
	if err != nil || student == nil || student.ID == 0 {
		return nil, nil, errors.New("student not found")
	}

	if student.FeeAmount <= 0 {
		return nil, nil, errors.New("no fee is assigned to this student")
	}

	pendingAmount := student.FeeAmount - student.PaidAmount
	if pendingAmount <= 0 || !student.Pending {
		return nil, nil, errors.New("fees already fully paid for this student")
	}

	if req.AmountPaid != student.FeeAmount {
		return nil, nil, errors.New("you need to pay all the fees")
	}

	paymentTx := model.StudentPayment{
		StudentID:           student.ID,
		DepartmentPaymentID: req.FeeID,
		AmountPaid:          req.AmountPaid,
		PaymentMode:         req.PaymentMode,
	}

	if err := s.paymentRepo.CreateStudentPayment(&paymentTx); err != nil {
		return nil, nil, fmt.Errorf("failed to process payment: %w", err)
	}

	student.PaidAmount += req.AmountPaid
	if student.PaidAmount >= student.FeeAmount {
		student.Pending = false
	} else {
		student.Pending = true
	}

	if err := s.paymentRepo.UpdateStudentPaymentStatus(student); err != nil {
		return nil, nil, fmt.Errorf("payment recorded but failed to update student status: %w", err)
	}

	return &paymentTx, student, nil
}
