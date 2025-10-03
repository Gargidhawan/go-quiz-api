package main

import (
	"log"

	"quizapi/internal/auth"
	"quizapi/internal/config"
	"quizapi/internal/db"
	"quizapi/internal/models"
	"quizapi/internal/quizzes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	d := db.Connect(cfg.MysqlDSN)

	quizsvc := quizzes.NewService(d)
	authSvc := auth.NewService(d, cfg.JWTSecret)

	quizH := quizzes.NewHandler(quizsvc)
	authH := auth.NewHandler(authSvc)

	r := gin.Default()

	// --Public routes--
	// Anyone can register/login, or see the list of available quizzes
	r.POST("/register", authH.Register)
	r.POST("/login", authH.Login)
	r.GET("/quizzes", quizH.ListQuizzes)

	// --Authenticated routes--
	// A user must have a valid token to access these, but any role is fine
	authRoutes := r.Group("/")
	authRoutes.Use(authSvc.AuthMiddleware())
	{
		authRoutes.GET("/quizzes/:quizID/questions", quizH.GetQuestions)
		authRoutes.POST("/quizzes/:quizID/submit", quizH.Submit)
	}
	// --Admin-Only routes--
	// A user must have a valid token and the "admin" role to access these
	adminRoutes := r.Group("/")
	adminRoutes.Use(authSvc.AuthMiddleware(), auth.RoleMiddleware(models.RoleAdmin))
	{
		adminRoutes.POST("/quizzes", quizH.CreateQuiz)
		adminRoutes.POST("/quizzes/:quizID/questions", quizH.AddQuestion)
	}
	log.Printf("listening on %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatal(err)
	}
}
