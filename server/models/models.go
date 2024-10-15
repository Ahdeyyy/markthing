package models

type Visibility = string

const (
	PRIVATE = "private"
	PUBLIC  = "public"
)

type Workspace struct {
	Id         int
	UserID     int
	Name       string
	Tags       string // each tag is separated by a comma
	Visibility Visibility
}

type Note struct {
	Id          int
	Name        string
	WorkspaceID int
	ParentID    int
	Tags        string
	Content     string
	Visibility  Visibility
}

type User struct {
	Id       int
	Username string
	Password string
}
