package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"time"
)

type Schedule struct {
	gorm.Model
	Day time.Weekday
	//Disciplines []Discipline // не сейвит
	DisciplinesIDs pq.Int64Array `gorm:"type:integer[]"`
	Semester       string        `sql:"type:semester"`
	Year           uint          // год начала обучения
	GroupID        uint
	IsEvenWeek     bool
}
