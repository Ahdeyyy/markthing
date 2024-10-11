package main

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

func NewWorkspace(name string, tags string) Workspace {
	return Workspace{
		Id:         0, //randomise or use db to generate
		Name:       name,
		Tags:       tags,
		Visibility: PRIVATE,
	}
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
