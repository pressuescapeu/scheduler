package postgres

import (
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

//go:embed school_schedule_by_term.csv
var csvData string

type CourseData struct {
	School      string
	Level       string
	CourseAbbr  string
	SectionType string
	CourseTitle string
	Credits     float64
	StartDate   string
	EndDate     string
	Days        string
	Time        string
	Enr         int
	Cap         int
	Faculty     string
	Room        string
}

// SeedDatabase imports course data from CSV if database is empty
func (s *Storage) SeedDatabase(ctx context.Context) error {
	// Check if we already have data
	var count int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM courses").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check courses: %w", err)
	}

	if count > 0 {
		log.Printf("Database already has %d courses, skipping seed", count)
		return nil
	}

	log.Println("Starting database seeding...")

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Skip first 3 rows (headers)
	if len(records) < 4 {
		return fmt.Errorf("CSV file too short")
	}

	semester := strings.TrimSpace(records[0][0]) // "Spring 2026"
	dataRows := records[3:]

	professorMap := make(map[string]int)
	courseMap := make(map[string]int)

	coursesAdded := 0
	sectionsAdded := 0

	for _, row := range dataRows {
		if len(row) < 14 || strings.TrimSpace(row[2]) == "" {
			continue
		}

		data := CourseData{
			School:      strings.TrimSpace(row[0]),
			Level:       strings.TrimSpace(row[1]),
			CourseAbbr:  strings.TrimSpace(row[2]),
			SectionType: strings.TrimSpace(row[3]),
			CourseTitle: strings.TrimSpace(row[4]),
			Credits:     parseCredits(row[6]), // Use ECTS credits (column 6), not US credits
			StartDate:   strings.TrimSpace(row[7]),
			EndDate:     strings.TrimSpace(row[8]),
			Days:        strings.TrimSpace(row[9]),
			Time:        strings.TrimSpace(row[10]),
			Enr:         parseInt(row[11]),
			Cap:         parseInt(row[12]),
			Faculty:     strings.TrimSpace(row[13]),
			Room:        strings.TrimSpace(row[14]),
		}

		// Extract course code and section number
		courseCode, sectionNum := parseCourseAbbr(data.CourseAbbr, data.SectionType)
		if courseCode == "" {
			continue
		}

		// Insert or get course
		courseID, exists := courseMap[courseCode]
		if !exists {
			courseID, err = s.insertCourse(ctx, courseCode, data.CourseTitle, int(data.Credits), semester)
			if err != nil {
				log.Printf("Failed to insert course %s: %v", courseCode, err)
				continue
			}
			courseMap[courseCode] = courseID
			coursesAdded++
		}

		// Insert professors
		professorIDs := s.insertProfessors(ctx, data.Faculty, professorMap)
		var professorID *int
		if len(professorIDs) > 0 {
			professorID = &professorIDs[0]
		}

		// Determine section type
		sectionType := determineSectionType(data.SectionType)

		// Insert section
		sectionID, err := s.insertSection(ctx, courseID, sectionNum, sectionType, professorID, data.Cap)
		if err != nil {
			log.Printf("Failed to insert section %s-%s: %v", courseCode, sectionNum, err)
			continue
		}
		sectionsAdded++

		// Insert section meetings
		if data.Days != "" && data.Time != "" {
			s.insertSectionMeetings(ctx, sectionID, data.Days, data.Time, data.Room)
		}
	}

	log.Printf("Seeding complete: %d courses, %d sections, %d professors",
		coursesAdded, sectionsAdded, len(professorMap))

	return nil
}

func (s *Storage) insertCourse(ctx context.Context, code, name string, credits int, semester string) (int, error) {
	var id int
	query := `INSERT INTO courses (course_code, course_name, credits, semester, is_internship)
	          VALUES ($1, $2, $3, $4, false)
	          ON CONFLICT (course_code) DO UPDATE SET course_name = EXCLUDED.course_name
	          RETURNING id`
	err := s.pool.QueryRow(ctx, query, code, name, credits, semester).Scan(&id)
	return id, err
}

func (s *Storage) insertProfessors(ctx context.Context, faculty string, profMap map[string]int) []int {
	if faculty == "" || faculty == "Online/Distant" {
		return nil
	}

	// Split multiple professors and only take the first one
	names := strings.Split(faculty, ",")
	if len(names) == 0 {
		return nil
	}

	name := strings.TrimSpace(names[0]) // Only use first professor
	if name == "" {
		return nil
	}

	if id, exists := profMap[name]; exists {
		return []int{id}
	}

	// Parse first and last name
	parts := strings.Fields(name)
	if len(parts) < 2 {
		return nil
	}

	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")

	var id int
	query := `INSERT INTO professors (first_name, last_name, email)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name
	          RETURNING id`
	email := fmt.Sprintf("%s.%s@nu.edu.kz",
		strings.ToLower(firstName),
		strings.ToLower(strings.ReplaceAll(lastName, " ", "")))

	err := s.pool.QueryRow(ctx, query, firstName, lastName, email).Scan(&id)
	if err != nil {
		return nil
	}

	profMap[name] = id
	return []int{id}
}

