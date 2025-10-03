package quizzes

import (
	"errors"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"quizapi/internal/models"
)

type Service struct{ db *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// --- Quiz management ---

func (s *Service) CreateQuiz(title string) (*models.Quiz, error) {
	q := &models.Quiz{Title: title}
	return q, s.db.Create(q).Error
}

func (s *Service) ListQuizzes(page, limit int) ([]models.Quiz, int64, error) {
	var quizzes []models.Quiz
	var total int64

	// First, count the total number of records without pagination.
	// This is for the API response metadata.
	if err := s.db.Model(&models.Quiz{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate the offset for pagination.
	offset := (page - 1) * limit

	// Now, fetch the actual page of data.
	err := s.db.Offset(offset).Limit(limit).Order("id desc").Find(&quizzes).Error
	if err != nil {
		return nil, 0, err
	}

	return quizzes, total, nil
}

// AddQuestion validates per type, then writes Question + Options
func (s *Service) AddQuestion(quizID uint, req CreateQuestionReq) (*models.Question, error) {
	qt := models.QuestionType(req.Type)

	switch qt {
	case models.QText:
		if len(req.Options) > 0 {
			return nil, errors.New("text questions must not have options")
		}
		if req.WordLimit == nil || *req.WordLimit <= 0 || *req.WordLimit > 300 {
			return nil, errors.New("text questions require word_limit in 1..300")
		}
	case models.QSingle, models.QMultiple:
		if len(req.Options) < 2 {
			return nil, errors.New("choice questions need at least 2 options")
		}
		corr := 0
		for _, o := range req.Options {
			if o.IsCorrect != nil && *o.IsCorrect {
				corr++
			}
		}
		if qt == models.QSingle && corr != 1 {
			return nil, errors.New("single choice requires exactly 1 correct option")
		}
		if qt == models.QMultiple && corr < 1 {
			return nil, errors.New("multiple choice requires >=1 correct option")
		}
	default:
		return nil, errors.New("unknown question type")
	}

	q := &models.Question{
		QuizID:    quizID,
		Text:      req.Text,
		Type:      qt,
		WordLimit: req.WordLimit,
	}
	if err := s.db.Create(q).Error; err != nil {
		return nil, err
	}

	// insert options for choice questions
	for _, o := range req.Options {
		isCorr := o.IsCorrect != nil && *o.IsCorrect
		op := &models.Option{QuestionID: q.ID, Text: o.Text, IsCorrect: isCorr}
		if err := s.db.Create(op).Error; err != nil {
			return nil, err
		}
	}
	return q, nil
}

// GetPublicQuestions returns questions + options without leaking answers
func (s *Service) GetPublicQuestions(quizID uint) ([]PublicQuestion, error) {
	var qs []models.Question
	if err := s.db.Preload("Options").Where("quiz_id = ?", quizID).Find(&qs).Error; err != nil {
		return nil, err
	}
	out := make([]PublicQuestion, 0, len(qs))
	for _, q := range qs {
		pq := PublicQuestion{
			ID:        q.ID,
			Text:      q.Text,
			Type:      string(q.Type),
			WordLimit: q.WordLimit,
		}
		for _, op := range q.Options {
			pq.Options = append(pq.Options, PublicOption{ID: op.ID, Text: op.Text})
		}
		out = append(out, pq)
	}
	return out, nil
}

// --- Submission & scoring ---

// SubmitAndScore persists a submission + answers (transaction) and returns (score,total).
// Policy: auto-grade only single/multiple; text is stored but not counted in "total".
func (s *Service) SubmitAndScore(quizID uint, req SubmitReq) (*models.Submission, int, int, error) {
	// Load all quiz questions + their options once.
	var qs []models.Question
	if err := s.db.Preload("Options").Where("quiz_id = ?", quizID).Find(&qs).Error; err != nil {
		return nil, 0, 0, err
	}
	if len(qs) == 0 {
		return nil, 0, 0, fmt.Errorf("quiz %d not found or has no questions", quizID)
	}

	// Build lookup maps per question
	qByID := map[uint]models.Question{}
	for _, q := range qs {
		qByID[q.ID] = q
	}

	sub := &models.Submission{QuizID: quizID}
	score, total := 0, 0

	// Use a DB transaction to keep submission + answers atomic.
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sub).Error; err != nil {
			return err
		}

		for _, a := range req.Answers {
			q, ok := qByID[a.QuestionID]
			if !ok {
				return fmt.Errorf("question %d does not belong to quiz", a.QuestionID)
			}

			ans := &models.Answer{SubmissionID: sub.ID, QuestionID: q.ID}

			switch q.Type {
			case models.QSingle:
				if a.SelectedOptionID == nil {
					return errors.New("single choice requires selected_option_id")
				}
				// option must belong to the question
				if !containsOptionID(q.Options, *a.SelectedOptionID) {
					return fmt.Errorf("option %d invalid for question %d", *a.SelectedOptionID, q.ID)
				}
				if err := tx.Create(ans).Error; err != nil {
					return err
				}
				if err := tx.Create(&models.AnswerOption{
					AnswerID: ans.ID, OptionID: *a.SelectedOptionID,
				}).Error; err != nil {
					return err
				}
				total++
				if isCorrectSingle(q.Options, *a.SelectedOptionID) {
					score++
				}

			case models.QMultiple:
				if len(a.SelectedOptionIDs) == 0 {
					return errors.New("multiple choice requires selected_option_ids")
				}
				// validate all options belong to the question (and dedupe)
				dedup := dedupUint(a.SelectedOptionIDs)
				for _, oid := range dedup {
					if !containsOptionID(q.Options, oid) {
						return fmt.Errorf("option %d invalid for question %d", oid, q.ID)
					}
				}
				if err := tx.Create(ans).Error; err != nil {
					return err
				}
				for _, oid := range dedup {
					if err := tx.Create(&models.AnswerOption{AnswerID: ans.ID, OptionID: oid}).Error; err != nil {
						return err
					}
				}
				total++
				if exactSetMatch(correctIDs(q.Options), dedup) {
					score++
				}

			case models.QText:
				if q.WordLimit == nil {
					return errors.New("text question missing word_limit")
				}
				if a.TextAnswer == nil {
					return errors.New("text question requires text_answer")
				}
				if runeCount(*a.TextAnswer) > *q.WordLimit {
					return fmt.Errorf("text answer exceeds word_limit %d", *q.WordLimit)
				}
				ans.TextAnswer = a.TextAnswer
				if err := tx.Create(ans).Error; err != nil {
					return err
				}
				// not auto-graded; don't increment total
			}
		}
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}
	return sub, score, total, nil
}

// --- helpers ---
func exactSetMatch(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsOptionID(opts []models.Option, id uint) bool {
	for _, o := range opts {
		if o.ID == id {
			return true
		}
	}
	return false
}

func isCorrectSingle(opts []models.Option, id uint) bool {
	for _, o := range opts {
		if o.IsCorrect && o.ID == id {
			return true
		}
	}
	return false
}

func correctIDs(opts []models.Option) []uint {
	var ids []uint
	for _, o := range opts {
		if o.IsCorrect {
			ids = append(ids, o.ID)
		}
	}
	slices.Sort(ids)
	return ids
}

func dedupUint(in []uint) []uint {
	m := make(map[uint]struct{}, len(in))
	var out []uint
	for _, v := range in {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			out = append(out, v)
		}
	}
	slices.Sort(out)
	return out
}

func runeCount(s string) int {
	return len([]rune(s))
}
