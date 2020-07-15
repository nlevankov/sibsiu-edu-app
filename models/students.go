package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

type Student struct {
	gorm.Model
	FirstName  string
	MiddleName string // отчество
	LastName   string
	Numbers    []Number // Номера зачеток, без тегов возможно не будет подгружаться
	Users      []*User  `gorm:"many2many:users_students"`
	Groups     []*Group `gorm:"many2many:students_groups"`
	Scores     []Score
	Skips      []Skip
	GroupID    uint
}
