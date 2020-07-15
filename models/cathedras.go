package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

// Кафедра не сделана просто полем у какой-нибудь структуры по той же причине, что и
// Number
type Cathedra struct {
	gorm.Model
	Name        string       `gorm:"unique_index"`
	Disciplines []Discipline `gorm:"foreignkey:Cathedra;association_foreignkey:Name"`
	Teachers    []*Teacher   `gorm:"many2many:cathedras_teachers"`
}
