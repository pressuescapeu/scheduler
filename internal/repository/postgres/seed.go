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

type MeetingInfo struct {
	Day      string
	Start    string
	End      string
	Building *string
	Room     *string
}

type SectionInfo struct {
	CourseCode   string
	SectionNum   string
	SectionType  string
	CourseTitle  string
	Credits      float64
	Semester     string
	ProfessorIDs []int
	TotalSeats   int
	Meetings     map[string]MeetingInfo
}

func (s *Storage) SeedDatabase(ctx context.Context) error {
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

	if len(records) < 4 {
		return fmt.Errorf("CSV file too short")
	}

	semester := strings.TrimSpace(records[0][0])
	dataRows := records[3:]

	professorMap := make(map[string]int)
	courseMap := make(map[string]int)
	sectionMap := make(map[string]*SectionInfo)
	courseCredits := make(map[string]float64)
	courseCreditSource := make(map[string]string)

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
			Credits:     parseCredits(row[6]),
			StartDate:   strings.TrimSpace(row[7]),
			EndDate:     strings.TrimSpace(row[8]),
			Days:        strings.TrimSpace(row[9]),
			Time:        strings.TrimSpace(row[10]),
			Enr:         parseInt(row[11]),
			Cap:         parseInt(row[12]),
			Faculty:     strings.TrimSpace(row[13]),
			Room:        strings.TrimSpace(row[14]),
		}

		courseCode, sectionNum := parseCourseAbbr(data.CourseAbbr, data.SectionType)
		if courseCode == "" {
			continue
		}

		sectionType := determineSectionType(data.SectionType)
		sectionKey := courseCode + "|" + sectionNum

		if sectionType == "Lecture" {
			if courseCreditSource[courseCode] != "Lecture" && data.Credits > 0 {
				courseCredits[courseCode] = data.Credits
				courseCreditSource[courseCode] = "Lecture"
			}
		} else {
			if _, exists := courseCredits[courseCode]; !exists && data.Credits > 0 {
				courseCredits[courseCode] = data.Credits
				courseCreditSource[courseCode] = "Other"
			}
		}

		if _, exists := sectionMap[sectionKey]; !exists {
			professorIDs := s.insertProfessors(ctx, data.Faculty, professorMap)
			sectionMap[sectionKey] = &SectionInfo{
				CourseCode:   courseCode,
				SectionNum:   sectionNum,
				SectionType:  sectionType,
				CourseTitle:  data.CourseTitle,
				Credits:      data.Credits,
				Semester:     semester,
				ProfessorIDs: professorIDs,
				TotalSeats:   data.Cap,
				Meetings:     make(map[string]MeetingInfo),
			}
		}

		section := sectionMap[sectionKey]

		if data.Days != "" && data.Time != "" {
			daysList := parseDays(data.Days)
			startTime, endTime := parseTime(data.Time)
			building, roomNum := parseRoom(data.Room)

			for _, day := range daysList {
				meetingKey := day + "|" + startTime + "|" + endTime
				if _, exists := section.Meetings[meetingKey]; !exists {
					section.Meetings[meetingKey] = MeetingInfo{
						Day:      day,
						Start:    startTime,
						End:      endTime,
						Building: building,
						Room:     roomNum,
					}
				}
			}
		}
	}

	coursesAdded := 0
	sectionsAdded := 0

	for _, section := range sectionMap {
		courseID, exists := courseMap[section.CourseCode]
		if !exists {
			credits := section.Credits
			if preferredCredits, ok := courseCredits[section.CourseCode]; ok {
				credits = preferredCredits
			}
			courseID, err = s.insertCourse(ctx, section.CourseCode, section.CourseTitle, int(credits), section.Semester)
			if err != nil {
				log.Printf("Failed to insert course %s: %v", section.CourseCode, err)
				continue
			}
			courseMap[section.CourseCode] = courseID
			coursesAdded++
		}

		var professorID *int
		if len(section.ProfessorIDs) > 0 {
			professorID = &section.ProfessorIDs[0]
		}

		sectionID, err := s.insertSection(ctx, courseID, section.SectionNum, section.SectionType, professorID, section.TotalSeats)
		if err != nil {
			log.Printf("Failed to insert section %s-%s: %v", section.CourseCode, section.SectionNum, err)
			continue
		}
		sectionsAdded++

		for _, meeting := range section.Meetings {
			s.insertSectionMeeting(ctx, sectionID, meeting)
		}
	}

	log.Printf("Seeding complete: %d courses, %d sections, %d professors",
		coursesAdded, sectionsAdded, len(professorMap))

	return nil
}

func (s *Storage) insertCourse(ctx context.Context, code, name string, credits int, semester string) (int, error) {
	var id int
	isInternship := strings.Contains(strings.ToLower(name), "internship")
	query := `INSERT INTO courses (course_code, course_name, credits, semester, is_internship)
	          VALUES ($1, $2, $3, $4, $5)
	          ON CONFLICT (course_code) DO UPDATE SET course_name = EXCLUDED.course_name, is_internship = EXCLUDED.is_internship
	          RETURNING id`
	err := s.pool.QueryRow(ctx, query, code, name, credits, semester, isInternship).Scan(&id)
	return id, err
}

func (s *Storage) insertProfessors(ctx context.Context, faculty string, profMap map[string]int) []int {
	if faculty == "" || faculty == "Online/Distant" {
		return nil
	}

	names := strings.Split(faculty, ",")
	if len(names) == 0 {
		return nil
	}

	name := strings.TrimSpace(names[0])
	if name == "" {
		return nil
	}

	if id, exists := profMap[name]; exists {
		return []int{id}
	}

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

func (s *Storage) insertSectionMeeting(ctx context.Context, sectionID int, meeting MeetingInfo) {
	query := `INSERT INTO section_meetings (section_id, day_of_week, start_time, end_time, room, building)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          ON CONFLICT DO NOTHING`
	_, err := s.pool.Exec(ctx, query, sectionID, meeting.Day, meeting.Start, meeting.End, meeting.Room, meeting.Building)
	if err != nil {
		log.Printf("Failed to insert meeting for section %d: %v", sectionID, err)
	}
}

func parseCourseAbbr(abbr, sectionType string) (courseCode, sectionNum string) {
	abbr = strings.TrimSpace(abbr)

	if strings.Contains(abbr, "/") {
		parts := strings.Split(abbr, "/")
		abbr = strings.TrimSpace(parts[0])
	}

	return abbr, sectionType
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

	parts := strings.Split(room, "-")
	if len(parts) == 0 {
		return nil, nil
	}

	location := strings.TrimSpace(parts[0])

	if strings.Contains(location, "(") && strings.Contains(location, ")") {
		start := strings.Index(location, "(")
		end := strings.Index(location, ")")
		buildingCode := location[start+1 : end]
		room := strings.TrimSpace(location[end+1:])

		return &buildingCode, &room
	}

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
