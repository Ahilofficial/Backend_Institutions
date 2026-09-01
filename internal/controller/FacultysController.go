package controller

import (
	"backend_institutions/internal/dto"
	"backend_institutions/internal/helper"
	"backend_institutions/internal/model"
	"backend_institutions/internal/services"
	"math"
	"strconv"
	"github.com/gofiber/fiber/v3"
)

type FacultyController struct {
	facultyService   *services.FacultyService
	userService      *services.UserService
	instituteService services.InstituteService
}

func NewFacultyController(
	facultyService *services.FacultyService,
	userService *services.UserService,
	instituteService *services.InstituteService,
) *FacultyController {
	return &FacultyController{
		facultyService:   facultyService,
		userService:      userService,
		instituteService: *instituteService,
	}
}
func (cl *FacultyController) GetFacultyByIDController(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid faculty ID")
	}

	facultyID := uint(id)

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_faculty_institution_id := cl.facultyService.GetInstitutionIDForUserRepo(facultyID)
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && (checking_faculty_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	faculty, err := cl.facultyService.GetFacultyServiceById(
		userID,
		facultyID,
		
	)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"Faculty fetched successfully",
		dto.ToFacultyResponseDTO(faculty),
	)
}

func (cl *FacultyController) CreateFacultyController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok || userID == 0 {
		return helper.Error(c, 401, "user not authenticated")
	}

	
	
   
	var body dto.CreateFacultyDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(
			c,
			400,
			"invalid request body: "+err.Error(),
		)
	}

	body.Sanitize()

	if err := body.Validate(); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	user, err := cl.userService.GetUserByID(userID)
	if err != nil  {
		return helper.Error(c, 404, "user not found")
	}

	var targetUserID uint
	if user.FacultyID == 0 {
		targetUserID = userID
	}

	

	createdFaculty, err := cl.facultyService.CreateFacultyService(
		userID,
		&model.Faculty{
			Name:         body.Name,
			Gender:       body.Gender,
			JoiningDate:  body.JoiningDate,
			DepartmentID: body.DepartmentID,
			UserID:       targetUserID,
			IsActive:     true,
		},
	
	)

	if err != nil {
		return helper.Error(c, 400, err.Error())
	}

	if targetUserID != 0 {
		_ = cl.userService.UpdateFacultyID(targetUserID, createdFaculty.ID)
	}

	return helper.Success(
		c,
		"Faculty created successfully",
		dto.ToFacultyResponseDTO(&createdFaculty),
	)
}


func (cl *FacultyController) GetAllFacultiesController(c fiber.Ctx) error {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page,err := strconv.Atoi(pageStr)
	limit,err := strconv.Atoi(limitStr)

	faculties, total, err := cl.facultyService.GetFacultyServicePaginated(page,limit)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return helper.Success(
		c,
		"Faculties fetched successfully",
		fiber.Map{
			"items":       dto.ToFacultyResponseListDTO(faculties),
			"total_count": total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	)
}



func (cl *FacultyController) GetLoggedInFacultyController(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)

	faculty, err := cl.facultyService.GetLoggedInFacultyProfile(userID)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"LoggedIn faculty profile fetched successfully",
		dto.ToFacultyResponseDTO(faculty),
	)
}

func (cl *FacultyController) GetLoggedInFacultyStudentsController(c fiber.Ctx) error {
	userID:= c.Locals("user_id").(uint)

	students, err := cl.facultyService.GetLoggedInFacultyStudents(userID)
	if err != nil {
		return helper.Error(c, 404, err.Error())
	}

	return helper.Success(
		c,
		"LoggedIn faculty students fetched successfully",
		dto.ToStudentResponseListDTO(students),
	)
}



func (cl *FacultyController) UpdateFacultyController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idParam := c.Params("id")

	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "Invalid faculty ID")
	}

	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_faculty_institution_id := cl.facultyService.GetInstitutionIDForUserRepo(uint(id))
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && (loginnedUserInstitutionID == 0 || checking_faculty_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}
	
	var body dto.UpdateFacultyDTO

	if err := c.Bind().Body(&body); err != nil {
		return helper.Error(c, 400, "Invalid request body")
	}

	err = cl.facultyService.UpdateFacultyService(
		userID,
		uint(id),
		&body,
	)
	if err != nil {
		return helper.Error(c, 500, err.Error())
	}

	return helper.Success(
		c,
		"Faculty updated successfully",
		nil,
	)
}

func(cl *FacultyController) GetPaidStudentsForFacultyController(c fiber.Ctx) error {
	user_id:=c.Locals("user_id").(uint)
	faculty_id,err:=cl.facultyService.GetFacultyIDForUserService(user_id)
	paid_students,err:=cl.facultyService.GetPaidStudentsForFacultyService(faculty_id)
	if err!=nil{
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, "Paid students fetched successfully", dto.ToStudentResponseListDTO(paid_students))
}

func(cl *FacultyController) GetNonPaidStudentsForFacultyController(c fiber.Ctx) error {
	user_id:=c.Locals("user_id").(uint)
	faculty_id,err:=cl.facultyService.GetFacultyIDForUserService(user_id)
	non_paid_students,err:=cl.facultyService.GetNonPaidStudentsForFacultyService(faculty_id)
	if err!=nil{
		return helper.Error(c, 500, err.Error())
	}
	return helper.Success(c, "Non-paid students fetched successfully", dto.ToStudentResponseListDTO(non_paid_students))
}

func (cl *FacultyController) DeleteFacultyController(c fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return helper.Error(c, 401, "Invalid user")
	}

	idStr := c.Params("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return helper.Error(c, 400, "invalid faculty id")
	}
	is_inst_admin := cl.instituteService.IsInstAdminService(userID)
	checking_faculty_institution_id := cl.facultyService.GetInstitutionIDForUserRepo(uint(id))
	loginnedUserInstitutionID := cl.instituteService.GetInstitutionIDForUserService(userID)
	if is_inst_admin && (checking_faculty_institution_id != loginnedUserInstitutionID) {
		return helper.Error(c, 403, "Cant able to access other institution")
	}

	if err := cl.facultyService.DeleteFacultyService(
		userID,
		uint(id),
	); err != nil {
		return helper.Error(c, 400, err.Error())
	}

	return helper.Success(
		c,
		"Faculty deleted successfully",
		nil,
	)
}
