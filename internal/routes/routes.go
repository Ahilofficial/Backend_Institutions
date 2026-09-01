package routes

import (
	"backend_institutions/internal/constants"
	"backend_institutions/internal/controller"
	"backend_institutions/internal/middleware"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func NewApp(
	userController *controller.UserController,
	instituteController *controller.InstituteController,
	departmentController *controller.DepartmentController,
	facultyController *controller.FacultyController,
	studentController *controller.StudentController,
	feesController *controller.FeesController,
	roleController *controller.RoleController,
	menuController *controller.MenuController,
) *fiber.App {
	app := fiber.New()
	RegisterRoutes(
		app,
		userController,
		instituteController,
		departmentController,
		facultyController,
		studentController,
		feesController,
		roleController,
		menuController,
	)
	return app
}

func RegisterRoutes(
	app *fiber.App,
	userController *controller.UserController,
	instituteController *controller.InstituteController,
	departmentController *controller.DepartmentController,
	facultyController *controller.FacultyController,
	studentController *controller.StudentController,
	feesController *controller.FeesController,
	roleController *controller.RoleController,
	menuController *controller.MenuController,
) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:4200"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	app.Use(middleware.RequestResponseLogger())

	app.Post("/signup/:role", userController.SignUpController)
	app.Post("/signin", userController.SignInController)
	app.Post("/logout", userController.Logout)
	app.Post("/resendmail", userController.ResendMail)
	app.Get("/auth/verify", userController.VerifyEmail)
	app.Post("/auth/forgot-password", userController.ForgotPassword)
	app.Post("/auth/reset-password", userController.ResetPassword)
	app.Get("/roles", roleController.FetchRoles)
	app.Get("/permission", roleController.FetchPermissions)

	protected := app.Group("", middleware.AuthRequired())

	protected.Get("/menus", menuController.GetMenus)
	protected.Get("/profile", userController.GetProfile)
	protected.Post("/users/assign-role", middleware.RequirePermission(constants.PermissionAssignRoles), userController.AssignRoleController)

	roleRoute := protected.Group("/roles", middleware.RequirePermission(constants.PermissionAssignRoles))
	roleRoute.Post("", roleController.CreateRoleController)
	roleRoute.Post("/:id/permissions", roleController.AssignPermissionsController)
	roleRoute.Get("/roleperms", roleController.FetchAllRoles)
	roleRoute.Get("/:id/permissions", roleController.GetRolePermissionsController)
	roleRoute.Get("/:id", roleController.GetRoleByIDController)
	roleRoute.Put("/:id", roleController.UpdateRoleController)
	roleRoute.Delete("/:id", roleController.DeleteRoleController)
	roleRoute.Delete("/:id/permissions/:permissionId", roleController.RemovePermissionController)

	protected.Get("/permissions/:id", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetPermissionByIDController)
	protected.Delete("/permissions/:id", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeletePermissionController)

	userRolesRoute := protected.Group("/user-roles", middleware.RequirePermission(constants.PermissionAssignRoles))
	userRolesRoute.Get("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.FetchUserRolesController)
	userRolesRoute.Post("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.CreateUserRoleController)
	userRolesRoute.Get("/user/:userId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetUserRolesByUserIDController)
	userRolesRoute.Get("/:userId/:roleId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetUserRoleByIDController)
	userRolesRoute.Put("/:userId/:roleId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.UpdateUserRoleController)
	userRolesRoute.Delete("/:userId/:roleId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeleteUserRoleController)

	rolePermsRoute := protected.Group("/role-permissions", middleware.RequirePermission(constants.PermissionAssignRoles))
	rolePermsRoute.Get("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.FetchRolePermissionsController)
	rolePermsRoute.Post("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.CreateRolePermissionController)
	rolePermsRoute.Get("/:roleId/:permissionId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetRolePermissionByIDController)
	rolePermsRoute.Put("/:roleId/:permissionId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.UpdateRolePermissionController)
	rolePermsRoute.Delete("/:roleId/:permissionId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeleteRolePermissionController)

	InstituteRoute := protected.Group("/institutes")
	InstituteRoute.Post("", middleware.RequirePermission(constants.PermissionCreateInstitutes), instituteController.CreateInstituteController)
	InstituteRoute.Get("", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.GetAllInstitutesController)
	InstituteRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDInstitutes), instituteController.GetInstituteByIDController)
	InstituteRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateInstitutes), instituteController.UpdateInstituteController)
	InstituteRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteInstitutes), instituteController.DeleteInstituteController)


	DepartmentRoute := protected.Group("/departments")
	DepartmentRoute.Post("", middleware.RequirePermission(constants.PermissionCreateDepartments), departmentController.CreateDepartmentController)
	DepartmentRoute.Get("", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.GetAllDepartmentsController)
	DepartmentRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDDepartments), departmentController.GetDepartmentByIDController)
	DepartmentRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateDepartments), departmentController.UpdateDepartmentController)
	DepartmentRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteDepartments), departmentController.DeleteDepartmentController)

	FacultyRoute := protected.Group("/faculties")
	FacultyRoute.Post("", middleware.RequirePermission(constants.PermissionCreateFaculties), facultyController.CreateFacultyController)
	FacultyRoute.Get("", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetAllFacultiesController)
	FacultyRoute.Get("/loginfaculty/students", middleware.RequirePermission(constants.PermissionFacultyViewStudents), facultyController.GetLoggedInFacultyStudentsController)
	FacultyRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDFaculties), facultyController.GetFacultyByIDController)
	FacultyRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateFaculties), facultyController.UpdateFacultyController)
	FacultyRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteFaculties), facultyController.DeleteFacultyController)
	FacultyRoute.Get("/student/paidstudents", middleware.RequirePermission(constants.PermissionFacultyViewStudents),facultyController.GetPaidStudentsForFacultyController)
	FacultyRoute.Get("/student/nonpaidstudents", middleware.RequirePermission(constants.PermissionFacultyViewStudents),facultyController.GetNonPaidStudentsForFacultyController)

	StudentRoute := protected.Group("/students")
	StudentRoute.Post("", middleware.RequirePermission(constants.PermissionCreateStudents), studentController.CreateStudentControllers)
	StudentRoute.Get("", middleware.RequirePermission(constants.PermissionViewStudents), studentController.FetchAllStudentsPaginatedControllers)
	StudentRoute.Get("/loginstudents", middleware.RequirePermission(constants.PermissionViewStudents), studentController.GetLoggedInStudentController)
	StudentRoute.Get("/active", middleware.RequirePermission(constants.PermissionViewStudents), studentController.GetActiveStudentController)
	StudentRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewStudentsID), studentController.GetStudentByIDControllers)
	StudentRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateStudents), studentController.UpdateStudentController)
	StudentRoute.Patch("/:id", studentController.UpdateStudentSemesterController)
	StudentRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteStudents), studentController.DeleteStudentControllers)

	FeesRoute := protected.Group("/fees")
	FeesRoute.Post("", middleware.RequirePermission(constants.PermissionCreateFees), feesController.CreateFeesController)
	FeesRoute.Get("/department/:departmentId/semester/:semester", middleware.RequirePermission(constants.PermissionViewFees), feesController.GetDepartmentFeeBySemesterController)
	FeesRoute.Get("/department/:departmentId", middleware.RequirePermission(constants.PermissionViewFees), feesController.GetDepartmentFeesController)
	FeesRoute.Get("", middleware.RequirePermission(constants.PermissionViewFees), feesController.GetAllFeesController)
	FeesRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewFees), feesController.FetchAllFeesController)
	FeesRoute.Get("/my-fees", middleware.RequirePermission(constants.PermissionViewFees), feesController.GetMyFeesController)
	FeesRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDFees), feesController.GetFeesByIDController)
	FeesRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateFees), feesController.UpdateFeesController)
	FeesRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteFees), feesController.DeleteFeesController)
	FeesRoute.Post("/payment", middleware.RequirePermission(constants.PermissionCreatePayment), feesController.CreatePayment)
	FeesRoute.Get("/payment/:id", middleware.RequirePermission(constants.PermissionViewPayments), feesController.GetPaymentByIDController)
	FeesRoute.Get("/:fee_id/payments", middleware.RequirePermission(constants.PermissionViewPayments), feesController.GetPaymentByFeeIDController)
	FeesRoute.Get("/student/:id", middleware.RequirePermission(constants.PermissionViewFees), feesController.FetchFeesByStudentID)

	userRoute := protected.Group("/users")
	userRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionAssignRoles), userController.DeleteUserController)

	fmt.Println("All routes registered successfully")
}
