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

		database.NewDB,

		repository.NewUserRepository,
		repository.NewInstitutionRepository,
		repository.NewDepartmentRepository,
		repository.NewFacultyRepository,
		repository.NewStudentRepository,
		repository.NewFeesRepository,
		repository.NewRoleRepository,
		repository.NewSessionRepository,
		repository.NewMenuRepository,

		services.NewSessionService,
		services.NewUserService,
		services.NewInstituteService,
		services.NewDepartmentService,
		services.NewFacultyService,
		services.NewStudentService,
		services.NewFeesService,
		services.NewRoleService,
		services.NewMenuService,

		controller.NewUserController,
		controller.NewInstituteController,
		controller.NewDepartmentController,
		controller.NewFacultyController,
		controller.NewStudentController,
		controller.NewFeesController,
		controller.NewRoleController,
		controller.NewMenuController,

		routes.NewApp,
	)

	return nil, nil
}