func (s *Storage) insertSection(ctx context.Context, courseID int, sectionNum, sectionType string, professorID *int, totalSeats int) (int, error) {
	var id int
	query := `INSERT INTO sections (course_id, section_number, section_type, professor_id, total_seats, available_seats)
	          VALUES ($1, $2, $3, $4, $5, $5)
	          ON CONFLICT (course_id, section_number) DO UPDATE SET total_seats = EXCLUDED.total_seats
	          RETURNING id`
	err := s.pool.QueryRow(ctx, query, courseID, sectionNum, sectionType, professorID, totalSeats).Scan(&id)
	return id, err
}

func (s *Storage) insertSectionMeetings(ctx context.Context, sectionID int, days, timeStr, room string) {
	daysList := parseDays(days)
	startTime, endTime := parseTime(timeStr)

	if startTime == "" || endTime == "" {
		return
	}

	building, roomNum := parseRoom(room)

	for _, day := range daysList {
		query := `INSERT INTO section_meetings (section_id, day_of_week, start_time, end_time, room, building)
		          VALUES ($1, $2, $3, $4, $5, $6)
		          ON CONFLICT DO NOTHING`
		_, err := s.pool.Exec(ctx, query, sectionID, day, startTime, endTime, roomNum, building)
		if err != nil {
			log.Printf("Failed to insert meeting for section %d: %v", sectionID, err)
		}
	}
}

// Helper functions

func parseCourseAbbr(abbr, sectionType string) (courseCode, sectionNum string) {
	// Remove slashes (cross-listed courses)
	parts := strings.Fields(strings.ReplaceAll(abbr, "/", " "))
	if len(parts) < 2 {
		return "", ""
	}

	// Last part is course code with section
	lastPart := parts[len(parts)-1]

	// Extract section number from type (e.g., "1L", "2S")
	sectionNum = sectionType

	// Course code is everything else
	courseCode = strings.Join(parts[:len(parts)-1], " ") + " " + strings.TrimRight(lastPart, "0123456789")

	return strings.TrimSpace(courseCode), sectionNum
}

func determineSectionType(typeCode string) string {
	if len(typeCode) == 0 {
		return "Lecture"
	}

	suffix := strings.ToUpper(string(typeCode[len(typeCode)-1]))
	switch suffix {
	case "L":
		return "Lecture"
	case "S":
		return "Seminar"
	case "B":
		return "Lab"
	case "R":
		return "Recitation"
	default:
		return "Lecture"
	}
}

func parseDays(days string) []string {
	days = strings.TrimSpace(days)
	if days == "" {
		return nil
	}

	dayMap := map[string]string{
		"M": "Monday",
		"T": "Tuesday",
		"W": "Wednesday",
		"R": "Thursday",
		"F": "Friday",
		"S": "Saturday",
	}

	var result []string
	for _, ch := range days {
		if day, ok := dayMap[string(ch)]; ok {
			result = append(result, day)
		}
	}
	return result
}

func parseTime(timeStr string) (start, end string) {
	if timeStr == "" || timeStr == "Online/Distant" {
		return "", ""
	}

	parts := strings.Split(timeStr, "-")
	if len(parts) != 2 {
		return "", ""
	}

	start = convertTo24Hour(strings.TrimSpace(parts[0]))
	end = convertTo24Hour(strings.TrimSpace(parts[1]))
	return
}

func convertTo24Hour(t string) string {
	t = strings.TrimSpace(t)
	if t == "" {
		return ""
	}

	// Parse time like "02:00 PM" to "14:00:00"
	parsed, err := time.Parse("03:04 PM", t)
	if err != nil {
		return ""
	}

	return parsed.Format("15:04:05")
}

func parseRoom(room string) (building, roomNum *string) {
	if room == "" {
		return nil, nil
	}

	// Format: "(C3) 1009 - cap:70" or "Green Hall - cap:231"
	parts := strings.Split(room, "-")
	if len(parts) == 0 {
		return nil, nil
	}

	location := strings.TrimSpace(parts[0])

	// Check if it has building code in parentheses
	if strings.Contains(location, "(") && strings.Contains(location, ")") {
		start := strings.Index(location, "(")
		end := strings.Index(location, ")")
		buildingCode := location[start+1 : end]
		room := strings.TrimSpace(location[end+1:])

		return &buildingCode, &room
	}

	// Otherwise, it's just a building name
	return &location, nil
}

func parseCredits(s string) float64 {
	s = strings.TrimSpace(s)
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	val, _ := strconv.Atoi(s)
	return val
}
