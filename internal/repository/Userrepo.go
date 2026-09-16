package repository

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/model"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpdateFacultyID(userID uint, facultyID uint) error {
	result := r.db.
		Exec("UPDATE users SET faculty_id = ? WHERE id = ?", facultyID, userID).Error

	return result
}

func (r *UserRepository) GetUserByID(userID uint) (*model.User, error) {
	var user model.User

	result := r.db.Raw(`
		SELECT *
		FROM users
		WHERE id = ?
		LIMIT 1
	`, userID).Scan(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &user, nil
}

func (r *UserRepository) AssignRoleByName(userID uint, roleName string) error {
	var roleID uint
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return errors.New("role name is required")
	}

	err := r.db.Raw("SELECT id FROM roles WHERE LOWER(name) = LOWER(?) AND deleted_at IS NULL LIMIT 1", roleName).Scan(&roleID).Error
	if err != nil || roleID == 0 {

		res := r.db.Exec("INSERT INTO roles (name) VALUES (?)", roleName)
		if res.Error == nil {
			_ = r.db.Raw("SELECT id FROM roles WHERE LOWER(name) = LOWER(?) AND deleted_at IS NULL LIMIT 1", roleName).Scan(&roleID)
		}
	}

	if roleID == 0 {
		return errors.New("failed to resolve role id for role: " + roleName)
	}

	return r.db.Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID).Error
}

func (r *UserRepository) GetUserInstitutionID(userID uint) (uint, error) {
	if userID == 0 {
		return 0, nil
	}

	var instID uint
	_ = r.db.Table("institution_admins").Where("user_id = ?", userID).Select("institution_id").Scan(&instID).Error
	if instID > 0 {
		return instID, nil
	}

	var faculty model.Faculty
	if err := r.db.Preload("Department").Where("user_id = ? AND deleted_at IS NULL", userID).First(&faculty).Error; err == nil && faculty.Department.InstitutionID > 0 {
		return faculty.Department.InstitutionID, nil
	}

	var student model.Student
	if err := r.db.Preload("Faculty.Department").Where("user_id = ? AND deleted_at IS NULL", userID).First(&student).Error; err == nil && student.Faculty.Department.InstitutionID > 0 {
		return student.Faculty.Department.InstitutionID, nil
	}

	return 0, nil
}

func (r *UserRepository) CheckUserRole(userID uint, targetRole string) (bool, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return false, err
	}
	for _, role := range user.Roles {
		if strings.EqualFold(role.Name, targetRole) {
			return true, nil
		}
	}
	return false, nil
}

func (r *UserRepository) GetUserRoles(userID uint) ([]string, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, strings.ToLower(strings.TrimSpace(role.Name)))
	}
	return roles, nil
}

