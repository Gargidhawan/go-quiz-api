# Online Quiz Application API (Go + Gin + GORM + MySQL)

A robust backend API for a quiz application built with Go, Gin, and GORM. This project features JWT authentication, role-based access control (Admin vs. Public), and a full suite of quiz management and participation endpoints.

#### Demo Link
https://drive.google.com/file/d/1N14VfH42IzAkeq_xpXNsEEw9_CjyFqYt/view?usp=sharing

## ✅ Features

* **JWT Authentication**: Secure user registration and login using JSON Web Tokens.
* **Role-Based Access Control**: Differentiates between `admin` users (who can create quizzes) and `public` users (who can take them).
* **Full Quiz Management**: Endpoints for creating quizzes and adding questions of different types (`single`, `multiple`, `text`).
* **Question Validation**: Enforces different rules for each question type (e.g., single-choice must have one correct answer).
* **Quiz Taking & Scoring**: Endpoints to fetch questions for a quiz (without revealing answers) and submit answers for automated scoring.
* **Paginated Lists**: The endpoint to list all available quizzes is paginated for efficiency.
* **Database Persistence**: Uses GORM with a MySQL database, managed via Docker for easy setup.
* **Unit Tested**: Core business logic, like scoring, is covered by unit tests using an in-memory SQLite database.

## 🛠️ Tech Stack

* **Backend**: Go (Golang)
* **Framework**: Gin (HTTP Web Framework)
* **Database**: MySQL (managed with Docker Compose)
* **ORM**: GORM (Go Object-Relational Mapper)
* **Key Libraries**:
    * `golang-jwt/jwt/v5` for JWT handling
    * `golang.org/x/crypto/bcrypt` for password hashing
    * `go-playground/validator/v10` for request validation
    * `stretchr/testify` for assertions in unit tests

## 🚀 Getting Started

Follow these instructions to get the project up and running on your local machine.

### Prerequisites

* [Go](https://go.dev/doc/install) (version 1.22+ recommended)
* [Docker](https://www.docker.com/products/docker-desktop/) (for running the MySQL database)
* `curl` or an API client like [Postman](https://www.postman.com/) for testing endpoints.

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/Gargidhawan/go-quiz-api.git](https://github.com/Gargidhawan/go-quiz-api.git)
    cd go-quiz-api
    ```

2.  **Configure Environment Variables:**
    The project reads its configuration from environment variables. A template is provided.
    ```bash
    # Copy the example file to create your own configuration
    cp .env.example .env
    ```
    *(The default values in this file match the Docker setup, so no changes are needed to get started.)*

3.  **Install Dependencies:**
    `go mod tidy` will download all the necessary libraries defined in `go.mod`.
    ```bash
    go mod tidy
    ```

### Running the Application

You will need **two separate terminal windows** open in the project directory.

1.  **Terminal 1: Start the Database**
    This command starts the MySQL database and the Adminer web interface in the background.
    ```bash
    docker compose up -d
    ```
    * The MySQL database will be available on port `3306`.
    * You can visually inspect the database by visiting `http://localhost:8081` (Adminer).

2.  **Terminal 2: Start the Go API Server**
    This command compiles and runs the main application.
    ```bash
    go run ./cmd/server
    ```
    * The server will start and listen on `http://localhost:8080`.

## 📖 API Endpoints

All endpoints are prefixed with `/`.

### Authentication

| Method | Endpoint | Description | Access | Example Body |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/register` | Creates a new user. The first user is an `admin`. | Public | `{"username":"user","password":"password123"}` |
| `POST` | `/login` | Logs in a user and returns a JWT token. | Public | `{"username":"user","password":"password123"}` |

### Quiz Management (Admin Only)

**Note:** All these endpoints require a valid JWT token in the `Authorization: Bearer <token>` header.

| Method | Endpoint | Description | Access | Example Body |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/quizzes` | Creates a new quiz. | Admin | `{"title":"New Go Quiz"}` |
| `POST` | `/quizzes/:quizID/questions` | Adds a new question to a specific quiz. | Admin | `{"text":"...", "type":"single", "options":[...]}` |

### Quiz Taking

| Method | Endpoint | Description | Access |
| :--- | :--- | :--- | :--- |
| `GET` | `/quizzes` | Lists all available quizzes. Supports pagination via query params `?page=1&limit=10`. | Public |
| `GET` | `/quizzes/:quizID/questions` | Fetches all questions for a quiz (without correct answers). | Authenticated |
| `POST` | `/quizzes/:quizID/submit` | Submits answers for a quiz and returns the score. | Authenticated |

## 🧪 Running Tests

To run the unit tests for the project, use the following command from the root directory:
```bash
go test ./...
