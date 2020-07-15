package controllers

import (
	"github.com/jinzhu/gorm"
	"net/http"
	"sibsiu.ru/context"
	"sibsiu.ru/models"
	"sibsiu.ru/views"
	"strconv"
)

type Skips struct {
	EditView *views.View
	HomeView *views.View
	ShowView *views.View

	ss models.SkipService
}

func NewSkips(ss models.SkipService) *Skips {
	return &Skips{
		ShowView: views.NewView("bootstrap", "skips/show"),
		EditView: views.NewView("bootstrap", "skips/edit"),
		HomeView: views.NewView("bootstrap", "static/home"),

		ss: ss,
	}
}

// GET /skips
func (s *Skips) Show(w http.ResponseWriter, r *http.Request) {
	var studentParams models.SkipsStudentQueryParams
	var starostaParams models.SkipsStarostaQueryParams

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

// GET /skips/edit
func (s *Skips) Edit(w http.ResponseWriter, r *http.Request) {
	var starostaParams models.SkipsEditStarostaQueryParams

	var err error
	var vd views.Data

	if _, err = parseURL(r, &starostaParams); err != nil {
		vd.SetAlert(err)
		s.HomeView.Render(w, r, vd)
		return
	}

	user := context.User(r.Context())

	result, err := s.ss.Edit(user, &starostaParams)
	if err != nil {
		vd.SetAlert(err)
		s.HomeView.Render(w, r, vd)
		return
	}

	vd.Yield = *result
	vd.ScriptsPaths = []string{"skips/edit/starosta.js"}
	s.EditView.Render(w, r, vd)
}

// POST /api/skips/update
func (s *Skips) Update(w http.ResponseWriter, r *http.Request) {

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

		err = s.ss.Update(&models.Skip{Model: gorm.Model{ID: uint(id)}, Reason: v[0]})
		if err != nil {
			views.RenderJSON(w, r, nil, err)
			return
		}
	}

	views.RenderJSON(w, r, "Пропуски обновлены.", nil)
}

// GET /api/filter/skips
func (s *Skips) Filter(w http.ResponseWriter, r *http.Request) {

	var studentParams models.SkipsStudentQueryParams
	var starostaParams models.SkipsStarostaQueryParams

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

	result, err := s.ss.ByUserFilterData(user, getParams)
	if err != nil {
		views.RenderJSON(w, r, nil, err)
		return
	}

	views.RenderJSON(w, r, result, nil)

}

// GET /api/filter/skips/edit
func (s *Skips) FilterEdit(w http.ResponseWriter, r *http.Request) {

	var filterURLData models.SkipsEditStarostaQueryParams

	if _, err := parseURL(r, &filterURLData); err != nil {
		views.RenderJSON(w, r, nil, err)
		return
	}

	user := context.User(r.Context())

	result, err := s.ss.ByUserEditFilterData(user, &filterURLData)
	if err != nil {
		views.RenderJSON(w, r, nil, err)
		return
	}

	if result != nil {
		views.RenderJSON(w, r, result.StarostaSkipsEditFilterData, nil)
	} else {
		views.RenderJSON(w, r, nil, nil)
	}
}
