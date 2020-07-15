package controllers

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"net/http"
	"sibsiu.ru/context"
	"sibsiu.ru/models"
	"sibsiu.ru/views"
	"strconv"
)

type Scores struct {
	HomeView *views.View
	ShowView *views.View
	ss       models.ScoreService
}

func NewScores(ss models.ScoreService) *Scores {
	return &Scores{
		HomeView: views.NewView("bootstrap", "static/home"),
		ShowView: views.NewView("bootstrap", "scores/show"),
		ss:       ss,
	}
}

// GET /scores
func (s *Scores) Show(w http.ResponseWriter, r *http.Request) {

	var studentParams models.ScoresStudentQueryParams
	var starostaParams models.ScoresStarostaQueryParams

	var getParams interface{}
	var err error
	var vd views.Data

	// важно не менять порядок парсинга структур
	if getParams, err = parseURL(r, &studentParams, &starostaParams); err != nil {
		vd.SetAlert(err)
		s.HomeView.Render(w, r, vd)
		return
	}

	user := context.User(r.Context())

	result, scriptsPaths, err := s.ss.ByUser(user, getParams)
	if err != nil {
		vd.SetAlert(err)
		s.HomeView.Render(w, r, vd)
		return
	}

	vd.Yield = *result
	vd.ScriptsPaths = scriptsPaths
	s.ShowView.Render(w, r, vd)
}

// GET /api/filter/scores
func (s *Scores) Filter(w http.ResponseWriter, r *http.Request) {

	var filterURLData models.FilterScoresQueryParams

	if _, err := parseURL(r, &filterURLData); err != nil {
		views.RenderJSON(w, r, nil, err)
		return
	}

	user := context.User(r.Context())

	result, err := s.ss.ByUserFilterData(user, &filterURLData)
	if err != nil {
		views.RenderJSON(w, r, nil, err)
		return
	}

	views.RenderJSON(w, r, result.StudentScoresFilterData, nil)

	return
}

// POST /api/scores/update
func (s *Scores) Update(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		views.RenderJSON(w, r, nil, err)
		return
	}

	delete(r.PostForm, "gorilla.csrf.Token")

	for k, v := range r.PostForm {

		id, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			views.RenderJSON(w, r, nil, err)
			return
		}

		err = s.ss.Update(&models.Score{Model: gorm.Model{ID: uint(id)}, AttestationOne: v[0], AttestationTwo: v[1], AttestationThree: v[2], IntermediateAttestation: v[3]})

		if err != nil {
			views.RenderJSON(w, r, nil, err)
			return
		}
	}

	views.RenderJSON(w, r, "Оценки обновлены.", nil)
}

func (s *Scores) Report(w http.ResponseWriter, r *http.Request) {

	var starostaParams models.ScoresStarostaQueryParams
	var err error

	if _, err = parseURL(r, &starostaParams); err != nil {
		return
	}

	user := context.User(r.Context())

	f, err := s.ss.ReportByUser(user, &starostaParams)
	if err != nil {
		return
	}

	views.RenderXLSX(w, f)
}

func PrepareAndReturnExcel() *excelize.File {
	f := excelize.NewFile()
	style, _ := f.NewStyle(`{"alignment":{"horizontal":"center","vertical":"center"}}`)
	f.SetColStyle("Sheet1", "A:Z", style)
	f.SetCellValue("Sheet1", "A1", "Usernameeeeee")
	f.SetCellValue("Sheet1", "A2", 1)
	f.SetCellValue("Sheet1", "B1", "Location")
	f.SetCellValue("Sheet1", "B2", 2)
	f.SetCellValue("Sheet1", "C1", "Occupation")
	f.SetCellValue("Sheet1", "C2", 3)

	return f
}
