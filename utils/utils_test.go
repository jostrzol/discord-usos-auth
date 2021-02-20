package utils

import (
	"testing"

	"github.com/Ogurczak/discord-usos-auth/usos"
)

func TestFilterRecProgrammeName(t *testing.T) {
	usosUser := &usos.User{
		ID:        "123123",
		FirstName: "Witold",
		LastName:  "Wysota",
		Programmes: []usos.Programme{
			{ID: "123123",
				Name:        "103C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		},
	}
	filter := &usos.User{
		Programmes: []usos.Programme{
			{Name: "103C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	want := true
	if err != nil {
		t.Error(err)
	}
	if matched != want {
		t.Errorf("want %v, got %v", want, matched)
	}
}
func TestFilterRecProgrammeNameWrong(t *testing.T) {
	usosUser := &usos.User{
		ID:        "123123",
		FirstName: "Witold",
		LastName:  "Wysota",
		Programmes: []usos.Programme{
			{ID: "123123",
				Name:        "102C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		},
	}
	filter := &usos.User{
		Programmes: []usos.Programme{
			{Name: "103C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	want := false
	if err != nil {
		t.Error(err)
	}
	if matched != want {
		t.Errorf("want %v, got %v", want, matched)
	}
}
func TestFilterRecProgrammeNameMany(t *testing.T) {
	usosUser := &usos.User{
		ID:        "123123",
		FirstName: "Witold",
		LastName:  "Wysota",
		Programmes: []usos.Programme{
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
	}
	filter := &usos.User{
		Programmes: []usos.Programme{
			{Name: "101C-ISP-IN"},
			{Name: "103C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	want := true
	if err != nil {
		t.Error(err)
	}
	if matched != want {
		t.Errorf("want %v, got %v", want, matched)
	}
}

func TestFilterRecProgrammeNameManyNotAll(t *testing.T) {
	usosUser := &usos.User{
		ID:        "123123",
		FirstName: "Witold",
		LastName:  "Wysota",
		Programmes: []usos.Programme{
			{ID: "123123",
				Name:        "101C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
			{ID: "123123",
				Name:        "102C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
			{ID: "123123",
				Name:        "104C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		},
	}
	filter := &usos.User{
		Programmes: []usos.Programme{
			{Name: "101C-ISP-IN"},
			{Name: "103C-ISP-IN"},
		},
	}
	matched, err := FilterRec(filter, usosUser)
	want := false
	if err != nil {
		t.Error(err)
	}
	if matched != want {
		t.Errorf("want %v, got %v", want, matched)
	}
}

func TestFilterRecName(t *testing.T) {
	usosUser := &usos.User{
		ID:        "123123",
		FirstName: "Witold",
		LastName:  "Wysota",
		Programmes: []usos.Programme{
			{ID: "123123",
				Name:        "101C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
			{ID: "123123",
				Name:        "102C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
			{ID: "123123",
				Name:        "104C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		},
	}
	filter := &usos.User{
		FirstName: "Witold",
		LastName:  "Wysota",
	}
	matched, err := FilterRec(filter, usosUser)
	want := true
	if err != nil {
		t.Error(err)
	}
	if matched != want {
		t.Errorf("want %v, got %v", want, matched)
	}
}

func TestFilterRecNameWrong(t *testing.T) {
	usosUser := &usos.User{
		ID:        "123123",
		FirstName: "Witold",
		LastName:  "Wysota",
		Programmes: []usos.Programme{
			{ID: "123123",
				Name:        "101C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
			{ID: "123123",
				Name:        "102C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
			{ID: "123123",
				Name:        "104C-ISP-IN",
				Description: "Informatyka, studia stacjonarne pierwszego stopnia"},
		},
	}
	filter := &usos.User{
		FirstName: "Witold",
		LastName:  "Gawkowski",
	}
	matched, err := FilterRec(filter, usosUser)
	want := false
	if err != nil {
		t.Error(err)
	}
	if matched != want {
		t.Errorf("want %v, got %v", want, matched)
	}
}
