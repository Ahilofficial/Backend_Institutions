package constants

const (
	PermissionAdminPermission  = "ADMIN_PERMISSION"
	PermissionCreateInstitutes = "CREATE_INSTITUTION"
	PermissionViewInstitutes   = "VIEW_INSTITUTIONS"
	PermissionUpdateInstitutes = "UPDATE_INSTITUTION"
	PermissionDeleteInstitutes = "DELETE_INSTITUTION"
	PermissionViewIDInstitutes = "VIEW_INSTITUTION_ID"

	PermissionCreateDepartments = "CREATE_DEPARTMENT"
	PermissionViewDepartments   = "VIEW_DEPARTMENTS"
	PermissionUpdateDepartments = "UPDATE_DEPARTMENT"
	PermissionDeleteDepartments = "DELETE_DEPARTMENT"
	PermissionViewIDDepartments = "VIEW_DEPARTMENT_ID"

	PermissionCreateFaculties = "CREATE_FACULTY"
	PermissionViewFaculties   = "VIEW_FACULTIES"
	PermissionUpdateFaculties = "UPDATE_FACULTY"
	PermissionDeleteFaculties = "DELETE_FACULTY"
	PermissionViewIDFaculties = "VIEW_FACULTY_ID"

	PermissionCreateStudents = "CREATE_STUDENT"
	PermissionViewStudents   = "VIEW_STUDENTS"
	PermissionUpdateStudents = "UPDATE_STUDENT"
	PermissionDeleteStudents = "DELETE_STUDENT"
	PermissionViewStudentsID = "VIEW_STUDENT_ID"
	PermissionPromoteStudent    = "PROMOTE_STUDENT"
	PermissionManageInstitution = "institution.manage"

	PermissionAssignRoles         = "ASSIGN_ROLE"
	PermissionFacultyViewStudents = "PermissionFacultyViewStudents"
)

var AllPermissions = []string{
	PermissionCreateInstitutes,
	PermissionViewInstitutes,
	PermissionUpdateInstitutes,
	PermissionDeleteInstitutes,
	PermissionViewIDInstitutes,

	PermissionCreateDepartments,
	PermissionViewDepartments,
	PermissionUpdateDepartments,
	PermissionDeleteDepartments,
	PermissionViewIDDepartments,

	PermissionCreateFaculties,
	PermissionViewFaculties,
	PermissionUpdateFaculties,
	PermissionDeleteFaculties,
	PermissionViewIDFaculties,

	PermissionCreateStudents,
	PermissionViewStudents,
	PermissionUpdateStudents,
	PermissionDeleteStudents,
	PermissionViewStudentsID,
	PermissionPromoteStudent,

	PermissionAdminPermission,
	PermissionManageInstitution,

	PermissionAssignRoles,

	PermissionFacultyViewStudents,
}
