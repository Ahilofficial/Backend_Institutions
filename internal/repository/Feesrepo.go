package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type FeesRepository struct {
	db *gorm.DB
}

func NewFeesRepository(db *gorm.DB) *FeesRepository {
	return &FeesRepository{db: db}
}

func (r *FeesRepository) CreateDepartmentFee(fee *model.Fees) error {
	fee.StudentID = nil
	if fee.Semester == 0 {
		fee.Semester = 1
	}
	if fee.TotalAmount == 0 && (fee.CollegeAmount > 0 || fee.HostelAmount > 0) {
		fee.TotalAmount = fee.CollegeAmount + fee.HostelAmount
	}
	fee.IsActive = true
	return r.db.Create(fee).Error
}

func (r *FeesRepository) GetDepartmentFeeBySemester(departmentID uint, semester uint) (*model.Fees, error) {
	var fee model.Fees
	err := r.db.Raw(`
		SELECT *
		FROM fees
		WHERE department_id = ?
		  AND semester = ?
		  AND student_id IS NULL
		  AND is_active = true
		  AND deleted_at IS NULL
		LIMIT 1
	`, departmentID, semester).Scan(&fee).Error

	if err != nil {
		return nil, err
	}
	if fee.ID == 0 {
		return nil, errors.New("fee configuration not found for this department and semester")
	}

	return &fee, nil
}

func (r *FeesRepository) GetDepartmentFees(departmentID uint) ([]model.Fees, error) {
	var fees []model.Fees
	err := r.db.
		Preload("Department").
		Where("department_id = ? AND student_id IS NULL AND is_active = true AND deleted_at IS NULL", departmentID).
		Order("semester ASC").
		Find(&fees).Error

	if err != nil {
		return nil, err
	}
	return fees, nil
}

func (r *FeesRepository) UpdateDepartmentFee(fee *model.Fees) error {
	if fee.TotalAmount == 0 && (fee.CollegeAmount > 0 || fee.HostelAmount > 0) {
		fee.TotalAmount = fee.CollegeAmount + fee.HostelAmount
	}
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	res, err := db.Exec(`
		UPDATE fees
		SET college_amount = ?, hostel_amount = ?, total_amount = ?, updated_at = ?
		WHERE id = ? AND student_id IS NULL AND deleted_at IS NULL
	`, fee.CollegeAmount, fee.HostelAmount, fee.TotalAmount, time.Now(), fee.ID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("department fee not found or already deleted")
	}

	return nil
}

func (r *FeesRepository) DeleteDepartmentFee(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	res, err := db.Exec(`
		UPDATE fees
		SET is_active = false, deleted_at = ?
		WHERE id = ? AND student_id IS NULL AND deleted_at IS NULL
	`, time.Now(), id)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("department fee not found or already deleted")
	}

	return nil
}

func (r *FeesRepository) CreateFees(fees *model.Fees) error {
	if fees.Semester == 0 {
		fees.Semester = 1
	}
	fees.IsActive = true
	return r.db.Create(fees).Error
}

func (r *FeesRepository) FetchFeeByStudentAndSemester(studentID uint, semester uint) (*model.Fees, error) {
	var fee model.Fees
	err := r.db.
		Preload("Payments").
		Where("student_id = ? AND semester = ? AND deleted_at IS NULL", studentID, semester).
		First(&fee).Error
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

func (r *FeesRepository) FetchFeeByDepartmentID(departmentID uint) (*model.Fees, error) {
	var fee model.Fees
	err := r.db.Where("department_id = ? AND student_id IS NULL AND deleted_at IS NULL", departmentID).First(&fee).Error
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

func (r *FeesRepository) FetchFees() ([]model.Fees, error) {
	var fees []model.Fees
	err := r.db.Raw("SELECT * FROM fees WHERE deleted_at IS NULL").Scan(&fees).Error
	return fees, err
}

func (r *FeesRepository) FetchFeesPaginated(search string, page, limit int) ([]model.Fees, int64, error) {
	var fees []model.Fees
	var total int64

	searchPattern := "%" + search + "%"

	query := r.db.Model(&model.Fees{}).
		Where(`
			deleted_at IS NULL
			AND (
				CAST(total_amount AS CHAR) LIKE ?
				OR CAST(total_paid AS CHAR) LIKE ?
				OR CAST(pending_amount AS CHAR) LIKE ?
				OR CAST(student_id AS CHAR) LIKE ?
			)
		`, searchPattern, searchPattern, searchPattern, searchPattern)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	err := query.
		Preload("Payments").
		Preload("Student").
		Preload("Department").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&fees).Error
	if err != nil {
		return nil, 0, err
	}

	return fees, total, nil
}

func (r *FeesRepository) FetchFeesById(id uint) (model.Fees, error) {
	var fee model.Fees

	err := r.db.
		Preload("Student").
		Preload("Payments").
		Preload("Department").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&fee).Error

	if err != nil {
		return model.Fees{}, err
	}

	return fee, nil
}

