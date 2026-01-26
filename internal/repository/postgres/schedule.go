package postgres

import (
	"context"
	"scheduler/internal/domain"
)

func (s *Storage) CreateSchedule(ctx context.Context, studentID int, req *domain.CreateScheduleRequest) (*domain.Schedule, error) {
	const query = `
		INSERT INTO schedules (student_id, schedule_name, description)
        VALUES ($1, $2, $3)
        RETURNING id, student_id, schedule_name, description, is_submitted, created_at;
	`

	var schedule domain.Schedule
	err := s.pool.QueryRow(ctx, query, studentID, req.ScheduleName, req.Description).Scan(
		&schedule.ID,
		&schedule.StudentID,
		&schedule.ScheduleName,
		&schedule.Description,
		&schedule.IsSubmitted,
		&schedule.CreatedAt,
	)

	return &schedule, err
}

func (s *Storage) GetStudentSchedules(ctx context.Context, studentID int) ([]domain.Schedule, error) {
	const query = `
		SELECT id, student_id, schedule_name, description, is_submitted, created_at
        FROM schedules
        WHERE student_id = $1
        ORDER BY created_at DESC;
	`

	rows, err := s.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var schedules []domain.Schedule
	for rows.Next() {
		var sch domain.Schedule
		err := rows.Scan(
			&sch.ID,
			&sch.StudentID,
			&sch.ScheduleName,
			&sch.Description,
			&sch.IsSubmitted,
			&sch.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, sch)
	}

	return schedules, nil
}

func (s *Storage) AddSectionToSchedule(ctx context.Context, scheduleID, sectionID int) error {
	const query = `
        INSERT INTO schedule_sections (schedule_id, section_id)
        VALUES ($1, $2);
    `
	_, err := s.pool.Exec(ctx, query, scheduleID, sectionID)
	return err
}

func (s *Storage) RemoveSectionFromSchedule(ctx context.Context, scheduleID, sectionID int) error {
	const query = `DELETE FROM schedule_sections WHERE schedule_id = $1 AND section_id = $2;`
	_, err := s.pool.Exec(ctx, query, scheduleID, sectionID)
	return err
}

func (s *Storage) GetScheduleWithSections(ctx context.Context, scheduleID int) (*domain.ScheduleWithSections, error) {
	var schedule domain.Schedule

	const query = `SELECT id, student_id, schedule_name, description, is_submitted, created_at FROM schedules WHERE id = $1;`

	const query2 = `
		SELECT s.id, s.course_id, s.section_number, s.section_type,
               s.professor_id, s.total_seats, s.available_seats, s.parent_section_id,
               c.course_code, c.course_name, c.credits
        FROM schedule_sections ss
        JOIN sections s ON ss.section_id = s.id
        JOIN courses c ON s.course_id = c.id
        WHERE ss.schedule_id = $1;`

	err := s.pool.QueryRow(ctx, query, scheduleID).Scan(
		&schedule.ID,
		&schedule.StudentID,
		&schedule.ScheduleName,
		&schedule.Description,
		&schedule.IsSubmitted,
		&schedule.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, query2, scheduleID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var sections []domain.SectionWithDetails
	totalCredits := 0

	for rows.Next() {
		var sd domain.SectionWithDetails
		err := rows.Scan(
			&sd.ID,
			&sd.CourseID,
			&sd.SectionNumber,
			&sd.SectionType,
			&sd.ProfessorID,
			&sd.TotalSeats,
			&sd.AvailableSeats,
			&sd.ParentSectionID,
			&sd.Course.CourseCode,
			&sd.Course.CourseName,
			&sd.Course.Credits,
		)
		if err != nil {
			return nil, err
		}

		meetings, _ := s.GetSectionMeetings(ctx, sd.ID)
		sd.Meetings = meetings

		sections = append(sections, sd)
		totalCredits += sd.Course.Credits
	}

	return &domain.ScheduleWithSections{
		Schedule:     schedule,
		Sections:     sections,
		TotalCredits: totalCredits,
	}, nil
}
