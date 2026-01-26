package domain

import "time"

type Student struct {
	ID                 int       `db:"id" json:"id"`
	Email              string    `db:"email" json:"email"`
	PasswordHash       string    `db:"password_hash" json:"-"`
	FirstName          string    `db:"first_name" json:"first_name"`
	LastName           string    `db:"last_name" json:"last_name"`
	StudentID          string    `db:"student_id" json:"student_id"`
	YearOfStudy        int       `db:"year_of_study" json:"year_of_study"`
	TotalCreditsEarned int       `db:"total_credits_earned" json:"total_credits_earned"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	StudentID   string `json:"student_id" validate:"required"`
	YearOfStudy int    `json:"year_of_study" validate:"required,min=1,max=5"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token   string  `json:"token"`
	Student Student `json:"student"`
}
