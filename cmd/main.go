package main

import (
	"net/http"
	"os"
	"scheduler/internal/handler"
	"scheduler/internal/middleware"
	"scheduler/internal/repository/postgres"

	_ "scheduler/docs"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (customValidator *CustomValidator) Validate(i interface{}) error {
	if err := customValidator.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// @title Student Schedule API
// @version 1.0
// @description API for managing student schedules and course registration

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /api
// @schemes https http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	godotenv.Load()
	e := echo.New()

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.Validator = &CustomValidator{validator: validator.New()}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		panic("DATABASE_URL not set")
	}

	storage, err := postgres.NewConnection(connString)
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	handler.SetupCourseRoutes(e, storage)

	authMiddleware := middleware.JWTAuth()
	handler.SetupStudentRoutes(e, storage, authMiddleware)
	handler.SetupScheduleRoutes(e, storage, authMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
