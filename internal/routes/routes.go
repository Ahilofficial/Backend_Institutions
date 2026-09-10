package routes

import (
	"backend_institutions/internal/constants"
	"backend_institutions/internal/controller"
	"backend_institutions/internal/middleware"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func NewApp(
	userController *controller.UserController,
	instituteController *controller.InstituteController,
	departmentController *controller.DepartmentController,
	facultyController *controller.FacultyController,
	studentController *controller.StudentController,
	roleController *controller.RoleController,
	paymentController *controller.PaymentController,
) *fiber.App {
	app := fiber.New()
	RegisterRoutes(
		app,
		userController,
		instituteController,
		departmentController,
		facultyController,
		studentController,
		roleController,
		paymentController,
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
	roleController *controller.RoleController,
	paymentController *controller.PaymentController,
) {

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

	protected.Get("/profile", userController.GetProfile)
	// protected.Post("/users/assign-role", middleware.RequirePermission(constants.PermissionAssignRoles), userController.AssignRoleController)

	roleRoute := protected.Group("/roles", middleware.RequirePermission(constants.PermissionAssignRoles))
	roleRoute.Post("", roleController.CreateRoleController)
	roleRoute.Post("/:id/permissions<min(1)>", roleController.AssignPermissionsController)
	roleRoute.Get("/roleperms", roleController.FetchAllRoles)
	roleRoute.Get("/:id/permissions<min(1)>", roleController.GetRolePermissionsController)
	roleRoute.Get("/:id<min(1)>", roleController.GetRoleByIDController)
	roleRoute.Put("/:id<min(1)>", roleController.UpdateRoleController)
	roleRoute.Delete("/:id<min(1)>", roleController.DeleteRoleController)
	roleRoute.Delete("/:id/permissions/:permissionId<min(1)>", roleController.RemovePermissionController)

	protected.Get("/permissions/:id<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetPermissionByIDController)
	protected.Delete("/permissions/:id<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeletePermissionController)

	userRolesRoute := protected.Group("/user-roles", middleware.RequirePermission(constants.PermissionAssignRoles))
	userRolesRoute.Get("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.FetchUserRolesController)
	userRolesRoute.Post("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.CreateUserRoleController)
	userRolesRoute.Get("/user/:userId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetUserRolesByUserIDController)
	userRolesRoute.Get("/:userId/:roleId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetUserRoleByIDController)
	userRolesRoute.Put("/:userId/:roleId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.UpdateUserRoleController)
	userRolesRoute.Delete("/:userId/:roleId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeleteUserRoleController)

	rolePermsRoute := protected.Group("/role-permissions", middleware.RequirePermission(constants.PermissionAssignRoles))
	rolePermsRoute.Get("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.FetchRolePermissionsController)
	rolePermsRoute.Post("", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.CreateRolePermissionController)
	rolePermsRoute.Get("/:roleId/:permissionId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.GetRolePermissionByIDController)
	rolePermsRoute.Put("/:roleId/:permissionId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.UpdateRolePermissionController)
	rolePermsRoute.Delete("/:roleId/:permissionId<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), roleController.DeleteRolePermissionController)
    
	

	InstituteRoute := protected.Group("/institutes")
	InstituteRoute.Post("", middleware.RequirePermission(constants.PermissionCreateInstitutes), instituteController.CreateInstituteController)
	InstituteRoute.Get("", middleware.RequirePermission(constants.PermissionViewInstitutes), instituteController.GetAllInstitutesController)
	InstituteRoute.Get("/:id<min(1)>", middleware.RequirePermission(constants.PermissionViewIDInstitutes), instituteController.GetInstituteByIDController)
	InstituteRoute.Put("/:id<min(1)>", middleware.RequirePermission(constants.PermissionUpdateInstitutes), instituteController.UpdateInstituteController)
	InstituteRoute.Delete("/:id<min(1)>", middleware.RequirePermission(constants.PermissionDeleteInstitutes), instituteController.DeleteInstituteController)

	DepartmentRoute := protected.Group("/departments")
	DepartmentRoute.Post("", middleware.RequirePermission(constants.PermissionCreateDepartments), departmentController.CreateDepartmentController)
	DepartmentRoute.Get("", middleware.RequirePermission(constants.PermissionViewDepartments), departmentController.GetAllDepartmentsController)
	DepartmentRoute.Get("/:id<min(1)>", middleware.RequirePermission(constants.PermissionViewIDDepartments), departmentController.GetDepartmentByIDController)
	DepartmentRoute.Put("/:id<min(1)>", middleware.RequirePermission(constants.PermissionUpdateDepartments), departmentController.UpdateDepartmentController)
	DepartmentRoute.Delete("/:id<min(1)>", middleware.RequirePermission(constants.PermissionDeleteDepartments), departmentController.DeleteDepartmentController)

	FacultyRoute := protected.Group("/faculties")
	FacultyRoute.Post("", middleware.RequirePermission(constants.PermissionCreateFaculties), facultyController.CreateFacultyController)
	FacultyRoute.Get("", middleware.RequirePermission(constants.PermissionViewFaculties), facultyController.GetAllFacultiesController)
	FacultyRoute.Get("/loginfaculty/students", middleware.RequirePermission(constants.PermissionFacultyViewStudents), facultyController.GetLoggedInFacultyStudentsController)
	FacultyRoute.Get("/student/paidstudents", middleware.RequirePermission(constants.PermissionFacultyViewStudents), facultyController.GetPaidStudentsForFacultyController)
	FacultyRoute.Get("/student/nonpaidstudents", middleware.RequirePermission(constants.PermissionFacultyViewStudents), facultyController.GetNonPaidStudentsForFacultyController)
	FacultyRoute.Get("/:id<min(1)>", middleware.RequirePermission(constants.PermissionViewIDFaculties), facultyController.GetFacultyByIDController)
	FacultyRoute.Put("/:id<min(1)>", middleware.RequirePermission(constants.PermissionUpdateFaculties), facultyController.UpdateFacultyController)
	FacultyRoute.Delete("/:id<min(1)>", middleware.RequirePermission(constants.PermissionDeleteFaculties), facultyController.DeleteFacultyController)

	StudentRoute := protected.Group("/students")
	StudentRoute.Post("", middleware.RequirePermission(constants.PermissionCreateStudents), studentController.CreateStudentControllers)
	StudentRoute.Get("", middleware.RequirePermission(constants.PermissionViewStudents), studentController.FetchAllStudentsPaginatedControllers)
	StudentRoute.Get("/:id<min(1)>", middleware.RequirePermission(constants.PermissionViewStudentsID), studentController.GetStudentByIDControllers)
	StudentRoute.Put("/:id<min(1)>" , middleware.RequirePermission(constants.PermissionUpdateStudents), studentController.UpdateStudentController)
	StudentRoute.Patch("/:id<min(1)>", middleware.RequirePermission(constants.PermissionUpdateStudents), studentController.UpdateStudentSemesterController)
	StudentRoute.Delete("/:id<min(1)>", middleware.RequirePermission(constants.PermissionDeleteStudents), studentController.DeleteStudentControllers)

	// val:=app.RegisterCustomConstraint
	
	userRoute := protected.Group("/users")
	userRoute.Delete("/:id<min(1)>", middleware.RequirePermission(constants.PermissionAssignRoles), userController.DeleteUserController)

	deptPaymentRoute := protected.Group("/department-payments",middleware.RequirePermission(constants.PaymentCreation))
	deptPaymentRoute.Post("",  paymentController.CreateDepartmentPaymentController)
	deptPaymentRoute.Put("/:id<min(1)>",  paymentController.UpdateDepartmentPaymentController)
	deptPaymentRoute.Delete("/:id<min(1)>",  paymentController.DeleteDepartmentPaymentController)
	deptPaymentRoute.Get("/department/:departmentId/semester/:semester<min(1)>",  paymentController.GetDepartmentPaymentBySemesterController)
	deptPaymentRoute.Get("/department/:departmentId<min(1)>",  paymentController.GetDepartmentPaymentsController)

	studentPaymentRoute := protected.Group("/payments")
	studentPaymentRoute.Post("/student", middleware.RequirePermission(constants.StudentPayments), paymentController.CreateStudentPaymentController)

	fmt.Println("All routes registered successfully")
}
