package usos

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/mitchellh/mapstructure"
)

var (
	activeTermID               string
	lastActiveTermIDUpdateDate time.Time
)

// checks if the given term is the active term
func termIsActive(client *http.Client, termID string) (bool, error) {
	nowDate := time.Now().Local().Truncate(time.Hour * 24)
	if nowDate == lastActiveTermIDUpdateDate {
		return activeTermID == termID, nil
	}

	dat, err := makeCall(client, true, "term", termID)
	if err != nil {
		return false, err
	}
	parsed := make(map[string]interface{})
	err = json.Unmarshal(dat, &parsed)
	if err != nil {
		return false, err
	}
	startDate, err := time.Parse("2006-01-02", parsed["start_date"].(string))
	if err != nil {
		return false, err
	}
	finishDate, err := time.Parse("2006-01-02", parsed["finish_date"].(string))
	if err != nil {
		return false, err
	}

	if nowDate.After(startDate) && nowDate.Before(finishDate) {
		activeTermID = termID
		lastActiveTermIDUpdateDate = nowDate
		return true, nil
	}

	return false, nil
}

// Course represents an usos course
type Course struct {
	ID     string
	Name   string
	TermID string
}

// Programme represents an usos student programme
type Programme struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// User represents an usos user
type User struct {
	ID         string      `json:"id,omitempty"`
	FirstName  string      `json:"first_name,omitempty"`
	LastName   string      `json:"last_name,omitempty"`
	Programmes []Programme `json:"student_programmes,omitempty"`
	Courses    []*Course   `json:"student_courses,omitempty"`

	token *oauth1.Token
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
type tmpCourse struct {
	ID     string        `mapstructure:"course_id"`
	Name   tmpCourseName `mapstructure:"course_name"`
	TermID string        `mapstructure:"term_id"`
}
type tmpCourseName struct {
	PL string `mapstructure:"pl"`
	EN string `mapstructure:"en"`
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

func newCourseFromTmpCourse(tc *tmpCourse) *Course {
	return &Course{
		ID:     tc.ID,
		Name:   tc.Name.PL,
		TermID: tc.TermID,
	}
}

// NewUsosUser returns an UsosUser object initialized from api calls using the given access token
func NewUsosUser(token *oauth1.Token) (*User, error) {
	client := config.Client(oauth1.NoContext, token)

	body, err := makeCall(client, false, "user", "id|first_name|last_name|student_programmes")
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
	user.token = token
	return user, nil
}

func (u *User) client() *http.Client {
	return config.Client(oauth1.NoContext, u.token)
}

// GetCourses returns and assigns his currently active courses to the user
func (u *User) GetCourses(activeOnly bool) ([]*Course, error) {
	client := u.client()
	dat, err := makeCall(client, true, "courses", "course_editions|terms")
	if err != nil {
		return nil, err
	}
	parsed := make(map[string]interface{})
	err = json.Unmarshal(dat, &parsed)
	if err != nil {
		return nil, err
	}

	var tmpCourses []*tmpCourse
	if activeOnly {
		for termID, _courses := range parsed["course_editions"].(map[string]interface{}) {
			courses := _courses.([]interface{})
			isActive, err := termIsActive(client, termID)
			if err != nil {
				return nil, err
			}
			if isActive {
				tmpCourses = make([]*tmpCourse, len(courses))
				for i, course := range courses {
					err := mapstructure.Decode(course, &tmpCourses[i])
					if err != nil {
						return nil, err
					}
				}
				break
			}
		}
		if tmpCourses == nil {
			// no active courses, return an empty slice
			tmpCourses = make([]*tmpCourse, 0)
		}
	} else {
		allCourses := make([]interface{}, 0)
		for _, _courses := range parsed["course_editions"].(map[string]interface{}) {
			courses := _courses.([]interface{})
			allCourses = append(allCourses, courses...)
		}
		tmpCourses = make([]*tmpCourse, len(allCourses))
		for i, course := range allCourses {
			err := mapstructure.Decode(course, &tmpCourses[i])
			if err != nil {
				return nil, err
			}
		}
	}
	courses := make([]*Course, len(tmpCourses))
	for i, tmpCourse := range tmpCourses {
		courses[i] = newCourseFromTmpCourse(tmpCourse)
	}
	return courses, nil
}
