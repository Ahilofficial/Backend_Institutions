package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"github.com/jinzhu/copier"
	"strings"
)

type CreateStudentDTO struct {
	Name          string `json:"name"`
	Gender        string `json:"gender"`
	FacultyID     uint   `json:"faculty_id"`
	Rank          uint   `json:"rank"`
	Hosteller     bool   `json:"Hosteller"`
	Scholorship   bool   `json:"Scholorship"`
}

func (dto *CreateStudentDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Gender = strings.TrimSpace(strings.ToLower(dto.Gender))
}

func (dto *CreateStudentDTO) Validate() error {

	if dto.Name == "" {
		return errors.New("name is required")
	}
	if dto.Gender == "" {
		return errors.New("gender is required")
	}
	if dto.FacultyID == 0 {
		return errors.New("faculty id is required")
	}
	return nil
}

type UpdateStudentDTO struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

func (dto *UpdateStudentDTO) Validate() error {
	dto.Sanitize()

	if dto.Name == "" {
		return errors.New("name is required")
	}
	if dto.Gender == "" {
		return errors.New("gender is required")
	}
	return nil
}

func (dto *UpdateStudentDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Gender = strings.TrimSpace(strings.ToLower(dto.Gender))
}

type StudentFacultyDTO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	DepartmentID uint   `json:"department_id"`
}

type StudentResponseDTO struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Gender    string             `json:"gender"`
	FacultyID uint               `json:"faculty_id"`
	UserID    uint               `json:"user_id,omitempty"`
	IsActive  bool               `json:"is_active"`
	Faculty   *StudentFacultyDTO `json:"faculty,omitempty"`
	Fees      []FeesResponseDTO  `json:"fees,omitempty"`
}

func ToStudentResponseDTO(stud *model.Student) StudentResponseDTO {
	var dto StudentResponseDTO
	copier.Copy(&dto, stud)
	return dto
}

func ToStudentResponseListDTO(studs []model.Student) []StudentResponseDTO {
	list := make([]StudentResponseDTO, len(studs))

	for i := range studs {
		list[i] = ToStudentResponseDTO(&studs[i])
	}

	return list
}

type StudentMonthlyStatusDTO struct {
	StudentID     uint    `json:"student_id"`
	StudentName   string  `json:"student_name"`
	Gender        string  `json:"gender"`
	FacultyID     uint    `json:"faculty_id"`
	FacultyName   string  `json:"faculty_name,omitempty"`
	DepartmentID  uint    `json:"department_id,omitempty"`
	InstitutionID uint    `json:"institution_id,omitempty"`
	Month         string  `json:"month"`
	IsPaid        bool    `json:"is_paid"`
	AmountPaid    float64 `json:"amount_paid"`
	PaymentMode   string  `json:"payment_mode,omitempty"`
	TotalFee      float64 `json:"total_fee"`
	PendingFee    float64 `json:"pending_fee"`
}

type MonthlyStudentsOverviewDTO struct {
	Month               string                    `json:"month"`
	InstitutionID       uint                      `json:"institution_id,omitempty"`
	TotalStudents       int                       `json:"total_students"`
	PaidStudentsCount   int                       `json:"paid_students_count"`
	UnpaidStudentsCount int                       `json:"unpaid_students_count"`
	Students            []StudentMonthlyStatusDTO `json:"students"`
}
