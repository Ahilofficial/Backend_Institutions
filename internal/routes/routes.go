package routes

import (
	"backend_institutions/internal/constants"
	"backend_institutions/internal/controller"
	"backend_institutions/internal/middleware"
	"backend_institutions/internal/repository"
	"backend_institutions/internal/services"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func NewApp(
	userController *controller.UserController,
	instituteController *controller.InstituteController,
	departmentController *controller.DepartmentController,
	facultyController *controller.FacultyController,
	principalController *controller.PrincipalControllers,
	studentController *controller.StudentController,
	departmentService *services.DepartmentService,
	facultyService *services.FacultyService,
	principalService *services.PrincipalService,
	studentService *services.StudentService,
	feesController *controller.FeesController,
	roleController *controller.RoleController,
	permissionController *controller.PermissionController,
	menuController *controller.MenuController,
	studentRepo *repository.StudentRepository,
	facultyRepo *repository.FacultyRepository,
	principalRepo *repository.PrincipalRepository,
	departmentRepo *repository.DepartmentRepository,
	feeRepo *repository.FeesRepository,
	userRepo *repository.UserRepository,
) *fiber.App {
	app := fiber.New()
	RegisterRoutes(
		app,
		userController,
		instituteController,
		departmentController,
		facultyController,
		principalController,
		studentController,
		studentService,
		departmentService,
		facultyService,
		principalService,
		feesController,
		roleController,
		permissionController,
		menuController,
		studentRepo,
		facultyRepo,
		principalRepo,
		departmentRepo,
		feeRepo,
		userRepo,
	)
	return app
}

func RegisterRoutes(
	app *fiber.App,
	userController *controller.UserController,
	instituteController *controller.InstituteController,
	departmentController *controller.DepartmentController,
	facultyController *controller.FacultyController,
	principalController *controller.PrincipalControllers,
	studentController *controller.StudentController,
	studentService *services.StudentService,
	departmentService *services.DepartmentService,
	facultyService *services.FacultyService,
	principalService *services.PrincipalService,
	feesController *controller.FeesController,
	roleController *controller.RoleController,
	permissionController *controller.PermissionController,
	menuController *controller.MenuController,
	studentRepo *repository.StudentRepository,
	facultyRepo *repository.FacultyRepository,
	principalRepo *repository.PrincipalRepository,
	departmentRepo *repository.DepartmentRepository,
	feeRepo *repository.FeesRepository,
	userRepo *repository.UserRepository,
) {
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:4200"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	app.Post("/signup/:role", userController.SignUpController)
	app.Post("/signup", userController.SignUpController)
	app.Post("/signin", userController.SignInController)
	app.Post("/logout", userController.Logout)
	app.Post("/resendmail", userController.ResendMail)

   

	app.Get("/auth/verify", userController.VerifyEmail)
	app.Post("/auth/forgot-password", userController.ForgotPassword)
	app.Post("/auth/reset-password", userController.ResetPassword)
	app.Get("/roles", roleController.FetchRoles)
	app.Get("/permission", roleController.FetchPermissions)

	

	app.Get("/menus", middleware.AuthRequired(), menuController.GetMenus)
	app.Use(middleware.RequestResponseLogger())

	protected := app.Group("", middleware.AuthRequired())

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

	// Permission CRUD
	protected.Get("/permissions/:id", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetPermissionByIDController)
	protected.Delete("/permissions/:id", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeletePermissionController)

	// UserRole CRUD
	userRolesRoute := protected.Group("/user-roles", middleware.RequirePermission(constants.PermissionAssignRoles))
	userRolesRoute.Get("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.FetchUserRolesController)
	userRolesRoute.Post("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.CreateUserRoleController)
	userRolesRoute.Get("/user/:userId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetUserRolesByUserIDController)
	userRolesRoute.Get("/:userId/:roleId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetUserRoleByIDController)
	userRolesRoute.Put("/:userId/:roleId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.UpdateUserRoleController)
	userRolesRoute.Delete("/:userId/:roleId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeleteUserRoleController)

	// RolePermission CRUD
	rolePermsRoute := protected.Group("/role-permissions", middleware.RequirePermission(constants.PermissionAssignRoles))
	rolePermsRoute.Get("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.FetchRolePermissionsController)
	rolePermsRoute.Post("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.CreateRolePermissionController)
	rolePermsRoute.Get("/:roleId/:permissionId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetRolePermissionByIDController)
	rolePermsRoute.Put("/:roleId/:permissionId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.UpdateRolePermissionController)
	rolePermsRoute.Delete("/:roleId/:permissionId", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeleteRolePermissionController)

	InstituteRoute := protected.Group("/institutes")
	InstituteRoute.Post("", middleware.RequirePermission(constants.PermissionCreateInstitutes), instituteController.CreateInstituteController)
	InstituteRoute.Get("", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.GetAllInstitutesController)
	InstituteRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.FetchAllInstitutesController)
	InstituteRoute.Get("/active", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.GetActiveInstituteController)
	InstituteRoute.Get("/inactive", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.GetInactiveInstituteController)
	InstituteRoute.Get("/deleted", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.GetDeletedInstitutesController)
	InstituteRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDInstitutes), instituteController.GetInstituteByIDController)
	InstituteRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateInstitutes), instituteController.UpdateInstituteController)
	InstituteRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteInstitutes), instituteController.DeleteInstituteController)

	// for principals

	PrincipalRoute := protected.Group("/principals")
	PrincipalRoute.Post("", middleware.RequirePermission(constants.PermissionCreatePrincipals), principalController.CreatePrincipalController)
	
	PrincipalRoute.Get("", middleware.RequirePermission(constants.PermissionViewPrincipals), principalController.GetAllPrincipalsController)
	PrincipalRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewPrincipals), principalController.FetchAllPrincipalsController)
	PrincipalRoute.Get("/active", middleware.RequirePermission(constants.PermissionViewPrincipals), principalController.GetActivePrincipalController)
	PrincipalRoute.Get("/inactive", middleware.RequirePermission(constants.PermissionViewPrincipals), principalController.GetInactivePrincipalController)
	PrincipalRoute.Get("/deleted", middleware.RequirePermission(constants.PermissionViewPrincipals), principalController.GetDeletedPrincipalsController)
	PrincipalRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDPrincipals), principalController.GetPrincipalByIDController)
	PrincipalRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdatePrincipals), principalController.UpdatePrincipalController)
	PrincipalRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeletePrincipals), principalController.DeletePrincipalController)

	//for departments

	DepartmentRoute := protected.Group("/departments")
	DepartmentRoute.Post("", middleware.RequirePermission(constants.PermissionCreateDepartments), departmentController.CreateDepartmentController)

	DepartmentRoute.Get("", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.GetAllDepartmentsController)
	DepartmentRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.FetchAllDepartmentsController)
	DepartmentRoute.Get("/active", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.GetActiveDepartmentController)
	DepartmentRoute.Get("/inactive", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.GetInactiveDepartmentController)
	DepartmentRoute.Get("/deleted", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.GetDeletedDepartmentsController)
	DepartmentRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDDepartments), departmentController.GetDepartmentByIDController)
	DepartmentRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateDepartments), departmentController.UpdateDepartmentController)
	DepartmentRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteDepartments), departmentController.DeleteDepartmentController)

	//for faculty

	FacultyRoute := protected.Group("/faculties")
	FacultyRoute.Post("", middleware.RequirePermission(constants.PermissionCreateFaculties), facultyController.CreateFacultyController)
	
	FacultyRoute.Get("", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetAllFacultiesController)
	FacultyRoute.Get("/loginfaculty", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetLoggedInFacultyController)
	FacultyRoute.Get("/loginfaculty/students", middleware.RequirePermission(constants.PermissionViewStudents), facultyController.GetLoggedInFacultyStudentsController)
	FacultyRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.FetchAllFacultiesController)
	FacultyRoute.Get("/active", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetActiveFacultyController)
	FacultyRoute.Get("/inactive", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetInactiveFacultyController)
	FacultyRoute.Get("/deleted", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetDeletedFacultiesController)
	FacultyRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDFaculties), facultyController.GetFacultyByIDController)
	FacultyRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateFaculties), facultyController.UpdateFacultyController)
	FacultyRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteFaculties), facultyController.DeleteFacultyController)

	

	// student routes

	StudentRoute := protected.Group("/students")
	StudentRoute.Post("", middleware.RequirePermission(constants.PermissionCreateStudents), studentController.CreateStudentControllers)
	
	StudentRoute.Get("", middleware.RequirePermission(constants.PermissionViewStudents), studentController.FetchAllStudentsPaginatedControllers)
	StudentRoute.Get("/loginstudents", middleware.RequirePermission(constants.PermissionViewStudents), studentController.GetLoggedInStudentController)
	StudentRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewStudents), studentController.FetchAllStudentsControllers)
	StudentRoute.Get("/active", middleware.RequirePermission(constants.PermissionViewStudents), studentController.GetActiveStudentController)
	StudentRoute.Get("/inactive", middleware.RequirePermission(constants.PermissionViewStudents), studentController.GetInactiveStudentController)
	StudentRoute.Get("/payment-month", middleware.RequirePermission(constants.StudentMonth), studentController.FetchStudentsByPaymentMonth)
	StudentRoute.Get("/not-paid-month", middleware.RequirePermission(constants.StudentMonth), studentController.FetchStudentsNotPaidByMonth)
	StudentRoute.Get("/faculty/paid", middleware.RequirePermission(constants.StudentMonth), studentController.FetchFacultyPaidStudents)
	StudentRoute.Get("/faculty/not-paid", middleware.RequirePermission(constants.StudentMonth), studentController.FetchFacultyUnpaidStudents)
	StudentRoute.Get("/paid", middleware.RequirePermission(constants.StudentMonth), studentController.FetchPaidStudents)
	StudentRoute.Get("/not-paid", middleware.RequirePermission(constants.StudentMonth), studentController.FetchNotPaidStudents)
	StudentRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewStudentsID), studentController.GetStudentByIDControllers)
	StudentRoute.Put("/:id", middleware.RequirePermission(constants.PermissionUpdateStudents), studentController.UpdateStudentControllers)
	StudentRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionDeleteStudents), studentController.DeleteStudentControllers)

	FeesRoute := protected.Group("/fees")

	FeesRoute.Post("", middleware.RequirePermission(constants.PermissionCreatePayment), feesController.CreateFeesController)
	FeesRoute.Get("", middleware.RequirePermission(constants.PermissionViewPayments), feesController.GetAllFeesController)
	FeesRoute.Get("/all", middleware.RequirePermission(constants.PermissionViewPayments), feesController.FetchAllFeesController)
	FeesRoute.Get("/inactive", middleware.RequirePermission(constants.PermissionViewPayments), feesController.GetInactiveFeesController)
	FeesRoute.Get("/:id", middleware.RequirePermission(constants.PermissionViewIDPayments), feesController.GetFeesByIDController)
	FeesRoute.Put("/:id", middleware.RequirePermission(constants.PermissionViewIDPayments), feesController.UpdateFeesController)
	FeesRoute.Delete("/:id", middleware.RequirePermission(constants.PermissionViewPayments), feesController.DeleteFeesController)
	FeesRoute.Post("/payment", middleware.RequirePermission(constants.PermissionCreatePayment), feesController.CreatePayment)
	FeesRoute.Get("/payment/:id", middleware.RequirePermission(constants.PermissionViewPayments), feesController.GetPaymentByIDController)
	FeesRoute.Get("/:fee_id/payments", middleware.RequirePermission(constants.PermissionViewPayments), feesController.GetPaymentByFeeIDController)
	FeesRoute.Get("/student/:id", middleware.RequirePermission(constants.PermissionViewFees), feesController.FetchFeesByStudentID)

	userRoute := protected.Group("/users")
	userRoute.Delete("/:id", userController.DeleteUserController)

	fmt.Println("All routes registered successfully")
}
