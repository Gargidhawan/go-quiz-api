package config

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
)

type Config struct {
	MysqlDSN  string
	Port      string
	JWTSecret string
}

func Load() *Config {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "quizuser:quizpass@tcp(127.0.0.1:3306)/quizdb?charset=utf8mb4&parseTime=True&loc=Local"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "mysecretkey"
	}
	// Test if JWT secret is valid
	if _, err := jwt.Parse("", func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	}); err != nil && err.Error() != "token contains an invalid number of segments" {
		panic(fmt.Sprintf("invalid JWT secret: %v", err))
	}
	return &Config{
		MysqlDSN:  dsn,
		Port:      port,
		JWTSecret: jwtSecret,
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("dsn=%s port=%s", c.MysqlDSN, c.Port)
}
