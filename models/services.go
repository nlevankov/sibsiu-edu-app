package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

// ServicesConfig is really just a function, but I find using
// types like this are easier to read in my source code.
type ServicesConfig func(*Services) error

type Services struct {
	User UserService

	Score ScoreService
	Group GroupService
	Skip  SkipService
	db    *gorm.DB
}

// NewServices now will accept a list of config functions to
// run. Each function will accept a pointer to the current
// Services object as its only argument and will edit that
// object inline and return an error if there is one. Once
// we have run all configs we will return the Services object.
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	// For each ServicesConfig function...
	for _, cfg := range cfgs {
		// Run the function passing in a pointer to our Services
		// object and catching any errors
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	// Then finally return the result
	return &s, nil
}

func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		s.db = db

		return nil
	}
}

func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)

		return nil
	}
}

func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db, pepper, hmacKey)
		return nil
	}
}

func WithScore() ServicesConfig {
	return func(s *Services) error {
		s.Score = NewScoreService(s.db)
		return nil
	}
}

func WithGroup() ServicesConfig {
	return func(s *Services) error {
		s.Group = NewGroupService(s.db)
		return nil
	}
}

func WithSkip() ServicesConfig {
	return func(s *Services) error {
		s.Skip = NewSkipService(s.db)
		return nil
	}
}

func WithTesting() ServicesConfig {
	return func(s *Services) error {

		setSchema(s.db)
		seedDB(s.db, s.User)

		return nil
	}
}

// Closes the database connection
func (s *Services) Close() error {
	return s.db.Close()
}

// AutoMigrate will attempt to automatically migrate all tables
// todo по мере реализации моделей ресурсов, нуно добавлять их в AutoMigrate ниже
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&Cathedra{},
		Discipline{},
		Group{},
		GroupSemesterInfo{},
		Number{},
		Schedule{},
		Score{},
		Skip{},
		Student{},
		Teacher{},
		User{},
	).Error
}

// DestructiveReset drops all tables and rebuilds them
// todo по мере реализации моделей ресурсов, нуно добавлять их в DestructiveReset ниже
func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&Cathedra{},
		Discipline{},
		Group{},
		GroupSemesterInfo{},
		Number{},
		Schedule{},
		Score{},
		Skip{},
		Student{},
		Teacher{},
		User{},
	).Error

	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

