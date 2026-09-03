package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"strings"

	"github.com/jinzhu/copier"
)

type CreateInstitutionDTO struct {
	Name            string `json:"name"`
	InstitutionCode string `json:"institution_code"`
	State           string `json:"state"`
}

func (dto *CreateInstitutionDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.InstitutionCode = strings.TrimSpace(dto.InstitutionCode)
	dto.State = strings.TrimSpace(dto.State)
}

func (dto *CreateInstitutionDTO) Validate() error {

	if dto.Name == "" {
		return errors.New("name is required")
	}
	if dto.InstitutionCode == "" {
		return errors.New("institution code is required")
	}
	if dto.State == "" {
		return errors.New("state is required")
	}
	return nil
}

type UpdateInstitutionDTO struct {
	Name            string `json:"name"`
	InstitutionCode string `json:"institution_code"`
	State           string `json:"state"`
}

func (dto *UpdateInstitutionDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.InstitutionCode = strings.TrimSpace(dto.InstitutionCode)
	dto.State = strings.TrimSpace(dto.State)
}

func (dto *UpdateInstitutionDTO) Validate() error {

	if dto.Name == "" {
		return errors.New("name is required")
	}
	if dto.InstitutionCode == "" {
		return errors.New("institution code is required")
	}
	if dto.State == "" {
		return errors.New("state is required")
	}
	return nil
}

type InstitutionResponseDTO struct {
	ID              uint                    `json:"id"`
	Name            string                  `json:"name"`
	InstitutionCode string                  `json:"institution_code"`
	State           string                  `json:"state"`
	IsActive        bool                    `json:"isactive"`
	Departments     []DepartmentResponseDTO `json:"departments"`
}

func ToInstitutionResponseDTO(inst *model.Institutions) InstitutionResponseDTO {
	var dto InstitutionResponseDTO
	copier.Copy(&dto, inst)
	dto.Departments = make([]DepartmentResponseDTO, len(inst.Departments))
	for i := range inst.Departments {
		dto.Departments[i] = ToDepartmentResponseDTO(&inst.Departments[i])
	}
	return dto
}

func ToInstitutionResponseListDTO(insts []model.Institutions) []InstitutionResponseDTO {
	list := make([]InstitutionResponseDTO, len(insts))

	for i := range insts {
		list[i] = ToInstitutionResponseDTO(&insts[i])
	}

	return list
}
