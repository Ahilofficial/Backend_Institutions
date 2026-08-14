package services

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"backend_institutions/internal/repository"
	"errors"
	"fmt"
)

type FeesService struct {
	feesRepo    *repository.FeesRepository
	studentRepo *repository.StudentRepository
	userRepo    *repository.UserRepository
}

func NewFeesService(
	feesRepo *repository.FeesRepository,
	studentRepo *repository.StudentRepository,
	userRepo    *repository.UserRepository,
) *FeesService {
	return &FeesService{
		feesRepo:    feesRepo,
		studentRepo: studentRepo,
		userRepo:    userRepo,
	}
}

func (s *FeesService) checkInstitutionAccess(
	userID uint,
	institutionID uint,
) error {
	hasAccess, err := s.userRepo.RequireInstitutionAdminAccess(
		userID,
		institutionID,
	)
	if err != nil {
		return err
	}

	if !hasAccess {
		return errors.New(
			"access denied",
		)
	}

	return nil
}

func (s *FeesService) CreateFeesService(
	userID uint,
	fee *model.Fees,
) (model.Fees, error) {

	if fee.StudentID == 0 {
		return model.Fees{}, errors.New("student_id is required")
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return model.Fees{}, err
	}
	if !canManage {
		return model.Fees{}, errors.New("access denied: cant create fee for this student")
	}

	if err := s.feesRepo.CreateFees(fee); err != nil {
		return model.Fees{}, err
	}

	return *fee, nil
}

func (s *FeesService) GetFeesService() ([]model.Fees, error) {
	return s.feesRepo.FetchFees()
}

func (s *FeesService) GetFeesServicePaginated(
	search string,
	page int,
	limit int,
) ([]model.Fees, int64, error) {

	return s.feesRepo.FetchFeesPaginated(
		search,
		page,
		limit,
	)
}

func (s *FeesService) GetFeesServiceById(
	userID uint,
	id uint,
) (model.Fees, error) {
	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return model.Fees{}, err
	}

	user, err := s.userRepo.FindByID(userID)
	if err == nil && user.StudentID > 0 {
		if user.StudentID != fee.StudentID {
			return model.Fees{}, errors.New("access denied: cannot access another student's fee")
		}
		return fee, nil
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return model.Fees{}, err
	}
	if !canManage {
		return model.Fees{}, errors.New("access denied: fee does not belong to your institution")
	}

	return fee, nil
}

func (s *FeesService) UpdateFeesService(
	userID uint,
	id uint,
	dto *dto.UpdateFeesDTO,
) error {

	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return err
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return err
	}
	if !canManage {
		return errors.New("access denied: fee does not belong to your institution")
	}

	if dto.Amount != 0 {
		fee.TotalAmount = dto.Amount
	}
	if dto.PaymentMode != "" {
		fee.PaymentMode = dto.PaymentMode
	}

	return s.feesRepo.UpdateFeesById(&fee)
}

func (s *FeesService) DeleteFeesService(
	userID uint,
	id uint,
) error {

	fee, err := s.feesRepo.FetchFeesById(id)
	if err != nil {
		return err
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return err
	}
	if !canManage {
		return errors.New("access denied: fee does not belong to your institution")
	}

	return s.feesRepo.DeleteFees(id)
}

func (s *FeesService) GetInactiveFeesService() ([]model.Fees, error) {
	return s.feesRepo.FetchInactiveFees()
}

func (s *FeesService) CreatePayment(
	userID uint,
	dto *dto.CreatePaymentDTO,
) (model.Payment, error) {

	fee, err := s.feesRepo.FetchFeesById(dto.FeeID)
	if err != nil {
		return model.Payment{}, err
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return model.Payment{}, err
	}
	if !canManage {
		return model.Payment{}, errors.New("access denied: cant process payment for this student")
	}

	if fee.PendingAmount <= 0 {
		return model.Payment{}, errors.New("fee is already fully paid")
	}

	if dto.AmountPaid > fee.PendingAmount {
		return model.Payment{}, fmt.Errorf("amount paid (%.2f) exceeds pending fee amount (%.2f)", dto.AmountPaid, fee.PendingAmount)
	}

	payment := model.Payment{
		FeeID:       dto.FeeID,
		AmountPaid:  dto.AmountPaid,
		PaymentMode: dto.PaymentMode,
		Month:       dto.Month,
	}

	if err := s.feesRepo.CreatePayment(&payment); err != nil {
		return model.Payment{}, err
	}

	if err := s.feesRepo.RecalculateFeeTotals(dto.FeeID, dto.PaymentMode); err != nil {
		return model.Payment{}, err
	}

	return payment, nil
}

func (s *FeesService) GetPaymentByID(
	userID uint,
	id uint,
) (model.Payment, error) {

	payment, err := s.feesRepo.FetchPaymentByID(id)
	if err != nil {
		return model.Payment{}, err
	}

	fee, err := s.feesRepo.FetchFeesById(payment.FeeID)
	if err != nil {
		return model.Payment{}, err
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return model.Payment{}, err
	}
	if !canManage {
		return model.Payment{}, errors.New("access denied: payment does not belong to your institution")
	}

	return payment, nil
}

func (s *FeesService) GetPaymentByFeeID(
	userID uint,
	feeID uint,
) ([]model.Payment, error) {

	fee, err := s.feesRepo.FetchFeesById(feeID)
	if err != nil {
		return nil, err
	}

	canManage, err := s.userRepo.CanManageStudentFees(userID, fee.StudentID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, errors.New("access denied: fee does not belong to your institution")
	}

	return s.feesRepo.FetchPaymentByFeeID(feeID)
}

func (s *FeesService) FetchFeesByStudentID(
	userID uint,
	studentID uint,
) (*model.Fees, error) {
	user, err := s.userRepo.FindByID(userID)
	if err == nil && user.StudentID > 0 && user.StudentID != studentID {
		return nil, errors.New("access denied")
	}

	institutionID, err := s.studentRepo.GetInstitutionByStudentID(
		studentID,
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

	return s.feesRepo.FetchFeesByStudentID(studentID)
}