func setSchema(db *gorm.DB) {

	//db.Debug().Exec("DROP DATABASE IF EXISTS sibsiu_dev")
	//db.Debug().Exec("CREATE DATABASE sibsiu_dev")
	db.Debug().Exec("DROP SCHEMA public CASCADE")
	db.Debug().Exec("CREATE SCHEMA public")

	db.Debug().Exec(`CREATE TYPE class AS ENUM (
	'Студент',
	'Староста',
	'');
	`)
	db.Debug().Exec(`CREATE TYPE group_status AS ENUM (
	'',
	'Активная',
	'Неактивная');
	`)
	db.Debug().Exec(`CREATE TYPE education_form AS ENUM (
	'',
	'Очная',
	'Заочная');
	`)
	db.Debug().Exec(`CREATE TYPE education_level AS ENUM (
	'',
	'Бакалавриат');
	`)
	db.Debug().Exec(`CREATE TYPE institute AS ENUM (
	'',
	'ИИТИАС');
	`)
	db.Debug().Exec(`CREATE TYPE score_value AS ENUM (
	'Н/А',
	'Зачтено',
	'Не зачтено',
	'5',
	'4',
	'3',
	'2',
	'');
	`)
	db.Debug().Exec(`CREATE TYPE semester AS ENUM (
	'Весенний',
	'Осенний',
	'');
	`)
	db.Debug().Exec(`CREATE TYPE reason AS ENUM (
	'У',
	'Н',
	'');
	`)

	db.CreateTable(
		&User{},
		&Student{},
		&Number{},
		&Group{},
		&Score{},
		&Discipline{},
		&Cathedra{},
		&Skip{},
		&Teacher{},
		&Schedule{},
		&GroupSemesterInfo{},
	)

	// При добавлении foreign key не забывай добавлять соотв. служебные записи ниже
	// логично добавлять foreign key со стороны подчиненных (child) таблиц, ведь ключи именно для них считаются foreign (внешние по отношению к этим табилцам)
	db.Model(&GroupSemesterInfo{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "CASCADE")
	db.Model(&Schedule{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "CASCADE")
	db.Model(&Student{}).AddForeignKey("group_id", "groups(id)", "CASCADE", "CASCADE")
	db.Model(&Number{}).AddForeignKey("student_id", "students(id)", "CASCADE", "CASCADE")
	db.Model(&Number{}).AddForeignKey("group", "groups(name)", "CASCADE", "CASCADE")
	db.Model(&Score{}).AddForeignKey("student_id", "students(id)", "CASCADE", "CASCADE")
	db.Model(&Score{}).AddForeignKey("discipline_id", "disciplines(id)", "CASCADE", "CASCADE")
	db.Model(&Discipline{}).AddForeignKey("cathedra", "cathedras(name)", "CASCADE", "CASCADE")
	db.Model(&Skip{}).AddForeignKey("discipline_id", "disciplines(id)", "CASCADE", "CASCADE")
	db.Model(&Skip{}).AddForeignKey("student_id", "students(id)", "CASCADE", "CASCADE")

	// Служебные (0-е) записи
	db.Exec(`INSERT  INTO "users" ("id","password_hash","remember_hash","login") VALUES (0,'','','')`)
	db.Exec(`INSERT  INTO "students" ("id") VALUES (0)`)
	db.Exec(`INSERT  INTO "numbers" ("id") VALUES (0)`)
	db.Exec(`INSERT  INTO "scores" ("id") VALUES (0)`)
	db.Exec(`INSERT  INTO "cathedras" ("id","name") VALUES (0,'')`)
	db.Exec(`INSERT  INTO "disciplines" ("id","name") VALUES (0,'')`)
	db.Exec(`INSERT  INTO "groups" ("id","name") VALUES (0,'')`)
}

func seedDB(db *gorm.DB, us UserService) {

	// Наполнение записями

	passwd := "123123123"
	class := "Студент"

	users := []User{
		{
			FirstName:  "Никита",
			MiddleName: "Владимирович",
			LastName:   "Леванков",
			Login:      "levankov_nv",
			Email:      "kellerock@yandex.ru",
			Password:   passwd,
			Class:      class,
		},
		{
			FirstName:  "Николай",
			MiddleName: "Юрьевич",
			LastName:   "Ефимов",
			Login:      "efimov_nu",
			Email:      "Efimov@yandex.ru",
			Password:   passwd,
			Class:      class,
		},
		{
			FirstName:  "Анастасия",
			MiddleName: "Григорьевна",
			LastName:   "Дмитриева",
			Login:      "dmitrieva_ag",
			Email:      "Dmitrieva@yandex.ru",
			Password:   passwd,
			Class:      "Староста",
		},
		{
			FirstName:  "Иван",
			MiddleName: "Иванович",
			LastName:   "Иванов",
			Login:      "ivanov_ii",
			Email:      "Ivanov@yandex.ru",
			Password:   passwd,
			Class:      class,
		},
		{
			FirstName:  "Василий",
			MiddleName: "Васильевич",
			LastName:   "Васильев",
			Login:      "vasilyev_vv",
			Email:      "vasilyev@yandex.ru",
			Password:   passwd,
			Class:      class,
		},
		{
			FirstName:  "Юрий",
			MiddleName: "Александрович",
			LastName:   "Завьялов",
			Login:      "zavyalov_ua",
			Email:      "zavyalov@yandex.ru",
			Password:   passwd,
			Class:      "Староста",
		},
	}
	for i := range users {
		err := us.Create(&users[i])
		if err != nil {
			panic(err)
		}
	}

	students := []Student{
		{
			FirstName:  "Никита",
			MiddleName: "Владимирович",
			LastName:   "Леванков",
		},
		{
			FirstName:  "Николай",
			MiddleName: "Юрьевич",
			LastName:   "Ефимов",
		},
		{
			FirstName:  "Анастасия",
			MiddleName: "Григорьевна",
			LastName:   "Дмитриева",
		},
		{
			FirstName:  "Иван",
			MiddleName: "Иванович",
			LastName:   "Иванов",
		},
		{
			FirstName:  "Василий",
			MiddleName: "Васильевич",
			LastName:   "Васильев",
		},
		{
			FirstName:  "Юрий",
			MiddleName: "Александрович",
			LastName:   "Завьялов",
		},
	}
	for i := range students {
		db.Create(&students[i])
	}

	numbers := []Number{
		{Number: 16061}, // Group не задаю, т.к. группы создаются позже
		{Number: 16062},
		{Number: 16063},
		{Number: 16064},
		{Number: 16065},
		{Number: 16066},
	}
	for i := range numbers {
		db.Create(&numbers[i])
	}

	groups := []Group{
		{
			Name:            "ИВТ-16",
			Status:          "Активная",
			Year:            2016,
			EducationPeriod: 4,
			EducationForm:   "Очная",
			EducationLevel:  "Бакалавриат",
			Institute:       "ИИТИАС",
		},
		{
			Name:            "ИП-16",
			Status:          "Активная",
			Year:            2016,
			EducationPeriod: 4,
			EducationForm:   "Очная",
			EducationLevel:  "Бакалавриат",
			Institute:       "ИИТИАС",
		},
	}
	for i := range groups {
		db.Create(&groups[i])
	}

	// У каждого студента 5 дисциплин, по каждой дисциплине одна оценка (точнее - запись об оценках)
	// в осеннем и весеннем семестрах в 2015 и 2016 году
	// + студентов 6 пока сделаем (выборка репрезентативна)
	scores := []Score{
		{
			Year:                    2016,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "Н/А",
			IntermediateAttestation: "Зачтено",
			Semester:                "Весенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "3",
			AttestationTwo:          "5",
			AttestationThree:        "4",
			IntermediateAttestation: "2",
			Semester:                "Весенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "Н/А",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "Не зачтено",
			Semester:                "Весенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "3",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Весенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Весенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "5",
			AttestationTwo:          "5",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Весенний",
		},

		{
			Year:                    2016,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "Н/А",
			IntermediateAttestation: "Зачтено",
			Semester:                "Осенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "3",
			AttestationTwo:          "5",
			AttestationThree:        "4",
			IntermediateAttestation: "2",
			Semester:                "Осенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "Н/А",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "Не зачтено",
			Semester:                "Осенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "3",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Осенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Осенний",
		},
		{
			Year:                    2016,
			AttestationOne:          "5",
			AttestationTwo:          "5",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Осенний",
		},

		{
			Year:                    2015,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "Н/А",
			IntermediateAttestation: "Зачтено",
			Semester:                "Весенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "3",
			AttestationTwo:          "5",
			AttestationThree:        "4",
			IntermediateAttestation: "2",
			Semester:                "Весенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "Н/А",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "Не зачтено",
			Semester:                "Весенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "3",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Весенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Весенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "5",
			AttestationTwo:          "5",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Весенний",
		},

		{
			Year:                    2015,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "Н/А",
			IntermediateAttestation: "Зачтено",
			Semester:                "Осенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "3",
			AttestationTwo:          "5",
			AttestationThree:        "4",
			IntermediateAttestation: "2",
			Semester:                "Осенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "Н/А",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "Не зачтено",
			Semester:                "Осенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "3",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Осенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "5",
			AttestationTwo:          "4",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Осенний",
		},
		{
			Year:                    2015,
			AttestationOne:          "5",
			AttestationTwo:          "5",
			AttestationThree:        "5",
			IntermediateAttestation: "5",
			Semester:                "Осенний",
		},
	}
	for j := 0; j < 6; j++ {
		for i := range scores {
			db.Create(&scores[i])

			//Create и Save имеют одну общую особенность: они меняют те структуры,
			//	и тот же экземпляр структуры ты либо не сможешь содзать (если используешь Create) либо
			//кт ты им скармливаешь, а именно устанавливают у них ID. Поэтому повторно один
			//обновишь одну и ту же запись (в случае Save). Насчет Update не проверял, но возможно подобное
			//поведение относится и к нему
			scores[i].ID = 0
		}
	}

	cathedras := []Cathedra{
		{Name: "ПИТИП"},
		{Name: "ЭЭИПЭ"},
	}
	for i := range cathedras {
		db.Create(&cathedras[i])
	}

	disciplines := []Discipline{
		// общие для всех групп
		{Name: "Информатика", Cathedra: "ПИТИП"},
		{Name: "Математика", Cathedra: "ПИТИП"},
		{Name: "Программирование", Cathedra: "ПИТИП"},
		{Name: "Электротехника", Cathedra: "ЭЭИПЭ"},
		// только у групп ИВТ
		{Name: "Web-технологии", Cathedra: "ПИТИП"},
		// только у групп ИП
		{Name: "Информационная безопасность", Cathedra: "ПИТИП"},
	}
	for i := range disciplines {
		db.Create(&disciplines[i])
	}

	teachers := []Teacher{
		{
			FirstName:  "Лариса",
			MiddleName: "Дмитриевна",
			LastName:   "Павлова",
		},
		{
			FirstName:  "Ольга",
			MiddleName: "Леонидовна",
			LastName:   "Базайкина",
		},
		{
			FirstName:  "Вадим",
			MiddleName: "Иванович",
			LastName:   "Кожемяченко",
		},
		{
			FirstName:  "Валерий",
			MiddleName: "Семёнович",
			LastName:   "Князев",
		},
		{
			FirstName:  "Максим",
			MiddleName: "Михайлович",
			LastName:   "Гусев",
		},
		{
			FirstName:  "Анна",
			MiddleName: "Валерьевна",
			LastName:   "Корнева",
		},
	}
	for i := range teachers {
		db.Create(&teachers[i])
	}

	// Для одной дисциплины создается по 2 пропуска в каждом семестре 2015 и 2016-го года
	// обучения, дисциплин у каждого 5, всего 6 студентов
	year2015semO, _ := time.Parse("2006-01-02", "2015-09-01")
	year2015semV, _ := time.Parse("2006-01-02", "2016-02-01")
	year2016semO, _ := time.Parse("2006-01-02", "2016-09-01")
	year2016semV, _ := time.Parse("2006-01-02", "2017-02-01")

	skips := []Skip{
		{
			Reason: "У",
			Date:   year2015semO,
		},
		{
			Reason: "Н",
			Date:   year2015semO,
		},

		{
			Reason: "У",
			Date:   year2015semV,
		},
		{
			Reason: "Н",
			Date:   year2015semV,
		},

		{
			Reason: "У",
			Date:   year2016semO,
		},
		{
			Reason: "Н",
			Date:   year2016semO,
		},

		{
			Reason: "У",
			Date:   year2016semV,
		},
		{
			Reason: "Н",
			Date:   year2016semV,
		},
	}
	for i := 0; i < 6; i++ {
		for j := 0; j < 5; j++ {
			for i := range skips {
				db.Create(&skips[i])
				// подобно созданию scores
				skips[i].ID = 0
			}
		}
	}

	// отсюда отошел от принципа "несмешивания создания связей и записей в бд"
	schedules := []Schedule{
		{
			Day:            time.Monday,
			DisciplinesIDs: []int64{1, 2, 3, 4},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
		},
		{
			Day:            time.Tuesday,
			DisciplinesIDs: []int64{1, 1, 3},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
		},
		{
			Day:            time.Wednesday,
			DisciplinesIDs: []int64{4, 3, 2, 1},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
		},
		{
			Day:            time.Thursday,
			DisciplinesIDs: []int64{3, 3, 1, 1},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
		},
		{
			Day:            time.Friday,
			DisciplinesIDs: []int64{1, 3, 1, 3},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
		},

		{
			Day:            time.Monday,
			DisciplinesIDs: []int64{2, 3, 4},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
			IsEvenWeek:     true,
		},
		{
			Day:            time.Tuesday,
			DisciplinesIDs: []int64{3, 1, 1},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
			IsEvenWeek:     true,
		},
		{
			Day:            time.Wednesday,
			DisciplinesIDs: []int64{1, 2, 3, 4},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
			IsEvenWeek:     true,
		},
		{
			Day:            time.Thursday,
			DisciplinesIDs: []int64{1, 1, 3, 3},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
			IsEvenWeek:     true,
		},
		{
			Day:            time.Friday,
			DisciplinesIDs: []int64{3, 1, 3, 1},
			Semester:       "Осенний",
			Year:           2016,
			GroupID:        1,
			IsEvenWeek:     true,
		},
	}
	for i := range schedules {
		db.Create(&schedules[i])
	}

	startdateAutumn2016, _ := time.Parse("2006-01-02", "2016-09-01")
	enddateAutumn2016, _ := time.Parse("2006-01-02", "2016-12-30")
	startdateSpring2016, _ := time.Parse("2006-01-02", "2017-02-01")
	enddateSpring2016, _ := time.Parse("2006-01-02", "2017-05-31")

	startdateAutumn2015, _ := time.Parse("2006-01-02", "2015-09-01")
	enddateAutumn2015, _ := time.Parse("2006-01-02", "2015-12-30")
	startdateSpring2015, _ := time.Parse("2006-01-02", "2016-02-01")
	enddateSpring2015, _ := time.Parse("2006-01-02", "2016-05-31")

	groupinfos := []GroupSemesterInfo{
		{
			Semester:           "Осенний",
			Year:               2016,
			GroupID:            1,
			EducationStartDate: startdateAutumn2016,
			EducationEndDate:   enddateAutumn2016,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 6},
		},
		{
			Semester:           "Осенний",
			Year:               2016,
			GroupID:            2,
			EducationStartDate: startdateAutumn2016,
			EducationEndDate:   enddateAutumn2016,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 5},
		},

		{
			Semester:           "Осенний",
			Year:               2015,
			GroupID:            1,
			EducationStartDate: startdateAutumn2015,
			EducationEndDate:   enddateAutumn2015,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 6},
		},
		{
			Semester:           "Осенний",
			Year:               2015,
			GroupID:            2,
			EducationStartDate: startdateAutumn2015,
			EducationEndDate:   enddateAutumn2015,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 5},
		},

		{
			Semester:           "Весенний",
			Year:               2016,
			GroupID:            1,
			EducationStartDate: startdateSpring2016,
			EducationEndDate:   enddateSpring2016,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 6},
		},
		{
			Semester:           "Весенний",
			Year:               2016,
			GroupID:            2,
			EducationStartDate: startdateSpring2016,
			EducationEndDate:   enddateSpring2016,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 5},
		},

		{
			Semester:           "Весенний",
			Year:               2015,
			GroupID:            1,
			EducationStartDate: startdateSpring2015,
			EducationEndDate:   enddateSpring2015,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 6},
		},
		{
			Semester:           "Весенний",
			Year:               2015,
			GroupID:            2,
			EducationStartDate: startdateSpring2015,
			EducationEndDate:   enddateSpring2015,
			DisciplinesIDs:     []int64{1, 2, 3, 4, 5},
		},
	}
	for i := range groupinfos {
		db.Create(&groupinfos[i])
	}

	// Задание связей

	// User - Students

	var s []Student
	db.Where("last_name in (?)", []string{"Леванков", "Ефимов", "Дмитриева", "Иванов", "Васильев", "Завьялов"}).Find(&s)

	var u []User
	// почему-то юзеры достаются в обратном порядке
	db.Where("last_name in (?)", []string{"Леванков", "Ефимов", "Дмитриева", "Иванов", "Васильев", "Завьялов"}).Order("users.id asc").Find(&u)

	for i := range u {
		u[i].Students = []*Student{&s[i]}
		db.Save(&u[i])
	}

	// Students - Numbers

	// связь с юзером не задается т.к. все связи у юзера уже были выше настроены, также и
	// со всеми структурами будет, т.к. не задавать связи с тем, что было выше

	var n []Number
	db.Where("number in (?)", []int{16061, 16062, 16063, 16064, 16065, 16066}).Find(&n)

	for i := range s {
		s[i].Numbers = []Number{n[i]}
		db.Save(&s[i])
	}

	// Students - Groups

	var g []Group
	db.Where("name in (?)", []string{"ИВТ-16", "ИП-16"}).Find(&g)

	for i := 0; i < 3; i++ {
		s[i].Groups = []*Group{&g[0]}
		db.Save(&s[i])
	}

	for i := 3; i < 6; i++ {
		s[i].Groups = []*Group{&g[1]}
		db.Save(&s[i])
	}

	// Students - Scores

	var sc []Score

	for i := range s {
		s[i].Scores = []Score{}
	}

	for i := 0; i < 16; i++ {
		db.Where("id BETWEEN ? AND ?", 1+6*i, 6+6*i).Find(&sc)

		for j := 0; j < 6; j++ {
			s[j].Scores = append(s[j].Scores, sc[j])
		}
	}

	// некоторые оценки не будут никому отданы
	// веб технологии
	for i := 16; i < 20; i++ {
		db.Where("id BETWEEN ? AND ?", 1+6*i, 6+6*i).Find(&sc)

		for j := 0; j < 3; j++ {
			s[j].Scores = append(s[j].Scores, sc[j])
		}
	}

	// некоторые оценки не будут никому отданы
	// иб
	for i := 20; i < 24; i++ {
		db.Where("id BETWEEN ? AND ?", 1+6*i, 6+6*i).Find(&sc)

		for j := 3; j < 6; j++ {
			s[j].Scores = append(s[j].Scores, sc[j])
		}
	}

	for i := range s {
		db.Save(&s[i])
	}

	// Students - Skips

	var sk []Skip

	// По 8 пропусков (по 2 (ув/неув) * 4 (семестра)  каждому студенту, 40 пропусков на одну дисцилину -
	// значит хватит на 5 студентов. т.к. пропусков всего 240, то дисциплин всего 6, на каждую по 40 пропусков
	//
	// По-хорошему нужно по 48 пропусков на 1-е 4 дисциплины (общие), по 8 пропусков на 6 студентов
	// и по 24 пропуска на оставшиеся 2 дисциплины (3-ем студентам из ИВТ-16 по 8 пропусков по 5-ой дисциплине и
	// 3-ем студентам из ИП-16 по 8 пропусков по 6-ой дисциплине). Можно сделать по 48 пропускам всем дисциплинам, это упростит
	// сопоставление пропусков дисциплинам (упрощение задания связи дисциплина-пропуски), и упростит задание связи пропуски-студент
	// но ценой будет то, что останутся неназначенные никому пропуски.
	//
	// Умышленно раздаю пропуски некорректно, т.к. в ином случае нужно генерировать больше пропусков это раз,
	// устанавливать связь пропуски-дисциплины это два. В результате действий ниже первым 5-и студентам будут разданы
	// оценки по всем дисциплинам (даже по тем, кт в его группе не ведутся (некорректно)), 6-ому ни одного пропуска не достанется.5
	for i := 0; i < 6; i++ {
		db.Where("id BETWEEN ? AND ?", 1+40*i, 40+40*i).Find(&sk)

		for j := 0; j < 5; j++ {
			s[j].Skips = append(s[j].Skips, sk[8*j:8+8*j]...)
			db.Save(&s[j])
		}
	}

	// Student - Group

	for i := 0; i < 3; i++ {
		s[i].Groups = []*Group{&g[0]}
	}
	for i := 3; i < 5; i++ {
		s[i].Groups = []*Group{&g[1]}
	}
	for i := range s {
		db.Save(&s[i])
	}

	// задание старосты

	s[2].GroupID = g[0].ID
	s[5].GroupID = g[1].ID
	db.Save(&s[2]).Save(&s[5])

	// Number - Group - впервые, где я задал связь НЕ со стороны parent
	// структуры (не со стороны структуры c []). Ниже чекнул запрос со стороны
	// parent структуры - робит

	var num []Number
	db.Where("number in (?)", []int{16061, 16062, 16063, 16064, 16065, 16066}).Find(&num)

	for i := 0; i < 3; i++ {
		num[i].Group = "ИВТ-16"
		db.Save(&num[i])
	}
	for i := 3; i < 6; i++ {
		num[i].Group = "ИП-16"
		db.Save(&num[i])
	}

	// Group - Teacher

	var t []Teacher
	db.Where("last_name in (?)", []string{"Павлова", "Базайкина", "Кожемяченко", "Князев", "Корнева", "Гусев"}).Find(&t)

	g[0].Teachers = []*Teacher{&t[0], &t[1], &t[2], &t[3], &t[5]}
	g[1].Teachers = []*Teacher{&t[0], &t[1], &t[2], &t[3], &t[4]}
	for i := range g {
		db.Save(&g[i])
	}

	// Group - Disciplines

	var d []Discipline
	db.Where("name in (?)", []string{"Информатика", "Математика", "Программирование", "Электротехника", "Информационная безопасность", "Web-технологии"}).Find(&d)
	g[0].Disciplines = []*Discipline{&d[0], &d[1], &d[2], &d[3], &d[5]}
	g[1].Disciplines = []*Discipline{&d[0], &d[1], &d[2], &d[3], &d[4]}
	for i := range g {
		db.Save(&g[i])
	}

	// Score - Discipline
	// Всего оценок 120, 6 дисциплин, следовательно по 20 оценок каждой дисциплине

	for i := range d {
		db.Where("id BETWEEN ? AND ?", 1+20*i, 20+20*i).Find(&sc)
		d[i].Scores = sc
		db.Save(&d[i])
	}

	// Discipline - Cathedra

	d[0].Cathedra = "ПИТИП"
	d[1].Cathedra = "ПИТИП"
	d[2].Cathedra = "ПИТИП"
	d[3].Cathedra = "ЭЭИПЭ"
	d[4].Cathedra = "ПИТИП"
	d[5].Cathedra = "ПИТИП"

	// Discipline - Teacher

	for i := range d {
		d[i].Teachers = []*Teacher{&t[i]}
		db.Save(&d[i])
	}

	// Discipline - Skip
	// Дисциплин всего 6, пропусков 240, следовательно, по 40 пропусков каждой дисциплине
	for i := range d {
		db.Where("id BETWEEN ? AND ?", 1+40*i, 40+40*i).Find(&sk)
		d[i].Skips = sk
		db.Save(&d[i])
	}

	// Cathedra - Teacher

	var ca []Cathedra
	db.Where("name in (?)", []string{"ПИТИП", "ЭЭИПЭ"}).Find(&ca)
	ca[0].Teachers = []*Teacher{&t[0], &t[1], &t[2], &t[4], &t[5]}
	ca[1].Teachers = []*Teacher{&t[3]}
	db.Save(&ca[0]).Save(&ca[1])
}
