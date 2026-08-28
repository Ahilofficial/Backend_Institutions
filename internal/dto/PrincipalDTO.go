package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

type CreatePrincipalDTO struct {
	Name          string    `json:"name"`
	Gender        string    `json:"gender"`
	JoiningDate   time.Time `json:"joining_date"`
	InstitutionID uint      `json:"institution_id"`
}

func (dto *CreatePrincipalDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Gender = strings.TrimSpace(strings.ToLower(dto.Gender))
}

func (dto *CreatePrincipalDTO) Validate() error {
	if dto.Name == "" {
		return errors.New("name is required")
	}
	if dto.Gender == "" {
		return errors.New("gender is required")
	}
	if dto.JoiningDate.IsZero() {
		return errors.New("joining date is required")
	}
	if dto.InstitutionID == 0 {
		return errors.New("institution id is required")
	}
	return nil
}

type UpdatePrincipalDTO struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

func (dto *UpdatePrincipalDTO) Sanitize() {
	dto.Name = strings.TrimSpace(dto.Name)
	dto.Gender = strings.TrimSpace(strings.ToLower(dto.Gender))
}

func (dto *UpdatePrincipalDTO) Validate() error {
	if dto.Name == "" {
		return errors.New("name is required")
	}
	if dto.Gender == "" {
		return errors.New("gender is required")
	}
	return nil
}

type PrincipalResponseDTO struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Gender        string    `json:"gender"`
	JoiningDate   time.Time `json:"joining_date"`
	InstitutionID uint      `json:"institution_id"`
	UserID        uint      `json:"user_id,omitempty"`
	IsActive      bool      `json:"is_active"`
}

func ToPrincipalResponseDTO(pr *model.Principal) PrincipalResponseDTO {
	var dto PrincipalResponseDTO
	copier.Copy(&dto, pr)
	return dto
}

func ToPrincipalResponseListDTO(prs []model.Principal) []PrincipalResponseDTO {
	list := make([]PrincipalResponseDTO, len(prs))
	for i := range prs {
		list[i] = ToPrincipalResponseDTO(&prs[i])
	}
	return list
}
