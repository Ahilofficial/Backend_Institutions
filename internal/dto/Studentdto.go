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
	UserID    uint   `json:"user_id"`
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
