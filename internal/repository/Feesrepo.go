package repository

import (
	"backend_institutions/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// FeesRepository handles persistence and queries for department fee templates, student fees, and payments
type FeesRepository struct {
	db *gorm.DB
}

// NewFeesRepository initializes a new instance of FeesRepository
func NewFeesRepository(db *gorm.DB) *FeesRepository {
	return &FeesRepository{db: db}
}

// CreateDepartmentFee inserts a department semester fee configuration template
func (r *FeesRepository) CreateDepartmentFee(fee *model.Fees) error {
	// 1. Insert fee template with NULL student_id
	err := r.db.Exec(`
INSERT INTO fees (department_id, semester, college_amount, hostel_amount, total_amount, pending_amount, payment_mode, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		fee.DepartmentID,
		fee.Semester,
		fee.CollegeAmount,
		fee.HostelAmount,
		fee.TotalAmount,
		fee.PendingAmount,
		fee.PaymentMode,
		fee.IsActive,
	).Error

	return err
}

// GetDepartmentFeeBySemester retrieves department fee template for a semester
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

// GetDepartmentFees retrieves all fee templates for a department ordered by semester
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

// UpdateDepartmentFee updates amounts for a department fee template
func (r *FeesRepository) UpdateDepartmentFee(fee *model.Fees) error {
	if fee.TotalAmount == 0 && (fee.CollegeAmount > 0 || fee.HostelAmount > 0) {
		fee.TotalAmount = fee.CollegeAmount + fee.HostelAmount
	}
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	// 1. Execute update query
	res, err := db.Exec(`
		UPDATE fees
		SET college_amount = ?, hostel_amount = ?, total_amount = ?, updated_at = ?
		WHERE id = ? AND student_id IS NULL AND deleted_at IS NULL
	`, fee.CollegeAmount, fee.HostelAmount, fee.TotalAmount, time.Now(), fee.ID)

	if err != nil {
		return err
	}

	// 2. Verify affected rows
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("department fee not found or already deleted")
	}

	return nil
}

// DeleteDepartmentFee soft deletes a department fee template
func (r *FeesRepository) DeleteDepartmentFee(id uint) error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	// 1. Execute soft delete query
	res, err := db.Exec(`
		UPDATE fees
		SET is_active = false, deleted_at = ?
		WHERE id = ? AND student_id IS NULL AND deleted_at IS NULL
	`, time.Now(), id)

	if err != nil {
		return err
	}

	// 2. Verify rows affected
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("department fee not found or already deleted")
	}

	return nil
}

// CreateFees inserts a student fee record
func (r *FeesRepository) CreateFees(fees *model.Fees) error {
	if fees.Semester == 0 {
		fees.Semester = 1
	}
	fees.IsActive = true
	return r.db.Create(fees).Error
}

// FetchFeeByStudentAndSemester fetches a fee record for a specific student and semester
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

// FetchFeeByDepartmentID retrieves template fee record for department
func (r *FeesRepository) FetchFeeByDepartmentID(departmentID uint) (*model.Fees, error) {
	var fee model.Fees
	err := r.db.Where("department_id = ? AND student_id IS NULL AND deleted_at IS NULL", departmentID).First(&fee).Error
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

// FetchFees retrieves all fee records
func (r *FeesRepository) FetchFees() ([]model.Fees, error) {
	var fees []model.Fees
	err := r.db.Raw("SELECT * FROM fees WHERE deleted_at IS NULL").Scan(&fees).Error
	return fees, err
}

// FetchFeesPaginated retrieves paginated fee records with optional search
func (r *FeesRepository) FetchFeesPaginated(search string, page, limit int) ([]model.Fees, int64, error) {
	var fees []model.Fees
	var total int64

	searchPattern := "%" + search + "%"

	// 1. Build query with search filter
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

	// 2. Count total matches
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	// 3. Query paginated results with preloads
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

// FetchFeesById retrieves a fee record by primary key ID with associated relations
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

// DeleteFees soft deletes a fee record
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

// FetchUserStudentID retrieves student_id for a given user_id
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

// FetchInactiveFees retrieves all deactivated fee records
func (r *FeesRepository) FetchInactiveFees() ([]model.Fees, error) {
	var fees []model.Fees
	err := r.db.Raw("SELECT * FROM fees WHERE is_active = ? AND deleted_at IS NULL", false).Scan(&fees).Error
	return fees, err
}

// CreatePayment inserts a payment transaction record into the payments table
func (r *FeesRepository) CreatePayment(payment *model.Payment) error {
	// 1. Get raw database handle
	db, err := r.db.DB()
	if err != nil {
		return err
	}

	now := time.Now()

	// 2. Insert payment transaction
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

	// 3. Set generated payment ID
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	payment.ID = uint(id)
	payment.CreatedAt = now
	payment.UpdatedAt = now

	return nil
}

// FetchPaymentByFeeID retrieves all payments associated with a fee ID
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

// FetchPaymentByID retrieves a single payment transaction by ID
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

// UpdateFeesById updates payment mode and balances on a fee record
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

// RecalculateFeeTotals calculates total paid amount from all payments and updates pending balance
func (r *FeesRepository) RecalculateFeeTotals(feeID uint, latestPaymentMode string) error {
	// 1. Fetch all payment records for this fee
	var payments []model.Payment
	if err := r.db.Where("fee_id = ? AND deleted_at IS NULL", feeID).Find(&payments).Error; err != nil {
		return err
	}

	// 2. Sum paid amount in Go
	var totalPaid float64
	for _, p := range payments {
		totalPaid += p.AmountPaid
	}

	// 3. Fetch fee record
	var fee model.Fees
	if err := r.db.Where("id = ? AND deleted_at IS NULL", feeID).First(&fee).Error; err != nil {
		return err
	}

	// 4. Calculate pending balance
	pendingAmount := fee.TotalAmount - totalPaid
	if pendingAmount < 0 {
		pendingAmount = 0
	}

	// 5. Update fee record
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

// FetchFeesByStudentID retrieves all fees for a student ordered by semester
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

// GetInstitutionByFeeID resolves the institution ID owning a fee via department relationship
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
