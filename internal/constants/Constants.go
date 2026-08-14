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
	PermissionViewIDFaculties= "VIEW_FACULTY_ID"

	PermissionCreatePrincipals = "CREATE_PRINCIPAL"
	PermissionViewPrincipals   = "VIEW_PRINCIPALS"
	PermissionUpdatePrincipals = "UPDATE_PRINCIPAL"
	PermissionDeletePrincipals = "DELETE_PRINCIPAL"
	PermissionViewIDPrincipals= "VIEW_PRINCIPAL_ID"

	PermissionCreateStudents = "CREATE_STUDENT"
	PermissionViewStudents   = "VIEW_STUDENTS"
	PermissionUpdateStudents = "UPDATE_STUDENT"
	PermissionDeleteStudents = "DELETE_STUDENT"
	PermissionViewStudentsID= "VIEW_STUDENT_ID"
	StudentMonth = "VIEW_STUDENT_MONTH"

	PermissionCreatePayments = "CREATE_PAYMENTS"
	PermissionViewPayments   = "VIEW_PAYMENTS"
	PermissionUpdatePayments = "UPDATE_PAYMENTS"
	PermissionDeletePayments = "DELETE_PAYMENTS"
	PermissionViewIDPayments ="VIEW_ID_PAYMENTS"
	PermissionCreatePayment= "CREATE_PAYMENT"

	PermissionCreateFees = "CREATE_FEE"
	PermissionViewFees   = "VIEW_FEES"
	PermissionUpdateFees = "UPDATE_FEE"
	PermissionDeleteFees = "DELETE_FEE"

	PermissionManageInstitution = "institution.manage"

	PermissionAssignRoles = "ASSIGN_ROLE"
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

	PermissionCreatePrincipals,
	PermissionViewPrincipals,
	PermissionUpdatePrincipals,
	PermissionDeletePrincipals,
	PermissionViewIDPrincipals,

	PermissionCreateStudents,
	PermissionViewStudents,
	PermissionUpdateStudents,
	PermissionDeleteStudents,
	PermissionViewStudentsID,
	StudentMonth,

	PermissionCreateFees,
	PermissionViewFees,
	PermissionUpdateFees,
	PermissionDeleteFees,
	PermissionViewPayments,
	PermissionAdminPermission,
	PermissionManageInstitution,
	PermissionViewIDPayments,
	PermissionCreatePayment,
	


	PermissionAssignRoles,
}
