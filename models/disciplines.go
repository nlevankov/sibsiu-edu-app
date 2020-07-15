package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

type Discipline struct {
	gorm.Model
	Name     string
	Cathedra string
	Teachers []*Teacher `gorm:"many2many:disciplines_teachers"`
	Skips    []Skip
	Scores   []Score
	Groups   []*Group `gorm:"many2many:disciplines_groups"`
}
