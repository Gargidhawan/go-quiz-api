package quizzes_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"quizapi/internal/models"
	"quizapi/internal/quizzes"
)

func memDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Quiz{}, &models.Question{}, &models.Option{},
		&models.Submission{}, &models.Answer{}, &models.AnswerOption{},
	))
	return db
}

func TestMultipleChoice_ExactMatchScores1(t *testing.T) {
	d := memDB(t)
	svc := quizzes.NewService(d)

	qz, err := svc.CreateQuiz("test")
	require.NoError(t, err)

	_, err = svc.AddQuestion(qz.ID, quizzes.CreateQuestionReq{
		Text: "Pick go tools", Type: "multiple",
		Options: []quizzes.CreateQuestionOption{
			{Text: "go test", IsCorrect: ptr(true)},
			{Text: "go vet", IsCorrect: ptr(true)},
			{Text: "pip", IsCorrect: ptr(false)},
		},
	})
	require.NoError(t, err)

	pub, err := svc.GetPublicQuestions(qz.ID)
	require.NoError(t, err)
	require.Len(t, pub, 1)

	q := pub[0]
	// Find IDs of the two correct options by text
	var ids []uint
	for _, o := range q.Options {
		if o.Text == "go test" || o.Text == "go vet" {
			ids = append(ids, o.ID)
		}
	}
	req := quizzes.SubmitReq{
		Answers: []quizzes.SubmitAnswer{
			{QuestionID: q.ID, SelectedOptionIDs: ids},
		},
	}

	_, score, total, err := svc.SubmitAndScore(qz.ID, req)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Equal(t, 1, score)
}

func ptr[T any](v T) *T { return &v }
