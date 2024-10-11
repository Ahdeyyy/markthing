package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
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

type Handler struct {
	db *sql.DB
}

func newHandler(db *sql.DB) Handler {
	return Handler{
		db: db,
	}
}

func (h Handler) FindAllUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("select  id, username, password from  users")
	defer rows.Close()

	if err != nil {
		resp := JsonResponse{}
		resp.Status = ERROR
		resp.Message = "failed to create user: internal server error"
		http.Error(w, "failed to create user: internal server error", http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	users := make([]User, 0)
	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.Id, &user.Username, &user.Password)
		if err != nil {
			log.Printf("error scanning user row: %s", err)
		}
		users = append(users, user)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	resp := JsonResponse{}
	form := r.Form
	username := form.Get("username")
	password := form.Get("password")

	if username == "" || password == "" {
		resp.Status = ERROR
		resp.Message = "failed to create user: username or password cannot be empty"

		http.Error(w, "failed to create user: username or password cannot be empty", http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return

	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		resp.Status = ERROR
		resp.Message = "failed to create user: internal server error"
		http.Error(w, "failed to create user: internal server error", http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	_, err = h.db.Exec("insert into users (username,password) values ($1,$2)", username, encryptedPassword)
	if err != nil {
		resp.Status = ERROR
		resp.Message = "failed to create user: username is taken"
		log.Printf("error: %s username: %s", err, username)
		http.Error(w, "failed to create user", http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Status = SUCCESS
	resp.Message = fmt.Sprintf("successfully  created %s", username)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	db := h.db
	token := GenerateSessionToken()
	r.ParseForm()
	resp := JsonResponse{}
	form := r.Form
	username := form.Get("username")
	password := form.Get("password")
	user := User{}

	if username == "" || password == "" {
		resp.Status = ERROR
		resp.Message = "failed to login  user: username or password cannot be empty"

		http.Error(w, "failed to login user: username or password cannot be empty", http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	row := db.QueryRow("select id, username, password from users where username = $1", username)
	err := row.Err()
	if err != nil {

		resp.Status = ERROR
		resp.Message = "failed to login  user: error quering row"
		http.Error(w, "failed to login user: error quering row", http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return

	}

	row.Scan(&user.Id, &user.Username, &user.Password)
	session := CreateSession(h.db, token, user.Id)
	cookie := http.Cookie{}
	cookie.Name = "session_token"
	cookie.Value = token
	cookie.Expires = session.ExpiresAt
	cookie.Secure = false
	cookie.HttpOnly = true
	w.Header().Set("session_token", token)

	http.SetCookie(w, &cookie)
}

func (h Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionId := r.Header.Get("session_token")
		if sessionId == "" {
			log.Println("session id is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		session, user := ValidateSessionToken(h.db, sessionId)
		if session == nil || user == nil {
			log.Printf("session: %v, user: %v", session, user)
			w.WriteHeader(http.StatusUnauthorized)
			log.Println("failed to validate session token")
			return
		}
		next.ServeHTTP(w, r)

	}

}
func (h Handler) ProtectRoute(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("session_token")
	_, user := ValidateSessionToken(h.db, sessionId)
	json.NewEncoder(w).Encode(user)

}
