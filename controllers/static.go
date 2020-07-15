package controllers

import (
	"log"
	"net/http"
	"sibsiu.ru/models"
	"sibsiu.ru/views"
)

func NewStatic(ss models.ScoreService) *Static {
	return &Static{
		HomeView: views.NewView("bootstrap", "static/home"),
		//Contact:  views.NewView("bootstrap", "static/contact"),
		ss: ss,
	}
}

type Static struct {
	HomeView *views.View
	//Contact  *views.View
	ss models.ScoreService
}

func (s *Static) Root(w http.ResponseWriter, r *http.Request) {
	time, err := s.ss.LastUpdated()
	if err != nil {
		log.Println(err)
	}

	var vd views.Data
	vd.Yield = time
	s.HomeView.Render(w, r, vd)
}
