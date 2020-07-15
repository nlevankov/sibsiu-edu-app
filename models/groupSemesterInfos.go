package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"time"
)

type GroupSemesterInfo struct {
	gorm.Model
	Semester           string `sql:"type:semester"`
	Year               uint   // год начала обучения
	GroupID            uint
	EducationStartDate time.Time
	EducationEndDate   time.Time
	DisciplinesIDs     pq.Int64Array `gorm:"type:integer[]"`
}
