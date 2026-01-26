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

func GetCourses(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		semester := c.QueryParam("semester")
		if semester == "" {
			semester = "Summer 2025"
		}

		courses, err := storage.GetAllCourses(c.Request().Context(), semester)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch courses"})
		}

		return c.JSON(http.StatusOK, courses)
	}
}

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
