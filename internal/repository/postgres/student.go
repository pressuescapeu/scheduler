package postgres

import (
	"context"
	"scheduler/internal/domain"
)

func (s *Storage) CreateStudent(ctx context.Context, req *domain.RegisterRequest, passwordHash string) (*domain.Student, error) {
	const query = `
        INSERT INTO students (email, password_hash, first_name, last_name, student_id, year_of_study)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, email, first_name, last_name, student_id, year_of_study, total_credits_earned, created_at;
    `

	var student domain.Student
	err := s.pool.QueryRow(ctx, query,
		req.Email, passwordHash, req.FirstName, req.LastName, req.StudentID, req.YearOfStudy,
	).Scan(
		&student.ID, &student.Email, &student.FirstName, &student.LastName,
		&student.StudentID, &student.YearOfStudy, &student.TotalCreditsEarned, &student.CreatedAt,
	)

	return &student, err
}

func (s *Storage) GetStudentByEmail(ctx context.Context, email string) (*domain.Student, error) {
	const query = `
        SELECT id, email, password_hash, first_name, last_name, student_id, year_of_study, total_credits_earned, created_at
        FROM students WHERE email = $1;
    `

	var student domain.Student
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&student.ID, &student.Email, &student.PasswordHash, &student.FirstName,
		&student.LastName, &student.StudentID, &student.YearOfStudy,
		&student.TotalCreditsEarned, &student.CreatedAt,
	)

	return &student, err
}

func (s *Storage) GetStudentByID(ctx context.Context, id int) (*domain.Student, error) {
	const query = `
        SELECT id, email, first_name, last_name, student_id, year_of_study, total_credits_earned, created_at
        FROM students WHERE id = $1;
    `

	var student domain.Student
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&student.ID, &student.Email, &student.FirstName, &student.LastName,
		&student.StudentID, &student.YearOfStudy, &student.TotalCreditsEarned, &student.CreatedAt,
	)

	return &student, err
}
