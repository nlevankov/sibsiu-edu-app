package controllers

import (
	"github.com/jinzhu/gorm"
	"net/http"
	"sibsiu.ru/context"
	"sibsiu.ru/models"
	"sibsiu.ru/views"
	"strconv"
)

type Groups struct {
	StatusesView *views.View
	gs           models.GroupService
}

func NewGroups(gs models.GroupService) *Groups {
	return &Groups{
		StatusesView: views.NewView("bootstrap", "groups/statuses"),
		gs:           gs,
	}
}

// GET /groups/statuses
func (g *Groups) ShowStatuses(w http.ResponseWriter, r *http.Request) {
	var err error
	var vd views.Data

	user := context.User(r.Context())

	result, err := g.gs.ByUserStatuses(user)
	if err != nil {
		return
	}

	vd.Yield = result
	vd.ScriptsPaths = []string{"groups/statuses/starosta.js"}
	g.StatusesView.Render(w, r, vd)
}

// POST /api/groups/statuses/update (по аналогии со scores)
func (g *Groups) UpdateStatuses(w http.ResponseWriter, r *http.Request) {
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

		status := "Неактивная"
		if len(v) == 2 {
			status = "Активная"
		}

		err = g.gs.Update(&models.Group{Model: gorm.Model{ID: uint(id)}, Status: status})
		if err != nil {
			views.RenderJSON(w, r, nil, err)
			return
		}
	}

	views.RenderJSON(w, r, "Статусы обновлены.", nil)
}
