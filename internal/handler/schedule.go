package handler

import (
	"net/http"
	"scheduler/internal/domain"
	"scheduler/internal/repository/postgres"
	"scheduler/internal/utils"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupScheduleRoutes(e *echo.Echo, storage *postgres.Storage, authMiddleware echo.MiddlewareFunc) {
	g := e.Group("/api/schedules", authMiddleware)

	g.GET("", GetMySchedules(storage))
	g.POST("", CreateSchedule(storage))
	g.GET("/:id", GetScheduleByID(storage))
	g.POST("/:id/sections", AddSectionToSchedule(storage))
	g.DELETE("/:id/sections/:sectionId", RemoveSectionFromSchedule(storage))
}

// GetMySchedules godoc
// @Summary Get all schedules for current student
// @Description Retrieve all schedules belonging to the authenticated student
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Schedule
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /schedules [get]
func GetMySchedules(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDContext := c.Get("user_id")
		userID, ok := userIDContext.(int)
		if !ok {
			return c.JSON(http.StatusBadRequest, utils.ErrValueConversion)
		}

		schedules, err := storage.GetStudentSchedules(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch schedules"})
		}

		return c.JSON(http.StatusOK, schedules)
	}
}

// CreateSchedule godoc
// @Summary Create a new schedule
// @Description Create a new schedule for the authenticated student
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param schedule body domain.CreateScheduleRequest true "Schedule details"
// @Success 201 {object} domain.Schedule
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /schedules [post]
func CreateSchedule(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID := c.Get("user_id").(int)

		var req domain.CreateScheduleRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		schedule, err := storage.CreateSchedule(c.Request().Context(), studentID, &req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create schedule"})
		}

		return c.JSON(http.StatusCreated, schedule)
	}
}

// GetScheduleByID godoc
// @Summary Get schedule by ID
// @Description Get a specific schedule with all its sections
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Success 200 {object} domain.ScheduleWithSections
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /schedules/{id} [get]
func GetScheduleByID(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID, ok := c.Get("user_id").(int)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
		}

		scheduleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule id"})
		}

		schedule, err := storage.GetScheduleWithSections(c.Request().Context(), scheduleID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "schedule not found"})
		}

		if schedule.StudentID != studentID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}

		return c.JSON(http.StatusOK, schedule)
	}
}

// AddSectionToSchedule godoc
// @Summary Add a section to schedule
// @Description Add a course section to an existing schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Param section body domain.AddSectionRequest true "Section to add"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /schedules/{id}/sections [post]
func AddSectionToSchedule(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID, ok := c.Get("user_id").(int)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
		}

		scheduleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule id"})
		}

		var req domain.AddSectionRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		schedule, err := storage.GetScheduleWithSections(c.Request().Context(), scheduleID)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "schedule not found"})
		}

		if schedule.StudentID != studentID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}

		err = storage.AddSectionToSchedule(c.Request().Context(), scheduleID, req.SectionID, req.MeetingID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to add section"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "section added"})
	}
}

// RemoveSectionFromSchedule godoc
// @Summary Remove a section from schedule
// @Description Remove a course section from an existing schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Param sectionId path int true "Section ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /schedules/{id}/sections/{sectionId} [delete]
func RemoveSectionFromSchedule(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		studentID, ok := c.Get("user_id").(int)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user context"})
		}

		scheduleID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule id"})
		}

		sectionID, err := strconv.Atoi(c.Param("sectionId"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid section id"})
		}

		schedule, err := storage.GetScheduleWithSections(c.Request().Context(), scheduleID)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "schedule not found"})
		}

		if schedule.StudentID != studentID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}

		err = storage.RemoveSectionFromSchedule(c.Request().Context(), scheduleID, sectionID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to remove section"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "section removed"})
	}
}
