package dto

import (
	"backend_institutions/internal/model"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

type CreateRoleDTO struct {
	Name string `json:"name"`
}

type UpdateRoleDTO struct {
	Name string `json:"name"`
}

type AssignPermissionsDTO struct {
	PermissionIDs   []uint   `json:"permission_ids,omitempty"`
	PermissionNames []string `json:"permission_names,omitempty"`
}

type CreateUserRoleDTO struct {
	UserID uint `json:"user_id"`
	RoleID uint `json:"role_id"`
}

type UpdateUserRoleDTO struct {
	RoleID uint `json:"role_id"`
}

type CreateRolePermissionDTO struct {
	RoleID       uint `json:"role_id"`
	PermissionID uint `json:"permission_id"`
}

type UpdateRolePermissionDTO struct {
	PermissionID uint `json:"permission_id"`
}

type RolesDTOResponse struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Permissions []PermissionDTO `json:"permissions"`
}
type PermissionDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type RoleResponseDTO struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (dto *CreateRoleDTO) Sanitize() {
	dto.Name = strings.TrimSpace(strings.ToLower(dto.Name))
}

func (dto *CreateRoleDTO) Validate() error {
	if dto.Name == "" {
		return errors.New("role name is required")
	}
	return nil
}

func (dto *UpdateRoleDTO) Sanitize() {
	dto.Name = strings.TrimSpace(strings.ToLower(dto.Name))
}

func (dto *UpdateRoleDTO) Validate() error {
	if dto.Name == "" {
		return errors.New("role name is required")
	}
	return nil
}

func (dto *AssignPermissionsDTO) Sanitize() {
	for i, name := range dto.PermissionNames {
		dto.PermissionNames[i] = strings.TrimSpace(name)
	}
}

func (dto *AssignPermissionsDTO) Validate() error {
	if len(dto.PermissionIDs) == 0 && len(dto.PermissionNames) == 0 {
		return errors.New("either permission_ids or permission_names must be provided")
	}
	return nil
}

func (dto *CreateUserRoleDTO) Sanitize() {
}

func (dto *CreateUserRoleDTO) Validate() error {
	if dto.UserID == 0 {
		return errors.New("user_id is required")
	}
	if dto.RoleID == 0 {
		return errors.New("role_id is required")
	}
	return nil
}

func (dto *UpdateUserRoleDTO) Sanitize() {
}

func (dto *UpdateUserRoleDTO) Validate() error {
	if dto.RoleID == 0 {
		return errors.New("role_id is required")
	}
	return nil
}

func (dto *CreateRolePermissionDTO) Sanitize() {
}

func (dto *CreateRolePermissionDTO) Validate() error {
	if dto.RoleID == 0 {
		return errors.New("role_id is required")
	}
	if dto.PermissionID == 0 {
		return errors.New("permission_id is required")
	}
	return nil
}

func (dto *UpdateRolePermissionDTO) Sanitize() {
}

func (dto *UpdateRolePermissionDTO) Validate() error {
	if dto.PermissionID == 0 {
		return errors.New("permission_id is required")
	}
	return nil
}

func ToRoleResponseDTO(role *model.Role) RoleResponseDTO {
	var dto RoleResponseDTO
	copier.Copy(&dto, role)
	return dto
}

func ToRoleResponseListDTO(roles []model.Role) []RoleResponseDTO {
	list := make([]RoleResponseDTO, len(roles))
	for i, r := range roles {
		list[i] = ToRoleResponseDTO(&r)
	}
	return list
}
