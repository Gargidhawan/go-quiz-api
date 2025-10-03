package models

import "time"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"type:varchar(100);not null" json:"-"`
	Role         Role      `gorm:"type:varchar(16);not null;default:'user'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Quiz struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Title     string     `gorm:"type:varchar(200);not null" json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	Questions []Question `json:"-"`
}

type QuestionType string

const (
	QSingle   QuestionType = "single"
	QMultiple QuestionType = "multiple"
	QText     QuestionType = "text"
)

type Question struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	QuizID    uint         `gorm:"index;not null" json:"quiz_id"`
	Text      string       `gorm:"type:text;not null" json:"text"`
	Type      QuestionType `gorm:"type:varchar(16);not null" json:"type"`
	WordLimit *int         `json:"word_limit"`
	Options   []Option     `gorm:"constraint:OnDelete:CASCADE" json:"options"`
}

type Option struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	QuestionID uint   `gorm:"index;not null" json:"question_id"`
	Text       string `gorm:"type:varchar(300);not null" json:"text"`
	IsCorrect  bool   `gorm:"not null" json:"-"` // never expose in public JSON
}

type Submission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	QuizID    uint      `gorm:"index;not null" json:"quiz_id"`
	CreatedAt time.Time `json:"created_at"`
	Answers   []Answer  `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

type Answer struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	SubmissionID uint           `gorm:"index;not null" json:"submission_id"`
	QuestionID   uint           `gorm:"index;not null" json:"question_id"`
	TextAnswer   *string        `json:"text_answer"`
	Options      []AnswerOption `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

type AnswerOption struct {
	AnswerID uint `gorm:"primaryKey"`
	OptionID uint `gorm:"primaryKey"`
}
