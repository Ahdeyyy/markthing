package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"log"
	"strings"
	"time"
)

type Role = string

const (
	ADMIN = "admin"
	USER  = "user"
	GUEST = "guest"
)

type SessionHandler struct {
	db *sql.DB
}

type Session struct {
	Id        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	// CreatedAt  time.Time `json:"created_at"`
	// Role       Role      `json:"role"`
}

func encodeHex(token string) string {

	encodedToken := []byte(token)
	hashedToken := sha256.New()
	hashedToken.Write(encodedToken)
	hashedBytes := hashedToken.Sum(nil)

	return strings.ToLower(base32.HexEncoding.EncodeToString(hashedBytes))
}

func GenerateSessionToken() string {
	b := make([]byte, 20)
	rand.Reader.Read(b)
	token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return token
}

func CreateSession(db *sql.DB, token string, userId int) Session {
	sessionId := encodeHex(token)
	sess := Session{
		Id:        sessionId,
		UserID:    userId,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}
	_, err := db.Exec(`
			insert into sessions (id,user_id, expires_at) 
			values ($1,$2, $3)`,
		sess.Id, sess.UserID, sess.ExpiresAt)
	if err != nil {
		log.Printf("error: could not create new session %s", err)
		return Session{Id: "error"}
	}
	return sess
}
func ValidateSessionToken(db *sql.DB, token string) (*Session, *User) {
	sessionId := encodeHex(token)
	session := Session{}
	user := User{}

	row := db.QueryRow(`select sessions.id, sessions.user_id, sessions.expires_at, users.id as user_id
		from sessions inner join users on users.id = sessions.user_id where sessions.id = $1`, sessionId)

	err := row.Err()
	if err != nil {
		log.Printf("error validating the session: %s\n", err)
		return nil, nil
	}
	row.Scan(&session.Id, &session.UserID, &session.ExpiresAt, &user.Id)

	if time.Now().After(session.ExpiresAt) {
		_, err = db.Exec("delete from sessions where id = $1", session.Id)
		if err != nil {
			log.Printf("error validating session token while deleting expired session: %s\n", err)
		}
		return nil, nil
	}

	if time.Now().Before(session.ExpiresAt.Add(-time.Hour * 24 * 15)) {
		session.ExpiresAt = time.Now().Add(time.Hour * 24 * 30)
		_, err = db.Exec("update sessions set expires_at =  $1  where id = $2", session.ExpiresAt, session.Id)
		if err != nil {
			log.Printf("error validating session token while updating session expiry time: %s\n", err)
		}
	}

	return &session, &user
}

func InvalidateSession(db *sql.DB, sessionId string) {
	_, err := db.Exec("delete from sessions where id = $1", sessionId)
	if err != nil {
		log.Printf("error validating session token while deleting expired session: %s\n", err)
	}

}

//
// func NewSession(db *sql.DB, userId int, role Role) (int64, error) {
// 	res, err := db.Exec(`insert into sessions (user_id,role) values ($1,$2)`, userId, role)
// 	if err != nil {
// 		log.Printf("error: could not create new session %s", err)
// 	}
// 	return res.LastInsertId()
// }
//
// func GetSession(db *sql.DB, session_id int) (Session, error) {
//
// 	session := Session{}
// 	row := db.QueryRow(`select (id, user_id, expiry_time, created_at,role) from sessions`)
// 	err := row.Err()
// 	if err != nil {
// 		return session, err
// 	}
// 	row.Scan(&session.Id, &session.UserID, &session.ExpiryTime, &session.CreatedAt, &session.Role)
// 	return session, nil
// }
//
// func CreateJWT(session Session) string {
// 	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
// 		"user_id":     session.UserID,
// 		"expiry_time": session.ExpiryTime,
// 		"role":        session.Role,
// 	})
// 	string, err := token.SignedString(secret_key)
// 	if err != nil {
// 		log.Panicf("error creating jwt: %s\n", err)
// 	}
// 	return string
// }
//
// func ParseJWT(token_string string) Session {
// 	session := Session{}
//
// 	return session
// }
