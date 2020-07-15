package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

type User struct {
	gorm.Model
	FirstName    string
	MiddleName   string // отчество
	LastName     string
	Email        string     //`gorm:"not null;unique_index"`
	Password     string     `gorm:"-"`
	PasswordHash string     `gorm:"not null"`
	Remember     string     `gorm:"-"`
	RememberHash string     `gorm:"not null;unique_index"`
	Login        string     `gorm:"not null;unique_index"`
	Class        string     `sql:"type:class"`
	Students     []*Student `gorm:"many2many:users_students"`
}
