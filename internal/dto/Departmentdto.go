package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"github.com/jinzhu/copier"
	"strings"
)

type CreateDepartmentDTO struct {
	DepartmentName string `json:"department_name"`
	InstitutionID  uint   `json:"institution_id"`
	CourseDuration uint   `json:"course_duration"`
}

func (dto *CreateDepartmentDTO) Sanitize() {
	dto.DepartmentName = strings.TrimSpace(dto.DepartmentName)
}

func (dto *CreateDepartmentDTO) Validate() error {

	if dto.DepartmentName == "" {
		return errors.New("department name is required")
	}
	if dto.InstitutionID == 0 {
		return errors.New("institution id is required")
	}
	if dto.CourseDuration == 0 {
		return errors.New("course duration is required")
	}
	return nil
}

type UpdateDepartmentDTO struct {
	DepartmentName string  `json:"department_name"`
	FeeAmount      float64 `json:"fee_amount"`
	CourseDuration uint    `json:"course_duration"`
}

func (dto *UpdateDepartmentDTO) Sanitize() {
	dto.DepartmentName = strings.TrimSpace(dto.DepartmentName)
}

func (dto *UpdateDepartmentDTO) Validate() error {

	if dto.DepartmentName == "" {
		return errors.New("department name is required")
	}
	return nil
}

type DepartmentResponseDTO struct {
	ID             uint                 `json:"id"`
	DepartmentName string               `json:"department_name"`
	CollegeAmount  float64              `json:"college_amount"`
	HostelAmount   float64              `json:"hostel_amount"`
	FeeAmount      float64              `json:"fee_amount"`
	PaymentID      uint                 `json:"payment_id"`
	InstitutionID  uint                 `json:"institution_id"`
	IsActive       bool                 `json:"isactive"`
	Faculties      []FacultyResponseDTO `json:"faculties"`
}

func ToDepartmentResponseDTO(dept *model.Department) DepartmentResponseDTO {
	var dto DepartmentResponseDTO
	copier.Copy(&dto, dept)

	return dto
}

func ToDepartmentResponseListDTO(depts []model.Department) []DepartmentResponseDTO {
	list := make([]DepartmentResponseDTO, len(depts))

	for i := range depts {
		list[i] = ToDepartmentResponseDTO(&depts[i])
	}

	return list
}
