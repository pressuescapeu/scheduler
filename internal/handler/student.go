package handler

import (
	"context"
	"net/http"
	"scheduler/internal/domain"
	"scheduler/internal/repository/postgres"
	"scheduler/internal/utils"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func SetupStudentRoutes(e *echo.Echo, storage *postgres.Storage, authMiddleware echo.MiddlewareFunc) {
	e.POST("/api/auth/register", Register(storage))
	e.POST("/api/auth/login", Login(storage))

	e.GET("/api/users/me", GetCurrentStudent(storage), authMiddleware)
}

// Login godoc
// @Summary Login student
// @Description Authenticate student and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body domain.LoginRequest true "Login credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func Login(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req domain.LoginRequest

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		student, err := storage.GetStudentByEmail(context.Background(), req.Email)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "no student found with such email"})
		}

		err = bcrypt.CompareHashAndPassword([]byte(student.PasswordHash), []byte(req.Password))

		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "wrong password"})
		}

		token, err := utils.GenerateToken(student.ID, student.Email)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		response := domain.AuthResponse{
			Token:   token,
			Student: *student,
		}

		return c.JSON(http.StatusOK, response)
	}
}

// Register godoc
// @Summary Register new student
// @Description Create a new student account
// @Tags auth
// @Accept json
// @Produce json
// @Param student body domain.RegisterRequest true "Student registration details"
// @Success 201 {object} domain.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func Register(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req domain.RegisterRequest

		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		if err := c.Validate(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		if !strings.Contains(req.Email, "@nu.edu.kz") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email format"})
		}

		if len(req.Password) < 8 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters long"})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		}

		student, err := storage.CreateStudent(context.Background(), &req, string(hashedPassword))

		if err != nil {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already exists"})
		}

		token, err := utils.GenerateToken(student.ID, student.Email)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}

		response := domain.AuthResponse{
			Token:   token,
			Student: *student,
		}

		return c.JSON(http.StatusCreated, response)
	}
}

// GetCurrentStudent godoc
// @Summary Get current student profile
// @Description Get the profile of the authenticated student
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Student
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me [get]
func GetCurrentStudent(storage *postgres.Storage) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIDContext := c.Get("user_id")
		userID, ok := userIDContext.(int)
		if !ok {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": utils.ErrValueConversion.Error()})
		}

		user, err := storage.GetStudentByID(context.Background(), userID)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		user.PasswordHash = ""

		return c.JSON(http.StatusOK, user)
	}
}
