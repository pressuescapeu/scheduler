CREATE TABLE IF NOT EXISTS students (
                          id SERIAL PRIMARY KEY,
                          email VARCHAR(255) UNIQUE NOT NULL,
                          password_hash VARCHAR(255) NOT NULL,
                          first_name VARCHAR(100) NOT NULL,
                          last_name VARCHAR(100) NOT NULL,
                          student_id VARCHAR(20) UNIQUE NOT NULL,
                          year_of_study INTEGER CHECK (year_of_study >= 1 AND year_of_study <= 5),
                          total_credits_earned INTEGER DEFAULT 0,
                          created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS professors (
                            id SERIAL PRIMARY KEY,
                            first_name VARCHAR(100) NOT NULL,
                            last_name VARCHAR(100) NOT NULL,
                            email VARCHAR(255) UNIQUE,
                            rating DECIMAL(2,1) CHECK (rating >= 0 AND rating <= 5),
                            created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courses (
                         id SERIAL PRIMARY KEY,
                         course_code VARCHAR(20) UNIQUE NOT NULL, -- "CSCI332"
                         course_name VARCHAR(200) NOT NULL,        -- "Operating Systems"
                         credits INTEGER NOT NULL CHECK (credits > 0 AND credits <= 12),
                         is_internship BOOLEAN DEFAULT FALSE,
                         description TEXT,
                         semester VARCHAR(20) NOT NULL,            -- "Summer 2025"
                         created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sections (
                          id SERIAL PRIMARY KEY,
                          course_id INTEGER NOT NULL,
                          section_number VARCHAR(10) NOT NULL,      -- "001", "002", "L01" (for labs)
                          section_type VARCHAR(20) NOT NULL         -- "Lecture", "Lab", "Recitation", "Seminar"
                              CHECK (section_type IN ('Lecture', 'Lab', 'Recitation', 'Seminar')),
                          professor_id INTEGER,                     -- Can be NULL if TBA
                          total_seats INTEGER DEFAULT 30 CHECK (total_seats > 0),
                          available_seats INTEGER DEFAULT 30,       -- Just for display
                          parent_section_id INTEGER,                -- For linking Lab to Lecture

                          FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
                          FOREIGN KEY (professor_id) REFERENCES professors(id) ON DELETE SET NULL,
                          FOREIGN KEY (parent_section_id) REFERENCES sections(id) ON DELETE CASCADE,
                          UNIQUE(course_id, section_number)
);

CREATE TABLE IF NOT EXISTS section_meetings (
                                  id SERIAL PRIMARY KEY,
                                  section_id INTEGER NOT NULL,
                                  day_of_week VARCHAR(10) NOT NULL
                                      CHECK (day_of_week IN ('Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday')),
                                  start_time TIME NOT NULL,
                                  end_time TIME NOT NULL CHECK (end_time > start_time),
                                  room VARCHAR(50),
                                  building VARCHAR(50),

                                  FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schedules (
                           id SERIAL PRIMARY KEY,
                           student_id INTEGER NOT NULL,
                           schedule_name VARCHAR(100) NOT NULL,
                           description TEXT,
                           is_submitted BOOLEAN DEFAULT FALSE,
                           created_at TIMESTAMP DEFAULT NOW(),

                           FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS schedule_sections (
                                   id SERIAL PRIMARY KEY,
                                   schedule_id INTEGER NOT NULL,
                                   section_id INTEGER NOT NULL,
                                   added_at TIMESTAMP DEFAULT NOW(),

                                   FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
                                   FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE,
                                   UNIQUE(schedule_id, section_id)
);

CREATE INDEX IF NOT EXISTS idx_sections_course ON sections(course_id);
CREATE INDEX IF NOT EXISTS idx_sections_professor ON sections(professor_id);
CREATE INDEX IF NOT EXISTS idx_section_meetings_section ON section_meetings(section_id);
CREATE INDEX IF NOT EXISTS idx_schedules_student ON schedules(student_id);
CREATE INDEX IF NOT EXISTS idx_schedule_sections_schedule ON schedule_sections(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_sections_section ON schedule_sections(section_id);