package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"github.com/jinzhu/copier"
	"strings"
)

type CreateStudentDTO struct {
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	FacultyID uint   `json:"faculty_id"`
	Semester  uint   `json:"semester"`

	Hosteller   bool `json:"hosteller"`
	Scholorship bool `json:"scholorship"`
	MQ          bool `json:"mq"`
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
	if dto.MQ && dto.Scholorship {
		return errors.New("management quota student cannot have scholarship")
	}
	return nil
}

type UpdateStudentDTO struct {
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	Semester  uint   `json:"semester"`

	
}
type UpdateSemesterDTO struct{
	Semester  uint   `json:"semester"`
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

// type UpdateStudentSemesterDTO struct {
// 	Semester    uint  `json:"semester"`
// 	Hosteller   *bool `json:"hosteller,omitempty"`
// 	Scholarship *bool `json:"scholarship,omitempty"`
// 	MQ          *bool `json:"mq,omitempty"`
// }

// func (dto *UpdateStudentSemesterDTO) Validate() error {
// 	if dto.Semester == 0 {
// 		return errors.New("semester is required and must be greater than 0")
// 	}
// 	if dto.MQ != nil && dto.Scholarship != nil && *dto.MQ && *dto.Scholarship {
// 		return errors.New("management quota student cannot have scholarship")
// 	}
// 	return nil
// }

type StudentFacultyDTO struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	DepartmentID uint   `json:"department_id"`
}


type StudentResponseDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	FacultyID   uint   `json:"faculty_id"`
	UserID      uint   `json:"user_id,omitempty"`
	IsActive    bool   `json:"is_active"`
	Hosteller   bool   `json:"hosteller"`
	Scholorship bool   `json:"scholorship"`

	MQ              bool                        `json:"mq"`
	FeeAmount       float64                     `json:"fee_amount"`
	BaseAmount      float64                     `json:"base_amount"`
	Semester        uint                        `json:"semester"`
	Pending         bool                        `json:"pending"`
	Faculty         *StudentFacultyDTO          `json:"faculty,omitempty"`
	Fees            []FeesResponseDTO           `json:"fees,omitempty"`
	StudentPayments []StudentPaymentResponseDTO `json:"student_payments,omitempty"`
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
