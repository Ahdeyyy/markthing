package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"markthing/repository"
	"markthing/session"
	"net/http"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Status = string

const (
	SUCCESS = "SUCCESS"
	ERROR   = "ERROR"
)

type JsonResponse struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type HandlerParams struct {
	Database *pgx.Conn
}

func GetAllUsers(params HandlerParams) func(w http.ResponseWriter, r *http.Request) {
	db := params.Database
	query := repository.New(db)
	response := JsonResponse{}
	users, err := query.ListUsers(context.Background())

	return func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			response.Status = ERROR
			response.Message = fmt.Sprintf("failed to get all users: %s", err)
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}

}

func CreateUser(params HandlerParams) func(w http.ResponseWriter, r *http.Request) {
	db := params.Database
	query := repository.New(db)
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		response := JsonResponse{}
		form := r.Form
		username := form.Get("username")
		password := form.Get("password")
		if username == "" || password == "" {
			response.Status = ERROR
			response.Message = "failed to create user: username or password cannot be empty"

			http.Error(w, "failed to create user: username or password cannot be empty", http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return

		}
		encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			response.Status = ERROR
			response.Message = "failed to create user: internal server error"
			http.Error(w, "failed to create user: internal server error", http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		existingUser, _ := query.GetUserByUsername(context.Background(), username)
		if existingUser.Username == username {
			response.Status = ERROR
			response.Message = "failed to create user: a user with the username exists"
			log.Printf("error creating user: %s username: %s", err, username)
			http.Error(w, "failed to create user", http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		user, err := query.CreateUser(context.Background(), repository.CreateUserParams{
			Username: username,
			Password: string(encryptedPassword),
		})
		if err != nil {
			response.Status = ERROR
			response.Message = "failed to create user: something went wrong"
			log.Printf("error creating user: %s username: %s", err, username)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		response.Status = SUCCESS
		response.Message = fmt.Sprintf("successfully  created %s", user.Username)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	}
}

func Login(params HandlerParams) func(w http.ResponseWriter, r *http.Request) {
	db := params.Database
	query := repository.New(db)
	return func(w http.ResponseWriter, r *http.Request) {
		token := session.GenerateSessionToken()
		r.ParseForm()
		resp := JsonResponse{}
		form := r.Form
		username := form.Get("username")
		password := form.Get("password")

		if username == "" || password == "" {
			resp.Status = ERROR
			resp.Message = "failed to login  user: username or password cannot be empty"

			http.Error(w, "failed to login user: username or password cannot be empty", http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		user, err := query.GetUserByUsername(context.Background(), username)
		if err != nil {
			resp.Status = ERROR
			resp.Message = fmt.Sprintf("failed to login  user: %s", err)
			http.Error(w, "failed to login user: error quering row", http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		session := session.CreateSession(db, token, int(user.ID))
		cookie := http.Cookie{}
		cookie.Name = "session_token"
		cookie.Value = token
		cookie.Expires = session.ExpiresAt.Time
		cookie.Secure = false
		cookie.HttpOnly = true
		w.Header().Set("session_token", token)
		http.SetCookie(w, &cookie)

	}
}

type MiddlewareFunc = func(HandlerParams) http.HandlerFunc

func AuthMiddleware(params HandlerParams, next MiddlewareFunc) http.HandlerFunc {
	db := params.Database
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session_token")
		if sessionId == "" {
			log.Println("session id is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		session := session.ValidateSessionToken(db, sessionId)

		if session.Session.ID == "" || session.User.Username == "" {
			log.Printf("session: %v", session)
			w.WriteHeader(http.StatusUnauthorized)
			log.Println("failed to validate session token")
			return
		}
		next(params).ServeHTTP(w, r)

	}
}

func ProtectedRoute(params HandlerParams) http.HandlerFunc {
	db := params.Database
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session_token")
		session := session.ValidateSessionToken(db, sessionId)
		json.NewEncoder(w).Encode(session.User)
	}
}

// func (h Handler) ProtectRoute(w http.ResponseWriter, r *http.Request) {
//
// }
