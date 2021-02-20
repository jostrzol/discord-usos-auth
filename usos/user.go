package usos

import (
	"encoding/json"

	"github.com/dghubble/oauth1"
)

// Programme represents an usos group
type Programme struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// User represents an usos user
type User struct {
	ID         string      `json:"id"`
	FirstName  string      `json:"first_name"`
	LastName   string      `json:"last_name"`
	Programmes []Programme `json:"student_programmes"`
}

type tmpDescription struct {
	PL string `json:"pl"`
	EN string `json:"en"`
}

type tmpSubProgramme struct {
	Name        string         `json:"id"`
	Description tmpDescription `json:"description"`
}

type tmpProgramme struct {
	ID        string          `json:"id"`
	Programme tmpSubProgramme `json:"programme"`
}

type tmpUser struct {
	ID         string         `json:"id"`
	FirstName  string         `json:"first_name"`
	LastName   string         `json:"last_name"`
	Programmes []tmpProgramme `json:"student_programmes"`
}

func newUserFromTmpUser(tu *tmpUser) *User {
	var progs []Programme
	for _, p := range tu.Programmes {
		progs = append(progs, Programme{
			ID:          p.ID,
			Name:        p.Programme.Name,
			Description: p.Programme.Description.PL,
		})
	}
	return &User{
		ID:         tu.ID,
		FirstName:  tu.FirstName,
		LastName:   tu.LastName,
		Programmes: progs,
	}
}

// NewUsosUser returns an UsosUser object initialized from api calls using the given access token
func NewUsosUser(token *oauth1.Token) (*User, error) {
	client := config.Client(oauth1.NoContext, token)

	body, err := makeCall(client, "user", "id|first_name|last_name|student_programmes")
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Raw Response Body:\n%v\n", string(body))

	tmp := tmpUser{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return nil, err
	}

	user := newUserFromTmpUser(&tmp)
	return user, nil
}
