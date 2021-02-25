package utils

import (
	"testing"

	"github.com/Ogurczak/discord-usos-auth/usos"
)

var usosUser *usos.User = &usos.User{
	ID:        "123123",
	FirstName: "Witold",
	LastName:  "Wysota",
	Programmes: []*usos.Programme{
		{ID: "123123",
			Name:        "101C-ISP-IN",
			Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		{ID: "123123",
			Name:        "102C-ISP-IN",
			Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		{ID: "123123",
			Name:        "103C-ISP-IN",
			Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		{ID: "123123",
			Name:        "104C-ISP-IN",
			Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
	},
	Courses: []*usos.Course{
		{
			ID:     "103A-INxxx-ISP-ANMA",
			Name:   "Analiza",
			TermID: "2020Z",
		},
		{
			ID:     "103A-INxxx-ISP-MAKO1",
			Name:   "Analiza",
			TermID: "2020Z",
		},
		{
			ID:     "103A-INxxx-ISP-PIPR",
			Name:   "Analiza",
			TermID: "2020Z",
		},
		{
			ID:     "103A-xxxxx-ISP-PRAUT",
			Name:   "Analiza",
			TermID: "2020Z",
		},
		{
			ID:     "103A-INxxx-ISP-PZSP1",
			Name:   "Analiza",
			TermID: "2020Z",
		},
	},
}

func TestFilterRecProgrammeName(t *testing.T) {
	filter := &usos.User{
		Programmes: []*usos.Programme{
			{Name: "103C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, true)
}
func TestFilterRecProgrammeNameWrong(t *testing.T) {
	filter := &usos.User{
		Programmes: []*usos.Programme{
			{Name: "106C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, false)
}
func TestFilterRecProgrammeNameMany(t *testing.T) {
	filter := &usos.User{
		Programmes: []*usos.Programme{
			{Name: "101C-ISP-IN"},
			{Name: "103C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, true)
}

func TestFilterRecProgrammeNameManyNotAll(t *testing.T) {
	filter := &usos.User{
		Programmes: []*usos.Programme{
			{Name: "101C-ISP-IN"},
			{Name: "105C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, false)
}

func TestFilterRecName(t *testing.T) {
	filter := &usos.User{
		FirstName: "Witold",
		LastName:  "Wysota",
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, true)
}

func TestFilterRecNameWrong(t *testing.T) {
	filter := &usos.User{
		FirstName: "Witold",
		LastName:  "Gawkowski",
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, false)
}

func TestFilterRecCourseID(t *testing.T) {
	filter := &usos.User{
		Courses: []*usos.Course{
			{ID: "103A-INxxx-ISP-ANMA"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, true)
}

func TestFilterRecCourseIDWrong(t *testing.T) {
	filter := &usos.User{
		Courses: []*usos.Course{
			{ID: "103B-INxxx-ISP-PTCY"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, false)
}

func TestFilterRecCourseIDMany(t *testing.T) {
	filter := &usos.User{
		Courses: []*usos.Course{
			{ID: "103A-INxxx-ISP-ANMA"},
			{ID: "103A-INxxx-ISP-PZSP1"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	want := true
	if want != matched {
		t.Errorf("want %v, got %v", want, matched)
	}
}

func TestFilterRecCourseIDManyNotAll(t *testing.T) {
	filter := &usos.User{
		Courses: []*usos.Course{
			{ID: "103A-INxxx-ISP-ANMA"},
			{ID: "103B-INxxx-ISP-PTCY"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, false)
}

func TestFilterRecProgrammeNameCourseID(t *testing.T) {
	filter := &usos.User{
		Programmes: []*usos.Programme{
			{Name: "103C-ISP-IN"},
		},
		Courses: []*usos.Course{
			{ID: "103A-INxxx-ISP-ANMA"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	if err != nil {
		t.Error(err)
	}
	assert(t, matched, true)

}

func assert(t *testing.T, got interface{}, want interface{}) {
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}
