package quizzes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	svc *Service
	val *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, val: validator.New()}
}

func (h *Handler) CreateQuiz(c *gin.Context) {
	var req CreateQuizReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.val.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	q, err := h.svc.CreateQuiz(req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, q)
}

func (h *Handler) ListQuizzes(c *gin.Context) {
	// --- Parse Pagination Parameters ---
	// Default to page 1
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	// Default to a limit of 10, with a max of 100
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 { // Enforce a max limit
		limit = 100
	}

	// --- Call the Service ---
	quizzes, total, err := h.svc.ListQuizzes(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// --- Construct and Send the Response ---
	c.JSON(http.StatusOK, ListQuizzesResp{
		Quizzes:      quizzes,
		TotalRecords: total,
		Page:         page,
		Limit:        limit,
	})
}

func (h *Handler) AddQuestion(c *gin.Context) {
	quizID, err := strconv.Atoi(c.Param("quizID"))
	if err != nil || quizID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quizID"})
		return
	}
	var req CreateQuestionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.val.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	q, err := h.svc.AddQuestion(uint(quizID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": q.ID})
}

func (h *Handler) GetQuestions(c *gin.Context) {
	quizID, err := strconv.Atoi(c.Param("quizID"))
	if err != nil || quizID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quizID"})
		return
	}
	qs, err := h.svc.GetPublicQuestions(uint(quizID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, qs)
}

func (h *Handler) Submit(c *gin.Context) {
	quizID, err := strconv.Atoi(c.Param("quizID"))
	if err != nil || quizID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quizID"})
		return
	}
	var req SubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.val.Struct(req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	_, score, total, serr := h.svc.SubmitAndScore(uint(quizID), req)
	if serr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": serr.Error()})
		return
	}
	c.JSON(http.StatusOK, ScoreResp{Score: score, Total: total})
}
