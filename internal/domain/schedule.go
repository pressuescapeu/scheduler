package domain

import "time"

type Schedule struct {
	ID           int       `db:"id" json:"id"`
	StudentID    int       `db:"student_id" json:"student_id"`
	ScheduleName string    `db:"schedule_name" json:"schedule_name"`
	Description  *string   `db:"description" json:"description"`
	IsSubmitted  bool      `db:"is_submitted" json:"is_submitted"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type ScheduleSection struct {
	ID         int       `db:"id" json:"id"`
	ScheduleID int       `db:"schedule_id" json:"schedule_id"`
	SectionID  int       `db:"section_id" json:"section_id"`
	MeetingID  *int      `db:"meeting_id" json:"meeting_id"`
	AddedAt    time.Time `db:"added_at" json:"added_at"`
}

type ScheduleWithSections struct {
	Schedule
	Sections     []SectionWithDetails `json:"sections"`
	TotalCredits int                  `json:"total_credits"`
}

type CreateScheduleRequest struct {
	ScheduleName string  `json:"schedule_name" validate:"required"`
	Description  *string `json:"description"`
}

type AddSectionRequest struct {
	SectionID int  `json:"section_id" validate:"required"`
	MeetingID *int `json:"meeting_id"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
}
