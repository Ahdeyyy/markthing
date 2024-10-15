package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"

	"encoding/base32"
	"log"
	"markthing/repository"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Role = string

const (
	ADMIN = "admin"
	USER  = "user"
	GUEST = "guest"
)

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
	buf := make([]byte, 20)
	rand.Reader.Read(buf)
	token := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	return strings.ToLower(token)
}

func CreateSession(db *pgx.Conn, token string, userId int) repository.Session {
	query := repository.New(db)
	sessionId := encodeHex(token)

	sess := repository.CreateSessionParams{
		ID:     sessionId,
		UserID: int32(userId),
		ExpiresAt: pgtype.Timestamp{
			Time:             time.Now().Add(time.Hour * 24 * 30),
			InfinityModifier: pgtype.Finite,
			Valid:            true,
		},
	}

	session, err := query.CreateSession(context.Background(), sess)

	if err != nil {
		log.Printf("error: could not create new session %s", err)
		session.ID = "Error"
		return session
	}
	return session
}
func ValidateSessionToken(db *pgx.Conn, token string) repository.GetSessionRow {
	query := repository.New(db)
	sessionId := encodeHex(token)
	session, err := query.GetSession(context.Background(), sessionId)

	if err != nil {
		log.Printf("error validating the session: %s,  %s\n", sessionId, err)
		return repository.GetSessionRow{}
	}

	expiresAt := session.Session.ExpiresAt.Time
	if time.Now().After(expiresAt) {
		err = query.DeleteSession(context.Background(), sessionId)
		if err != nil {
			log.Printf("error validating session token while deleting expired session: %s\n", err)
		}
		return repository.GetSessionRow{}
	}

	if time.Now().Before(expiresAt.Add(-time.Hour * 24 * 15)) {
		session.Session.ExpiresAt.Time = time.Now().Add(time.Hour * 24 * 30)
		err = query.UpdateSessionExpiresAt(context.Background(), repository.UpdateSessionExpiresAtParams{
			ExpiresAt: session.Session.ExpiresAt,
			ID:        session.Session.ID,
		})
		if err != nil {
			log.Printf("error validating session token while updating session expiry time: %s\n", err)
		}
	}

	return session
}

func InvalidateSession(db *pgx.Conn, sessionId string) {
	query := repository.New(db)
	err := query.DeleteSession(context.Background(), sessionId)
	if err != nil {
		log.Printf("error validating session token while deleting expired session: %s\n", err)
	}

}
