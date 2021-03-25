package usos

import (
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
)

// Term represents an usos term
type Term struct {
	ID         string    `mapstructure:"id"`
	Name       string    `mapstructure:"name"`
	StartDate  time.Time `mapstructure:"start_date"`
	EndDate    time.Time `mapstructure:"end_date"`
	FinishDate time.Time `mapstructure:"finish_date"`
}

// IsActive checks if the term is active
func (t *Term) IsActive() bool {
	now := time.Now()
	return now.After(t.StartDate) && now.Before(t.EndDate.AddDate(0, 0, 1)) || now == t.StartDate
}

// Course represents an usos course
type Course struct {
	ID     string `json:"course_id,omitempty" mapstructure:"course_id"`
	Name   string `json:"course_name,omitempty" mapstructure:"course_name"`
	TermID string `json:"term_id,omitempty" mapstructure:"term_id"`
}

// Programme represents an usos student programme
type Programme struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// User represents an usos user
type User struct {
	ID         string       `json:"id,omitempty"`
	FirstName  string       `json:"first_name,omitempty"`
	LastName   string       `json:"last_name,omitempty"`
	Programmes []*Programme `json:"student_programmes,omitempty"`
	Courses    []*Course    `json:"student_courses,omitempty"`

	token *oauth1.Token
}

// NewUsosUser returns an UsosUser object initialized from api calls using the given access token
func NewUsosUser(token *oauth1.Token) (*User, error) {
	client := config.Client(oauth1.NoContext, token)

	resp, err := makeCall(client, "user", "id|first_name|last_name|student_programmes")
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	user, err := parseUserResponse(resp)
	if err != nil {
		return nil, err
	}
	user.token = token
	return user, nil
}

func (u *User) client() *http.Client {
	return config.Client(oauth1.NoContext, u.token)
}

// GetCourses returns and assigns his currently active courses to the user
func (u *User) GetCourses(activeOnly bool) ([]*Course, error) {
	client := u.client()
	resp, err := makeCall(client, "courses", "course_editions|terms")
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	courses, err := parseCoursesResponse(activeOnly, resp)
	if err != nil {
		return nil, err
	}

	u.Courses = courses
	return courses, nil
}

// GetCoursesLight returns and assigns his currently active courses to the user,
// does not download unneeded information
func (u *User) GetCoursesLight(activeOnly bool) ([]*Course, error) {
	client := u.client()
	resp, err := makeCall(client, "groups", "course_id|term_id|course_name", activeOnly)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	courses, err := parseGroupsResponseToCourses(activeOnly, resp)
	if err != nil {
		return nil, err
	}

	u.Courses = courses
	// dat, err := ioutil.ReadAll(resp)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Print(string(dat))
	return courses, nil
}
