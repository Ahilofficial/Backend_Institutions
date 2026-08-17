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

func (r *FeesRepository) CreateFees(fees *model.Fees) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	res, err := db.Exec(
		`INSERT INTO fees
		(payment_mode, total_amount, total_paid, pending_amount, student_id, created_at, updated_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		fees.PaymentMode,
		fees.TotalAmount,
		fees.TotalPaid,
		fees.PendingAmount,
		fees.StudentID,
		now,
		now,
		true,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	fees.ID = uint(id)
	fees.IsActive = true
	return nil
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

	// Fetch fees with related payments
	err := query.
		Preload("Payments").
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
		(month, amount_paid, payment_mode, fee_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		payment.Month,
		payment.AmountPaid,
		payment.PaymentMode,
		payment.FeeID,
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
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	var totalPaid float64
	query := `SELECT COALESCE(SUM(amount_paid), 0) FROM payments WHERE fee_id = ? AND deleted_at IS NULL`
	if err := db.QueryRow(query, feeID).Scan(&totalPaid); err != nil {
		return err
	}

	var totalAmount float64
	if err := db.QueryRow(`SELECT total_amount FROM fees WHERE id = ?`, feeID).Scan(&totalAmount); err != nil {
		return err
	}

	pendingAmount := totalAmount - totalPaid
	if pendingAmount < 0 {
		pendingAmount = 0
	}

	if latestPaymentMode != "" {
		_, err = db.Exec(`
			UPDATE fees
			SET total_paid = ?, pending_amount = ?, payment_mode = ?, updated_at = ?
			WHERE id = ?
		`, totalPaid, pendingAmount, latestPaymentMode, time.Now(), feeID)
	} else {
		_, err = db.Exec(`
			UPDATE fees
			SET total_paid = ?, pending_amount = ?, updated_at = ?
			WHERE id = ?
		`, totalPaid, pendingAmount, time.Now(), feeID)
	}

	return err
}

func (r *FeesRepository) FetchFeesByStudentID(studentID uint) (*model.Fees, error) {
	var fees model.Fees

	err := r.db.
		Preload("Payments").
		Where("student_id = ? AND deleted_at IS NULL", studentID).
		First(&fees).Error

	if err != nil {
		return nil, err
	}

	return &fees, nil
}

func (r *FeesRepository) GetInstitutionByFeeID(feeID uint) (uint, error) {
	var institutionID uint

	err := r.db.Raw(`
		SELECT d.institution_id
		FROM fees fe
		JOIN students s ON fe.student_id = s.id
		JOIN faculties f ON s.faculty_id = f.id
		JOIN departments d ON f.department_id = d.id
		WHERE fe.id = ?
		  AND fe.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND f.deleted_at IS NULL
		  AND d.deleted_at IS NULL
	`, feeID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}

	return institutionID, nil
}
