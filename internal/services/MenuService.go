package services

import (
	"backend_institutions/internal/dto"

	"backend_institutions/internal/repository"
)

type MenuService struct {
	menuRepo *repository.MenuRepository
}

func NewMenuService(repo *repository.MenuRepository) *MenuService {
	return &MenuService{
		menuRepo: repo,
	}
}

func (s *MenuService) GetMenus(userID uint) ([]dto.MenuResponse, error) {

	menus, err := s.menuRepo.GetMenusByUser(userID)
	if err != nil {
		return nil, err
	}

	var response []dto.MenuResponse

	for _, menu := range menus {
		response = append(response, dto.MenuResponse{
			ID:    menu.ID,
			Name:  menu.Name,
			Route: menu.Route,
			Icon:  menu.Icon,
		})
	}

	return response, nil
}
