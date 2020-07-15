package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

type StarostaGroupStatuses struct {
	GroupName string
	GroupID   uint
	IsActual  bool
}

type GroupService interface {
	GroupDB
}

type GroupDB interface {
	ByUserStatuses(user *User) ([]StarostaGroupStatuses, error)
	Update(group *Group) error
}

var _ GroupService = &groupService{}

type groupService struct {
	GroupDB
}

func NewGroupService(db *gorm.DB) GroupService {
	gg := &groupGorm{db}

	return &groupService{
		GroupDB: gg,
	}
}

var _ GroupDB = &groupGorm{}

type groupGorm struct {
	db *gorm.DB
}

func (gg *groupGorm) ByUserStatuses(user *User) ([]StarostaGroupStatuses, error) {
	var studentIDs []uint
	var Starosta Student

	gg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	err := gg.db.
		Where("id = ?", studentIDs[0]).
		Preload("Groups.Starosta").
		First(&Starosta).
		Error
	if err != nil {
		return nil, err
	}

	result := []StarostaGroupStatuses{}

	for _, v := range Starosta.Groups {
		if v.Starosta.ID == Starosta.ID {

			var isActual bool
			if v.Status == "Активная" {
				isActual = true
			}

			result = append(result, StarostaGroupStatuses{GroupName: v.Name, GroupID: v.ID, IsActual: isActual})
		}
	}

	return result, err
}

func (gg *groupGorm) Update(group *Group) error {
	return gg.db.Model(&Group{}).Update(group).Error
}

type Group struct {
	gorm.Model
	Name     string     `gorm:"not null;unique_index"`
	Students []*Student `gorm:"many2many:students_groups"` //`gorm:"many2many:students_groups;foreignkey:Name;association_foreignkey:ID"`
	Status   string     `sql:"type:group_status"`
	//Status bool // пока оставлю логический тип, дабы избавиться от if-ов лучше использовать enum тип, если есть [возможность]
	Year               uint          // год начала обучения
	EducationPeriod    int           // период обучения, например, 4
	Numbers            []Number      `gorm:"foreignkey:Group;association_foreignkey:Name"`
	EducationForm      string        `sql:"type:education_form"`
	EducationLevel     string        `sql:"type:education_level"`
	Institute          string        `sql:"type:institute"`
	Teachers           []*Teacher    `gorm:"many2many:groups_teachers"`
	Disciplines        []*Discipline `gorm:"many2many:disciplines_groups"`
	Starosta           Student
	Schedules          []Schedule
	GroupSemesterInfos []GroupSemesterInfo
}
