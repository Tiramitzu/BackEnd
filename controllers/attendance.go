package controllers

import (
	"attendit/backend/models"
	db "attendit/backend/models/db"
	"attendit/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"
)

// AttendanceCheckIn godoc
// @Summary      AttendanceCheckIn
// @Description  checks in the user
// @Tags         attendance
// @Accept       json
// @Produce      json
// @Param        ipAddress  body  string  true  "IP Address"
// @Param        status     body  string  true  "Status"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /users/{userId}/attendances [post]
func AttendanceCheckIn(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	isIpSame := false

	var requestBody models.CheckInRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	company, err := services.GetCompany()
	if err != nil {
		response.Message = err.Error()
		response.SendErrorResponse(c)
		return
	}

	ipAddress, err := services.GetClientIP(c.Request)
	if err != nil {
		response.Message = err.Error()
		response.SendErrorResponse(c)
		return
	}

	for _, ip := range company.IPAddresses {
		if ip == ipAddress {
			isIpSame = true
			break
		}
	}

	if !isIpSame {
		response.Message = "Anda tidak diizinkan untuk melakukan absensi dari alamat IP ini."
		response.SendErrorResponse(c)
		return
	}

	userId, _ := c.Get("userId")

	loc := time.FixedZone("UTC", 7*60*60)
	currentDate := time.Now().In(loc).Format("02-01-2006")
	currentTime := time.Now().In(loc).Format("15:04:05")

	user, err := services.GetUserById(userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendErrorResponse(c)
		return
	}

	getAttendance, err := services.GetAttendanceByUserAndDate(user.ID, currentDate)
	if getAttendance != nil {
		response.Message = "Anda sudah melakukan absensi untuk hari ini."
		response.SendErrorResponse(c)
		return
	}

	attendance := db.NewAttendance(user.ID, ipAddress, currentDate, requestBody.Status, currentTime, "")
	newAttendance, err := services.AttendanceCheckIn(attendance)
	if err != nil {
		response.Message = err.Error()
		response.SendErrorResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"attendance": newAttendance}
	response.SendResponse(c)
}

// AttendanceCheckOut godoc
// @Summary      AttendanceCheckOut
// @Description  checks out the user
// @Tags         attendance
// @Accept       json
// @Produce      json
// @Param        attendanceId  path  string  true  "Attendance ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /users/{userId}/attendances/{attendanceId} [patch]
func AttendanceCheckOut(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	loc := time.FixedZone("UTC", 7*60*60)
	currentTime := time.Now().In(loc).Format("15:04:05")

	attendanceIdHex := c.Param("attendanceId")
	attendanceId, err := primitive.ObjectIDFromHex(attendanceIdHex)
	if err != nil {
		response.Message = "Error converting attendance ID"
		response.SendErrorResponse(c)
		return
	}

	attendance, _ := services.GetAttendanceById(attendanceId)
	if attendance == nil {
		response.Message = "Absensi tidak ditemukan"
		response.SendErrorResponse(c)
		return
	}
	if attendance.CheckOut != "" {
		response.Message = "Anda telah melakukan absen keluar untuk hari ini"
		response.SendErrorResponse(c)
		return
	}
	attendance.CheckOut = currentTime

	updatedAttendance, err := services.AttendanceCheckOut(attendance)
	if err != nil {
		response.Message = "Absen keluar gagal."
		response.SendErrorResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"attendance": updatedAttendance}
	response.SendResponse(c)
}
