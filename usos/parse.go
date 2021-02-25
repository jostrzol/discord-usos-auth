package usos

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func parseUserResponse(body io.Reader) (*User, error) {

	var respUser struct {
		ID         string `json:"id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Programmes []struct {
			ID        string `json:"id"`
			Programme struct {
				Name        string `json:"id"`
				Description struct {
					PL string `json:"pl"`
					EN string `json:"en"`
				} `json:"description"`
			} `json:"programme"`
		} `json:"student_programmes"`
	}

	dat, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(dat, &respUser)
	if err != nil {
		return nil, err
	}
	progs := make([]*Programme, len(respUser.Programmes))
	for i, respProg := range respUser.Programmes {
		progs[i] = &Programme{
			ID:          respProg.ID,
			Name:        respProg.Programme.Name,
			Description: respProg.Programme.Description.PL,
		}
	}
	return &User{
		ID:         respUser.ID,
		FirstName:  respUser.FirstName,
		LastName:   respUser.LastName,
		Programmes: progs,
	}, nil
}

func parseCoursesResponse(activeOnly bool, resp io.Reader) ([]*Course, error) {

	jParsed := make(map[string]interface{})
	err := json.NewDecoder(resp).Decode(&jParsed)
	if err != nil {
		return nil, err
	}

	dateHookFunc := mapstructure.StringToTimeHookFunc("2006-01-02")
	termHookFunc := mapstructure.ComposeDecodeHookFunc(
		sliceToMapInterfaceHookFunc("id"),
		multilangToPLHookFunc,
		dateHookFunc)
	terms := make(map[string]*Term)
	decoderTerm, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: termHookFunc,
		Result:     &terms,
	})
	if err != nil {
		return nil, err
	}
	err = decoderTerm.Decode(jParsed["terms"])
	if err != nil {
		return nil, err
	}

	courseHookFunc := mapstructure.ComposeDecodeHookFunc(
		editionsActiveHookFunc(activeOnly, terms),
		// sliceToMapInterfaceHookFunc("course_id"),
		multilangToPLHookFunc)
	courses := make([]*Course, 0)
	decoderCourse, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: courseHookFunc,
		Result:     &courses,
	})
	if err != nil {
		return nil, err
	}
	err = decoderCourse.Decode(jParsed["course_editions"])
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func multilangToPLHookFunc(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if !(f == reflect.Map && t == reflect.String) {
		return data, nil
	}
	m := data.(map[string]interface{})
	if len(m) != 2 {
		return data, nil
	}
	pl, ok := m["pl"]
	if !ok {
		return data, nil
	}
	_, ok = m["en"]
	if !ok {
		return data, nil
	}
	return pl.(string), nil
}

func sliceToMapInterfaceHookFunc(idKey string) func(f reflect.Value, t reflect.Value) (interface{}, error) {
	return func(f reflect.Value, t reflect.Value) (interface{}, error) {
		if !(f.Kind() == reflect.Slice && t.Kind() == reflect.Map) {
			return f.Interface(), nil
		}
		result := make(map[string]interface{})
		for i := 0; i < f.Len(); i++ {
			m, ok := f.Index(i).Interface().(map[string]interface{})
			if !ok {
				return f.Interface(), nil
			}
			id, ok := m[idKey].(string)
			if !ok {
				return f.Interface(), nil
			}
			result[id] = m
		}
		return result, nil
	}
}

func editionsActiveHookFunc(activeOnly bool, terms map[string]*Term) func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	return func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
		if !(f == reflect.Map && t == reflect.Map) {
			return data, nil
		}
		m, ok := data.(map[string]interface{})
		if !ok {
			return data, nil
		}
		result := make([]interface{}, 0)
		for termID, v := range m {
			s, ok := v.([]interface{})
			if !ok {
				return data, nil
			}
			term, exists := terms[termID]
			if !exists {
				return data, nil
			}
			if activeOnly {
				if term.IsActive() {
					result = s
					break
				}
			} else {
				result = append(result, s...)
			}
		}
		return result, nil
	}
}
