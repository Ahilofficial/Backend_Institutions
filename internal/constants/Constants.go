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

	PermissionCreatePayments = "CREATE_PAYMENTS"
	PermissionViewPayments   = "VIEW_PAYMENTS"
	PermissionUpdatePayments = "UPDATE_PAYMENTS"
	PermissionDeletePayments = "DELETE_PAYMENTS"
	PermissionViewIDPayments = "VIEW_ID_PAYMENTS"
	PermissionCreatePayment  = "CREATE_PAYMENT"

	PermissionCreateFees = "CREATE_FEE"
	PermissionViewFees   = "VIEW_FEES"
	PermissionUpdateFees = "UPDATE_FEE"
	PermissionDeleteFees = "DELETE_FEE"
	PermissionViewIDFees = "VIEW_ID_FEES"

	PermissionCreateSemesterFee = "CREATE_SEMESTER_FEE"
	PermissionViewSemesterFee   = "VIEW_SEMESTER_FEE"
	PermissionUpdateSemesterFee = "UPDATE_SEMESTER_FEE"
	PermissionDeleteSemesterFee = "DELETE_SEMESTER_FEE"
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

	PermissionCreateFees,
	PermissionViewFees,
	PermissionUpdateFees,
	PermissionDeleteFees,
	PermissionViewIDFees,

	PermissionCreateSemesterFee,
	PermissionViewSemesterFee,
	PermissionUpdateSemesterFee,
	PermissionDeleteSemesterFee,

	PermissionViewPayments,
	PermissionAdminPermission,
	PermissionManageInstitution,
	PermissionViewIDPayments,
	PermissionCreatePayment,

	PermissionAssignRoles,

	PermissionFacultyViewStudents,
}
