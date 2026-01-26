package domain

import "time"

type Course struct {
	ID           int       `db:"id" json:"id"`
	CourseCode   string    `db:"course_code" json:"course_code"`
	CourseName   string    `db:"course_name" json:"course_name"`
	Credits      int       `db:"credits" json:"credits"`
	IsInternship bool      `db:"is_internship" json:"is_internship"`
	Description  *string   `db:"description" json:"description"`
	Semester     string    `db:"semester" json:"semester"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type Section struct {
	ID              int    `db:"id" json:"id"`
	CourseID        int    `db:"course_id" json:"course_id"`
	SectionNumber   string `db:"section_number" json:"section_number"`
	SectionType     string `db:"section_type" json:"section_type"`
	ProfessorID     *int   `db:"professor_id" json:"professor_id"`
	TotalSeats      int    `db:"total_seats" json:"total_seats"`
	AvailableSeats  int    `db:"available_seats" json:"available_seats"`
	ParentSectionID *int   `db:"parent_section_id" json:"parent_section_id"`
}

type SectionMeeting struct {
	ID        int     `db:"id" json:"id"`
	SectionID int     `db:"section_id" json:"section_id"`
	DayOfWeek string  `db:"day_of_week" json:"day_of_week"`
	StartTime string  `db:"start_time" json:"start_time"`
	EndTime   string  `db:"end_time" json:"end_time"`
	Room      *string `db:"room" json:"room"`
	Building  *string `db:"building" json:"building"`
}

// Composite types for API responses

type SectionWithDetails struct {
	Section
	Course        Course           `json:"course"`
	Professor     *Professor       `json:"professor"`
	Meetings      []SectionMeeting `json:"meetings"`
	ChildSections []Section        `json:"child_sections,omitempty"` // Labs, Recitations
}

type CourseWithSections struct {
	Course
	Sections []SectionWithDetails `json:"sections"`
}