func (r *FeesRepository) DeleteFees(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	res, err := db.Exec(
		"UPDATE fees SET is_active = ?, deleted_at = ? WHERE id = ? AND is_active = ? AND deleted_at IS NULL",
		false, time.Now(), id, true,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("record not found or already deleted")
	}
	return nil
}

func (r *FeesRepository) FetchUserStudentID(userID uint) (uint, error) {
	var studentID uint
	err := r.db.Raw(`
		SELECT student_id FROM users WHERE id = ? AND deleted_at IS NULL
	`, userID).Scan(&studentID).Error
	if err != nil {
		return 0, err
	}
	return studentID, nil
}

func (r *FeesRepository) FetchInactiveFees() ([]model.Fees, error) {
	var fees []model.Fees
	err := r.db.Raw("SELECT * FROM fees WHERE is_active = ? AND deleted_at IS NULL", false).Scan(&fees).Error
	return fees, err
}

func (r *FeesRepository) CreatePayment(payment *model.Payment) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	res, err := db.Exec(
		`INSERT INTO payments
		(amount_paid, payment_mode, fee_id, student_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		payment.AmountPaid,
		payment.PaymentMode,
		payment.FeeID,
		payment.StudentID,
		now,
		now,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	payment.ID = uint(id)
	payment.CreatedAt = now
	payment.UpdatedAt = now

	return nil
}

func (r *FeesRepository) FetchPaymentByFeeID(feeID uint) ([]model.Payment, error) {
	var payments []model.Payment

	err := r.db.Raw(`
		SELECT *
		FROM payments
		WHERE fee_id = ?
		AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, feeID).Scan(&payments).Error

	if err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *FeesRepository) FetchPaymentByID(id uint) (model.Payment, error) {
	var payment model.Payment

	err := r.db.
		Preload("Fee").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&payment).Error

	if err != nil {
		return model.Payment{}, err
	}

	return payment, nil
}

func (r *FeesRepository) UpdateFeesById(fees *model.Fees) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE fees
		SET
			payment_mode = ?,
			total_paid = ?,
			pending_amount = ?,
			updated_at = ?
		WHERE id = ?
	`,
		fees.PaymentMode,
		fees.TotalPaid,
		fees.PendingAmount,
		time.Now(),
		fees.ID,
	)

	return err
}

func (r *FeesRepository) RecalculateFeeTotals(feeID uint, latestPaymentMode string) error {
	var payments []model.Payment
	if err := r.db.Where("fee_id = ? AND deleted_at IS NULL", feeID).Find(&payments).Error; err != nil {
		return err
	}

	var totalPaid float64
	for _, p := range payments {
		totalPaid += p.AmountPaid
	}

	var fee model.Fees
	if err := r.db.Where("id = ? AND deleted_at IS NULL", feeID).First(&fee).Error; err != nil {
		return err
	}

	pendingAmount := fee.TotalAmount - totalPaid
	if pendingAmount < 0 {
		pendingAmount = 0
	}

	updates := map[string]interface{}{
		"total_paid":     totalPaid,
		"pending_amount": pendingAmount,
		"updated_at":     time.Now(),
	}
	if latestPaymentMode != "" {
		updates["payment_mode"] = latestPaymentMode
	}

	return r.db.Model(&model.Fees{}).Where("id = ? AND deleted_at IS NULL", feeID).Updates(updates).Error
}

func (r *FeesRepository) FetchFeesByStudentID(studentID uint) ([]model.Fees, error) {
	var fees []model.Fees

	err := r.db.
		Preload("Payments").
		Preload("Department").
		Where("student_id = ? AND deleted_at IS NULL", studentID).
		Order("semester ASC").
		Find(&fees).Error

	if err != nil {
		return nil, err
	}

	return fees, nil
}

func (r *FeesRepository) GetInstitutionByFeeID(feeID uint) (uint, error) {
	var fee model.Fees
	err := r.db.
		Preload("Department").
		Where("id = ? AND deleted_at IS NULL", feeID).
		First(&fee).Error
	if err != nil {
		return 0, err
	}
	return fee.Department.InstitutionID, nil
}