func (r *UserRepository) UpdateStudentID(userID uint, studentID uint) error {
	result := r.db.Exec(`
		UPDATE users
		SET student_id = ?
		WHERE id = ?
	`, studentID, userID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) FindByVerificationToken(token string) (model.User, error) {
	var user model.User

	query := `
		SELECT *
		FROM users
		WHERE verification_token = ?
		LIMIT 1
	`

	err := r.db.Raw(query, token).Scan(&user).Error
	if err != nil {
		return model.User{}, err
	}
	if user.ID == 0 {
		return model.User{}, gorm.ErrRecordNotFound
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(user *model.User) error {
	var tokenVal any = user.VerificationToken
	if user.VerificationToken == "" {
		tokenVal = nil
	}

	var expiresAt any = user.TokenExpiresAt
	if user.TokenExpiresAt.IsZero() {
		expiresAt = nil
	}

	query := `
		UPDATE users
		SET
			is_active = ?,
			is_verified = ?,
			verification_token = ?,
			token_expires_at = ?,
			updated_at = NOW()
		WHERE id = ?
	`

	return r.db.Exec(
		query,
		user.IsActive,
		user.IsVerified,
		tokenVal,
		expiresAt,
		user.ID,
	).Error
}

func (r *UserRepository) CreateUser(user *model.User) error {

	var phonecount uint

	err := r.db.Raw(`select count(*) from users where phone=?`, user.Phone).Scan(&phonecount).Error
	if err != nil {
		return err
	}
	if phonecount > 0 {
		return errors.New("Phone number  already there")
	}
	var emailcount uint
	email_err := r.db.Raw(`select count(*) from users where email=?`, user.Email).Scan(&emailcount).Error
	if email_err != nil {
		return email_err
	}
	if emailcount > 0 {
		return errors.New("Email  already there")
	}
	now := time.Now()
	var tokenVal any = user.VerificationToken
	var expiresAt any = user.TokenExpiresAt

	result, _ := r.db.DB()

	query := `
		INSERT INTO users (name, email, phone, password, is_active, is_verified, verification_token, token_expires_at, created_at, updated_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insert_records, err := result.Exec(query, user.Name, user.Email, user.Phone, user.Password, user.IsActive, user.IsVerified, tokenVal, expiresAt, now, now)
	if err != nil {
		return err
	}

	lastInsertedID, err := insert_records.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = uint(lastInsertedID)
	return nil

}

func (r *UserRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Raw("SELECT * FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1", email).Scan(&user).Error
	if err != nil {
		return user, err
	}
	if user.ID == 0 {
		return user, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *UserRepository) FindByPhone(phone string) (model.User, error) {
	var user model.User
	err := r.db.Raw("SELECT * FROM users WHERE phone = ? AND deleted_at IS NULL LIMIT 1", phone).Scan(&user).Error
	if err != nil {
		return user, err
	}
	if user.ID == 0 {
		return user, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (r *UserRepository) AssignRoleToUser(userID uint, roleID uint) error {
	if err := r.db.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
		return err
	}

	result := r.db.Exec(
		"INSERT INTO user_roles (user_id, role_id) SELECT ?, id FROM roles WHERE id = ?",
		userID,
		roleID,
	)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("role does not exist")
	}

	return nil
}

func (r *UserRepository) FindRoleByName(name string) (model.Role, error) {
	var role model.Role
	err := r.db.Raw("SELECT id, name FROM roles WHERE name = ? LIMIT 1", name).Scan(&role).Error
	if err != nil {
		return role, err
	}
	if role.ID == 0 {
		return role, gorm.ErrRecordNotFound
	}
	return role, nil
}

func (r *UserRepository) DeleteUser(id uint) error {
	res := r.db.Exec(
		"UPDATE users SET is_active = ?, deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		false,
		time.Now(),
		time.Now(),
		id,
	)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("user not found or already deleted")
	}

	return nil
}

func (s *UserRepository) ForgotPasswordRepo(dto dto.ForgotPasswordDTO) (model.User, error) {
	var user model.User
	query := `select * from users where email=? limit 1`
	result := s.db.Raw(query, dto.Email).Scan(&user)
	if result.Error != nil {
		return model.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return model.User{}, errors.New("email not found")
	}

	return user, nil
}

func (r *UserRepository) AssignRole(
	userID uint,
	roleID uint,
) error {

	err := r.db.Exec(
		`INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`,
		userID,
		roleID,
	).Error

	return err
}

func (r *UserRepository) UpdateResetToken(user model.User) error {
	query := `
		UPDATE users
		SET
			reset_password_token = ?,
			reset_token_expires_at = ?
		WHERE id = ?
	`

	return r.db.Exec(
		query,
		user.ResetPasswordToken,
		user.ResetTokenExpiresAt,
		user.ID,
	).Error
}

func (r *UserRepository) FetchUsertoken(token string) (model.User, error) {
	var user model.User

	query := `
        SELECT *
        FROM users
        WHERE reset_password_token = ?
        LIMIT 1
    `

	result := r.db.Raw(query, token).Scan(&user)

	if result.Error != nil {
		return model.User{}, result.Error
	}

	if result.RowsAffected == 0 {
		return model.User{}, errors.New("invalid reset token")
	}

	return user, nil
}

func (r *UserRepository) UpdatePassword(id uint, password string) error {
	query := `
		UPDATE users
		SET
			password = ?,
			reset_password_token = NULL,
			reset_token_expires_at = NULL
		WHERE id = ?
	`

	return r.db.Exec(query, password, id).Error
}

func (r *UserRepository) Logout(dto *dto.LogoutDTO) error {
	if dto.Token == "" {
		return errors.New("refresh token is required")
	}

	return r.db.Exec(`
		UPDATE sessions
		SET access_token = NULL,
		    refresh_token = NULL,
		    is_active = FALSE
		WHERE refresh_token = ?
	`, dto.Token).Error
}

func (r *UserRepository) FindByID(userID uint) (model.User, error) {
	var user model.User

	err := r.db.Raw(`
		SELECT *
		FROM users
		WHERE id = ?
		AND deleted_at IS NULL
		LIMIT 1
	`, userID).Scan(&user).Error

	if err != nil {
		return model.User{}, err
	}

	if user.ID == 0 {
		return model.User{}, gorm.ErrRecordNotFound
	}

	return user, nil
}

func (r *UserRepository) FetchUserRoles(userID uint) (model.Role, error) {
	var roles model.Role

	query := `
		SELECT r.id, r.name
		FROM roles r
		INNER JOIN user_roles ur
			ON ur.role_id = r.id
		WHERE ur.user_id = ?
	`

	result := r.db.Raw(query, userID).Scan(&roles)
	if result.Error != nil {
		return model.Role{}, result.Error
	}

	if result.RowsAffected == 0 {
		return model.Role{}, errors.New("no roles assigned to user")
	}

	return roles, nil
}

func (r *UserRepository) HasPermission(
	userID uint,
	permission string,
) (bool, error) {
	var count int64

	query := `
		SELECT COUNT(*)
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.id = ?
		  AND p.name = ?
	`

	err := r.db.Raw(
		query,
		userID,
		permission,
	).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) GetUserFacultyID(userID uint) (uint, error) {
	if userID == 0 {
		return 0, nil
	}

	var facultyID *uint

	err := r.db.Raw(`
		SELECT faculty_id
		FROM users
		WHERE id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, userID).Scan(&facultyID).Error

	if err != nil {
		return 0, err
	}

	if facultyID == nil {
		return 0, nil
	}

	return *facultyID, nil
}

func (r *UserRepository) GetUserStudentID(userID uint) (uint, error) {
	if userID == 0 {
		return 0, nil
	}

	var studentID *uint

	err := r.db.Raw(`
		SELECT student_id
		FROM users
		WHERE id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, userID).Scan(&studentID).Error

	if err != nil {
		return 0, err
	}

	if studentID == nil {
		return 0, nil
	}

	return *studentID, nil
}

func (r *UserRepository) GetInstitutionAdminID(userID uint) (uint, error) {
	var institutionID uint

	err := r.db.Raw(`
		SELECT institution_id
		FROM institution_admins
		WHERE user_id = ?
		LIMIT 1
	`, userID).Scan(&institutionID).Error

	if err != nil {
		return 0, err
	}

	return institutionID, nil
}

func (r *UserRepository) GetUserRoleID(userID uint) (uint, error) {
	var roleID uint

	err := r.db.Raw(`
		SELECT role_id
		FROM user_roles
		WHERE user_id = ?
		LIMIT 1
	`, userID).Scan(&roleID).Error

	if err != nil {
		return 0, err
	}

	if roleID == 0 {
		return 0, errors.New("role not assigned to user")
	}

	return roleID, nil
}

func (r *UserRepository) CheckUserExistingProfileFaculty(userID uint) (bool, string) {
	var user model.User

	err := r.db.Raw(`
	SELECT id, student_id, faculty_id
	FROM users
	WHERE id = ?
	AND deleted_at IS NULL
	LIMIT 1
`, userID).Scan(&user).Error

	if err != nil {
		return false, "User not found"
	}

	if user.StudentID > 0 {
		return true, "Student profile already registered"
	}

	if user.FacultyID > 0 {
		return true, "Faculty profile already registered"
	}

	return false, ""
}

func (r *UserRepository) UpdateUserStudentID(userID uint, studentID uint) error {
	if userID == 0 {
		return nil
	}

	res := r.db.Exec("UPDATE users SET student_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", studentID, userID)
	return res.Error
}

func (r *UserRepository) UpdateUserFacultyID(userID uint, facultyID uint) error {
	if userID == 0 {
		return nil
	}

	res := r.db.Exec("UPDATE users SET faculty_id = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", facultyID, userID)
	return res.Error
}

func (r *UserRepository) IsSuperAdminRepo(userID uint) bool {
	if userID == 0 {
		return false
	}

	var roleName string

	err := r.db.Raw(`
		SELECT r.name
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ?
		  AND r.deleted_at IS NULL
		LIMIT 1
	`, userID).Scan(&roleName).Error

	if err != nil {
		return false
	}

	return roleName == "super_admin"
}

func(r *UserRepository)IsSuperAdmin(userID uint)(bool,error){
	var user_role uint
	err:=r.db.Raw(`select role_id from user_role where user_id=?`,userID).Scan(&user_role).Error
	if err!=nil{
		return false,err
	}
	var role_name string
	role_err:=r.db.Raw(`select names from roles where id=?`,user_role).Scan(&role_name).Error
	if role_err!=nil{
		return false, role_err
	}
	if role_name!="super_admin"{
		return false,nil
	}
	return true,nil
}