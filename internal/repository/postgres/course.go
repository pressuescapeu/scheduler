package postgres

import (
	"context"
	"scheduler/internal/domain"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetAllCourses(ctx context.Context, semester string) ([]domain.Course, error) {
	var query string
	var rows pgx.Rows
	var err error

	if semester == "" {
		// Return all courses if no semester specified
		query = `
			SELECT id, course_code, course_name, credits, is_internship, description, semester, created_at
			FROM courses
			ORDER BY semester DESC, course_code;
		`
		rows, err = s.pool.Query(ctx, query)
	} else {
		// Filter by semester if provided
		query = `
			SELECT id, course_code, course_name, credits, is_internship, description, semester, created_at
			FROM courses
			WHERE semester = $1
			ORDER BY course_code;
		`
		rows, err = s.pool.Query(ctx, query, semester)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var courses []domain.Course
	for rows.Next() {
		var c domain.Course
		err := rows.Scan(
			&c.ID,
			&c.CourseCode,
			&c.CourseName,
			&c.Credits,
			&c.IsInternship,
			&c.Description,
			&c.Semester,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}

	return courses, nil
}

func (s *Storage) GetCourseByID(ctx context.Context, id int) (*domain.Course, error) {
	const query = `
		SELECT id, course_code, course_name, credits, is_internship, description, semester, created_at
        FROM courses WHERE id = $1;
	`

	var c domain.Course
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.CourseCode,
		&c.CourseName,
		&c.Credits,
		&c.IsInternship,
		&c.Description,
		&c.Semester,
		&c.CreatedAt,
	)

	return &c, err
}

func (s *Storage) GetSectionsForCourse(ctx context.Context, courseID int) ([]domain.SectionWithDetails, error) {
	const query = `
		SELECT s.id, s.course_id, s.section_number, s.section_type,
            s.professor_id, s.total_seats, s.available_seats, s.parent_section_id,
            c.id, c.course_code, c.course_name, c.credits,
            p.id, p.first_name, p.last_name, p.email, p.rating
        FROM sections s
        JOIN courses c ON s.course_id = c.id
        LEFT JOIN professors p ON s.professor_id = p.id
        WHERE s.course_id = $1
        ORDER BY s.section_type, s.section_number
	`

	rows, err := s.pool.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sections []domain.SectionWithDetails
	for rows.Next() {
		var sd domain.SectionWithDetails
		var course domain.Course
		var prof domain.Professor

		err := rows.Scan(
			&sd.ID, &sd.CourseID, &sd.SectionNumber, &sd.SectionType,
			&sd.ProfessorID, &sd.TotalSeats, &sd.AvailableSeats, &sd.ParentSectionID,
			&course.ID, &course.CourseCode, &course.CourseName, &course.Credits,
			&prof.ID, &prof.FirstName, &prof.LastName, &prof.Email, &prof.Rating,
		)
		if err != nil {
			return nil, err
		}

		sd.Course = course

		if sd.ProfessorID != nil {
			sd.Professor = &prof
		}

		// Get meetings for this section
		meetings, err := s.GetSectionMeetings(ctx, sd.ID)
		if err != nil {
			return nil, err
		}

		sd.Meetings = meetings

		sections = append(sections, sd)
	}

	return sections, nil
}

func (s *Storage) GetSectionMeetings(ctx context.Context, sectionID int) ([]domain.SectionMeeting, error) {
	const query = `
		SELECT id, section_id, day_of_week, start_time::text, end_time::text, room, building
		FROM section_meetings
		WHERE section_id = $1
		        ORDER BY 
            CASE day_of_week
                WHEN 'Monday' THEN 1
                WHEN 'Tuesday' THEN 2
                WHEN 'Wednesday' THEN 3
                WHEN 'Thursday' THEN 4
                WHEN 'Friday' THEN 5
                WHEN 'Saturday' THEN 6
                WHEN 'Sunday' THEN 7
            END;
	`

	rows, err := s.pool.Query(ctx, query, sectionID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var meetings []domain.SectionMeeting
	for rows.Next() {
		var m domain.SectionMeeting
		err := rows.Scan(
			&m.ID,
			&m.SectionID,
			&m.DayOfWeek,
			&m.StartTime,
			&m.EndTime,
			&m.Room,
			&m.Building,
		)

		if err != nil {
			return nil, err
		}

		meetings = append(meetings, m)
	}

	return meetings, nil
}
