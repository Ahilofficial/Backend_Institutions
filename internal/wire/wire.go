//go:build wireinject
// +build wireinject

package wire

import (
	"backend_institutions/internal/controller"
	"backend_institutions/internal/database"
	"backend_institutions/internal/repository"
	"backend_institutions/internal/routes"
	"backend_institutions/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

func InitializeApp() (*fiber.App, error) {
	wire.Build(
		// Database
		database.NewDB,

		// Repositories
		repository.NewUserRepository,
		repository.NewInstitutionRepository,
		repository.NewDepartmentRepository,
		repository.NewFacultyRepository,
		repository.NewPrincipalRepository,
		repository.NewStudentRepository,
		repository.NewFeesRepository,
		repository.NewRoleRepository,
		repository.NewPermissionRepository,
		repository.NewSessionRepository,
		repository.NewMenuRepository,

		// Services
		services.NewSessionService,
		services.NewUserService,
		services.NewInstituteService,
		services.NewDepartmentService,
		services.NewFacultyService,
		services.NewPrincipalService,
		services.NewStudentService,
		services.NewFeesService,
		services.NewRoleService,
		services.NewPermissionService,
		services.NewMenuService,

		// Controllers
		controller.NewUserController,
		controller.NewInstituteController,
		controller.NewDepartmentController,

		// FacultyController requires:
		// FacultyService + UserService
		controller.NewFacultyController,

		controller.NewPrincipalControllers,
		controller.NewStudentController,
		controller.NewFeesController,
		controller.NewRoleController,
		controller.NewPermissionController,
		controller.NewMenuController,

		// Routes
		routes.NewApp,
	)

	return nil, nil
}