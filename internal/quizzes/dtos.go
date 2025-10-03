package quizzes

import "quizapi/internal/models"

type CreateQuizReq struct {
	Title string `json:"title" validate:"required,min=1,max=200"`
}

type CreateQuestionOption struct {
	Text      string `json:"text" validate:"required,min=1,max=300"`
	IsCorrect *bool  `json:"is_correct"`
}

type CreateQuestionReq struct {
	Text      string                 `json:"text" validate:"required,min=1"`
	Type      string                 `json:"type" validate:"required,oneof=single multiple text"`
	WordLimit *int                   `json:"word_limit"`
	Options   []CreateQuestionOption `json:"options"`
}
type ListQuizzesResp struct {
	Quizzes      []models.Quiz `json:"quizzes"`
	TotalRecords int64         `json:"total_records"`
	Page         int           `json:"page"`
	Limit        int           `json:"limit"`
}

type PublicOption struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
}

type PublicQuestion struct {
	ID        uint           `json:"id"`
	Text      string         `json:"text"`
	Type      string         `json:"type"`
	WordLimit *int           `json:"word_limit"`
	Options   []PublicOption `json:"options"`
}

type SubmitAnswer struct {
	QuestionID        uint    `json:"question_id" validate:"required"`
	SelectedOptionID  *uint   `json:"selected_option_id"`
	SelectedOptionIDs []uint  `json:"selected_option_ids"`
	TextAnswer        *string `json:"text_answer"`
}

type SubmitReq struct {
	Answers []SubmitAnswer `json:"answers" validate:"required,dive"`
}

type ScoreResp struct {
	Score int `json:"score"`
	Total int `json:"total"`
}
