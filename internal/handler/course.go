package handler

import (
	"net/http"
	"scheduler/internal/repository/postgres"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupCourseRoutes(e *echo.Echo, storage *postgres.Storage) {
	e.GET("/api/courses", GetCourses(storage))
	e.GET("/api/courses/:id", GetCourseByID(storage))
	e.GET("/api/courses/:id/sections", GetCourseSections(storage))
}

// GetCourses godoc
// @Summary Get all courses
// @Description Get all courses, optionally filtered by semester
// @Tags courses
// @Accept json
// @Produce json
// @Param semester query string false "Semester to filter by (e.g., 'Summer 2025'). If omitted, returns all courses."
// @Success 200 {array} domain.Course
// @Failure 500 {object} map[string]string
// @Router /courses [get]
func GetCourses(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		semester := c.QueryParam("semester")

		courses, err := storage.GetAllCourses(c.Request().Context(), semester)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch courses"})
		}

		return c.JSON(http.StatusOK, courses)
	}
}

// GetCourseByID godoc
// @Summary Get course by ID
// @Description Get detailed information about a specific course
// @Tags courses
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {object} domain.Course
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /courses/{id} [get]
func GetCourseByID(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid course id"})
		}

		course, err := storage.GetCourseByID(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "course not found"})
		}

		return c.JSON(http.StatusOK, course)
	}
}

// GetCourseSections godoc
// @Summary Get all sections for a course
// @Description Get all sections (lectures, labs, recitations) for a specific course
// @Tags courses
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {array} domain.SectionWithDetails
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /courses/{id}/sections [get]
func GetCourseSections(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid course id"})
		}

		sections, err := storage.GetSectionsForCourse(c.Request().Context(), id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch sections"})
		}

		return c.JSON(http.StatusOK, sections)
	}
}
