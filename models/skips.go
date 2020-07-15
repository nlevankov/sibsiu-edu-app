package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

type FilterSkipsQueryParams struct {
	GroupID  *uint   `schema:"group_id"`
	Year     *int    `schema:"year"`
	Semester *string `schema:"semester"`
}

type SkipsResult struct {
	StudentSkipsResult
	StarostaSkipsResult
}

type StudentSkipsResult struct {
	StudentSkipsFilterData
	StudentSkipsResultData
}

type StarostaSkipsResult struct {
	StarostaSkipsFilterData
	StarostaSkipsResultData
}

type StudentSkipsResultData struct {
	Rows []StudentSkipsRow
}

type StarostaSkipsResultData struct {
	Rows []StarostaSkipsRow
}

type StudentSkipsRow struct {
	DisciplineName string
	Skips          []SkipRow
	NumberOfUHours float32
	NumberOfNHours float32
	TotalHours     float32
}

type StarostaSkipsRow struct {
	DisciplineName string
	Skips          []SkipRow
	NumberOfUHours float32
	NumberOfNHours float32
	TotalHours     float32
}

type SkipRow struct {
	SkipDate   string
	SkipReason string
}

type StudentSkipsFilterData struct {
	Groups    []SelectOptionData
	Years     []SelectOptionData
	Semesters []SelectOptionData
}

type StarostaSkipsFilterData struct {
	Groups    []SelectOptionData
	Years     []SelectOptionData
	Semesters []SelectOptionData
	Students  []SelectOptionData
}

type StarostaSkipsEditResult struct {
	StarostaSkipsEditFilterData
	StarostaSkipsEditResultData
}

type StarostaSkipsEditFilterData struct {
	Groups      []SelectOptionData
	Date        string
	Disciplines []SelectOptionData
}

type StarostaSkipsEditResultData struct {
	Rows []StarostaSkipsEditRow
}

type StarostaSkipsEditRow struct {
	StudentName string
	SkipID      uint
	SkipReason  string
}

type SkipsEditStarostaQueryParams struct {
	GroupID        *uint   `schema:"group_id"`
	Date           *string `schema:"date"`
	DisciplineInfo *string `schema:"discipline_info"`
}

type SkipsStudentQueryParams struct {
	GroupID  *uint   `schema:"group_id"`
	Year     *uint   `schema:"year"`
	Semester *string `schema:"semester"`
}

type SkipsStarostaQueryParams struct {
	GroupID   *uint   `schema:"group_id"`
	Year      *uint   `schema:"year"`
	Semester  *string `schema:"semester"`
	StudentID *uint   `schema:"student_id"`
}

type SkipService interface {
	SkipDB
}

type SkipDB interface {
	ByUser(user *User, GETParams interface{}) (*SkipsResult, []string, error)
	Update(skip *Skip) error
	Edit(user *User, GETParams *SkipsEditStarostaQueryParams) (*StarostaSkipsEditResult, error)
	ByUserFilterData(user *User, GETParams interface{}) (interface{}, error)
	ByUserEditFilterData(user *User, GETParams *SkipsEditStarostaQueryParams) (*StarostaSkipsEditResult, error)
}

var _ SkipService = &skipService{}

type skipService struct {
	SkipDB
}

func NewSkipService(db *gorm.DB) SkipService {
	sg := &skipGorm{db}

	return &skipService{
		SkipDB: sg,
	}
}

var _ SkipDB = &skipGorm{}

type skipGorm struct {
	db *gorm.DB
}

func (sg *skipGorm) ByUser(user *User, GETParams interface{}) (*SkipsResult, []string, error) {

	var studentIDs []uint
	var Student Student

	sg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	sg.db.Where("id = ?", studentIDs[0]).Preload("Groups.Starosta").First(&Student)

	// todo чую что не совсем хорошо написано
	//  Вообще я хотел бы получить доступ к полю GroupID вне зависимости от того,
	//  какая структура нам пришла (более "полиморфное" решение хочется)
	switch gParams := GETParams.(type) {

	case *SkipsStudentQueryParams:

		if gParams.GroupID == nil {
			if Student.Groups[0].Starosta.ID == Student.ID {
				result, err := sg.skipsStarosta(&Student, &SkipsStarostaQueryParams{GroupID: gParams.GroupID})
				return result, []string{"skips/starosta.js"}, err
			}

			result, err := sg.skipsStudent(&Student, gParams)
			return result, []string{"skips/student.js"}, err

		} else {

			// возможно неоптимально
			for _, v := range Student.Groups {
				if v.ID == *gParams.GroupID {
					// выполняется ли хоть раз это условие? - да (случай с только параметром - группой от старосты)
					if v.Starosta.ID == Student.ID {
						result, err := sg.skipsStarosta(&Student, &SkipsStarostaQueryParams{GroupID: gParams.GroupID})
						return result, []string{"skips/starosta.js"}, err
					}
				}
			}

			result, err := sg.skipsStudent(&Student, gParams)
			return result, []string{"skips/student.js"}, err
		}

	case *SkipsStarostaQueryParams:

		// рассмотрение условия if gParams.GroupID == nil здесь не нужно, т.к. парсится будет без параметров
		// в любом случае в структуру студента, это выше учтено

		result, err := sg.skipsStarosta(&Student, gParams)
		return result, []string{"skips/starosta.js"}, err
	}

	// todo возвращать собственую ошибку по типу ErrЮзерСНедопустимымКлассом
	return nil, nil, nil
}

func (sg *skipGorm) Edit(user *User, GETParams *SkipsEditStarostaQueryParams) (*StarostaSkipsEditResult, error) {
	// todo про обработку ошибок и валидацию не забудь

	// придти в GETParams может ничего, либо всё, т.к. подгружатор будет следить за этим

	var studentIDs []uint
	var Starosta Student

	sg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	sg.db.Where("id = ?", studentIDs[0]).Preload("Groups").First(&Starosta)

	var groups []SelectOptionData
	var targetGroupID uint

	if GETParams.GroupID != nil {
		targetGroupID = *GETParams.GroupID
	} else {
		for _, v := range Starosta.Groups {
			if v.Starosta.ID == Starosta.ID {
				targetGroupID = v.ID
				break
			}
		}
	}

	for _, v := range Starosta.Groups {
		if v.ID == targetGroupID {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID, IsSelected: true})
		} else {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID})
		}
	}

	var date time.Time
	if GETParams.Date != nil {
		date, _ = time.Parse("2006-01-02", *GETParams.Date)
	} else {
		date = time.Now()
	}

	// тут возможно рассогласование с датами начала и конца в структуре GroupSemesterInfo ниже
	var semester string
	if date.Month() == time.January || date.Month() >= time.September && date.Month() <= time.December {
		semester = "Осенний"
	} else {
		semester = "Весенний"
	}

	var year int
	if semester == "Осенний" {
		year = date.Year()
	} else {
		year = date.Year() - 1
	}

	groupInfo := []GroupSemesterInfo{}
	sg.db.
		Where(`group_semester_infos.semester = ? AND group_semester_infos.year = ? AND group_semester_infos.group_id = ?`, semester, year, targetGroupID).
		Find(&groupInfo)

	var r StarostaSkipsEditResult
	if len(groupInfo) == 0 {
		r = StarostaSkipsEditResult{
			StarostaSkipsEditFilterData{
				Groups:      groups,
				Date:        date.Format("2006-01-02"),
				Disciplines: nil,
			},
			StarostaSkipsEditResultData{},
		}

		return &r, nil
	}

	if date.After(groupInfo[0].EducationEndDate) && date.Before(groupInfo[0].EducationStartDate) {
		r = StarostaSkipsEditResult{
			StarostaSkipsEditFilterData{
				Groups:      groups,
				Date:        date.Format("2006-01-02"),
				Disciplines: nil,
			},
			StarostaSkipsEditResultData{},
		}

		return &r, nil
	}

	_, currentWeekNumber := date.ISOWeek()
	_, startWeekNumber := groupInfo[0].EducationStartDate.ISOWeek()
	isEven := (currentWeekNumber-startWeekNumber+1)%2 == 0

	schelude := []Schedule{}
	sg.db.
		Where(`schedules.day = ? AND schedules.semester = ? AND schedules.year = ? AND schedules.group_id = ? AND schedules.is_even_week = ?`, date.Weekday(), semester, year, targetGroupID, isEven).
		Find(&schelude)

	if len(schelude) == 0 {
		r = StarostaSkipsEditResult{
			StarostaSkipsEditFilterData{
				Groups:      groups,
				Date:        date.Format("2006-01-02"),
				Disciplines: nil,
			},
			StarostaSkipsEditResultData{},
		}

		return &r, nil
	}

	if GETParams.DisciplineInfo != nil {
		if *GETParams.DisciplineInfo == "Нет данных" {
			r = StarostaSkipsEditResult{
				StarostaSkipsEditFilterData{
					Groups:      groups,
					Date:        date.Format("2006-01-02"),
					Disciplines: nil,
				},
				StarostaSkipsEditResultData{},
			}

			return &r, nil
		}
	}

	var Disciplines []SelectOptionData

	counters := map[int64]int{}
	counters2 := map[int64]int{}
	for _, v := range schelude[0].DisciplinesIDs {
		counters[v]++
	}

	for _, v := range schelude[0].DisciplinesIDs {
		d := Discipline{}
		sg.db.Where("disciplines.id = ?", v).First(&d)
		if counters[v] > 1 {
			counters2[v]++
			Disciplines = append(Disciplines, SelectOptionData{Text: d.Name + " " + strconv.Itoa(counters2[v]), Value: strconv.Itoa(int(v)) + "_" + strconv.Itoa(counters2[v])})
		} else {
			Disciplines = append(Disciplines, SelectOptionData{Text: d.Name, Value: strconv.Itoa(int(v)) + "_1"})
		}
	}

	var targetDisciplineInfo string
	if GETParams.DisciplineInfo != nil {
		targetDisciplineInfo = *GETParams.DisciplineInfo
	} else {
		targetDisciplineInfo = Disciplines[0].Value.(string)
	}

	for i := range Disciplines {
		if Disciplines[i].Value.(string) == targetDisciplineInfo {
			Disciplines[i].IsSelected = true
			break
		}
	}

	// формирование результата

	var group Group

	sg.db.
		Where("id = ?", targetGroupID).
		Preload("Students", func(db *gorm.DB) *gorm.DB { return db.Order("students.last_name") }).
		Find(&group)

	var starostaSkipsEditResultData StarostaSkipsEditResultData
	id, lessonNumber := parseDisciplineInfo(targetDisciplineInfo)

	for _, student := range group.Students {
		studentInfo := StarostaSkipsEditRow{}
		skip := Skip{}

		studentInfo.StudentName = student.LastName + " " + student.FirstName + " " + student.MiddleName

		sg.db.
			Where(`skips.student_id = ? AND skips.date = ? AND skips.discipline_id = ? AND skips.lesson_number = ?`, student.ID, date, id, lessonNumber-1).
			FirstOrCreate(&skip, Skip{StudentID: student.ID, DisciplineID: uint(id), Reason: "", Date: date, LessonNumber: lessonNumber - 1})

		studentInfo.SkipID = skip.ID
		studentInfo.SkipReason = skip.Reason

		starostaSkipsEditResultData.Rows = append(starostaSkipsEditResultData.Rows, studentInfo)
	}

	r = StarostaSkipsEditResult{
		StarostaSkipsEditFilterData{
			Groups:      groups,
			Date:        date.Format("2006-01-02"),
			Disciplines: Disciplines,
		},
		starostaSkipsEditResultData,
	}

	return &r, nil

}

func parseDisciplineInfo(info string) (int, int) {
	split := strings.Split(info, "_")
	id, _ := strconv.Atoi(split[0])
	lessonNumber, _ := strconv.Atoi(split[1])

	return id, lessonNumber
}

func (sg *skipGorm) skipsStarosta(Starosta *Student, GETParams *SkipsStarostaQueryParams) (*SkipsResult, error) {

	var groups []SelectOptionData
	var targetGroupID uint

	if GETParams.GroupID != nil {
		targetGroupID = *GETParams.GroupID
	} else {
		for _, v := range Starosta.Groups {
			if v.Starosta.ID == Starosta.ID {
				targetGroupID = v.ID
				break
			}
		}
	}

	for _, v := range Starosta.Groups {
		if v.ID == targetGroupID {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID, IsSelected: true})
		} else {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID})
		}
	}

	var group Group
	sg.db.
		Where("id = ?", targetGroupID).
		Preload("Students", func(db *gorm.DB) *gorm.DB { return db.Order("students.last_name") }).
		Find(&group)

	var students []SelectOptionData
	var targetStudentID uint

	if GETParams.StudentID != nil {
		targetStudentID = *GETParams.StudentID
	} else {
		targetStudentID = Starosta.ID
	}

	for _, v := range group.Students {
		if v.ID == targetStudentID {
			students = append(students, SelectOptionData{Text: v.LastName + " " + v.FirstName + " " + v.MiddleName, Value: v.ID, IsSelected: true})
		} else {
			students = append(students, SelectOptionData{Text: v.LastName + " " + v.FirstName + " " + v.MiddleName, Value: v.ID})
		}
	}

	var years []SelectOptionData
	var targetYear uint

	// todo не могу понять, почему уникальные не получаются + возможно неоптимально
	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?)", targetStudentID, []string{""}).
		Order("date_part('year', skips.date) desc").
		Select("DISTINCT date_part('year', skips.date), CASE WHEN (date_part('month', skips.date)) IN (2,3,4,5,6,7,8) THEN date_part('year', skips.date) - 1 ELSE date_part('year', skips.date) END AS value,  CASE WHEN (date_part('month', skips.date)) IN (2,3,4,5,6,7,8) THEN concat(date_part('year', skips.date) - 1, ' - ', date_part('year', skips.date)) ELSE concat(date_part('year', skips.date), ' - ', date_part('year', skips.date) + 1) END AS text").
		Scan(&years)

	// костыль
	seen := make(map[string]struct{}, len(years))
	j := 0
	for _, v := range years {
		if _, ok := seen[v.Text]; ok {
			continue
		}
		seen[v.Text] = struct{}{}
		years[j] = v
		j++
	}
	years = years[:j]

	if GETParams.Year != nil {
		targetYear = *GETParams.Year
	} else {
		targetYear = uint(years[0].Value.(float64))
	}

	for i := range years {
		if uint(years[i].Value.(float64)) == targetYear {
			years[i].IsSelected = true
			break
		}
	}

	var semesters []SelectOptionData
	var targetSemester string
	var skip Skip

	autumnSemesterStart := time.Date(int(targetYear), time.September, 1, 0, 0, 0, 0, time.UTC)
	autumnSemesterEnd := time.Date(int(targetYear)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
	springSemesterStart := time.Date(int(targetYear)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
	springSemesterEnd := time.Date(int(targetYear)+1, time.June, 31, 23, 59, 59, 59, time.UTC)

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", targetStudentID, []string{""}, autumnSemesterStart, autumnSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Осенний"})
	}

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", targetStudentID, []string{""}, springSemesterStart, springSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Весенний"})
	}

	if GETParams.Semester != nil {
		targetSemester = *GETParams.Semester
	} else {
		targetSemester = semesters[0].Text
	}

	for i := range semesters {
		if semesters[i].Text == targetSemester {
			semesters[i].IsSelected = true
			break
		}
	}

	groupInfo := GroupSemesterInfo{}
	sg.db.
		Where(`group_semester_infos.semester = ? AND group_semester_infos.year = ? AND group_semester_infos.group_id = ?`, targetSemester, targetYear, targetGroupID).
		First(&groupInfo)

	var rows []StarostaSkipsRow

	for _, v := range groupInfo.DisciplinesIDs {
		row := StarostaSkipsRow{}

		var skips []SkipRow
		var from time.Time
		var to time.Time

		if targetSemester == "Осенний" {
			// todo не уверен насчет utc (как он отличается от gmt + 7)
			//  + к тому, что всё ок здесь из документации к пакету time:
			//  "in the appropriate zone for that time in the given location."
			from = time.Date(int(targetYear), time.September, 1, 0, 0, 0, 0, time.UTC)
			to = time.Date(int(targetYear)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
		} else {
			from = time.Date(int(targetYear)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
			to = time.Date(int(targetYear)+1, time.June, 31, 23, 59, 59, 59, time.UTC)
		}

		sg.db.
			Model(Skip{}).
			Where("skips.discipline_id = ? AND skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", v, targetStudentID, []string{""}, from, to).
			Order("skips.date").
			Select("to_char(skips.date, 'DD.MM.YYYY') as skip_date, skips.reason as skip_reason").
			Scan(&skips)

		if len(skips) == 0 {
			continue
		}

		sg.db.
			Model(Discipline{}).
			Where(`disciplines.id = ?`, v).
			Select("disciplines.name as discipline_name").
			Scan(&row)

		for _, v := range skips {
			if v.SkipReason == "У" {
				row.NumberOfUHours += 2
			} else {
				row.NumberOfNHours += 2
			}
			row.TotalHours += 2
		}

		row.Skips = skips

		rows = append(rows, row)
	}

	r := SkipsResult{
		StudentSkipsResult{},
		StarostaSkipsResult{
			StarostaSkipsFilterData{
				Groups:    groups,
				Years:     years,
				Semesters: semesters,
				Students:  students,
			},
			StarostaSkipsResultData{
				rows,
			},
		},
	}

	return &r, nil
}

func (sg *skipGorm) skipsStudent(Student *Student, GETParams *SkipsStudentQueryParams) (*SkipsResult, error) {
	// todo про обработку ошибок и валидацию не забудь

	var groups []SelectOptionData
	var targetGroupID uint

	if GETParams.GroupID != nil {
		targetGroupID = *GETParams.GroupID
	} else {
		targetGroupID = Student.Groups[0].ID
	}

	for _, v := range Student.Groups {
		if v.ID == targetGroupID {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID, IsSelected: true})
		} else {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID})
		}
	}

	var years []SelectOptionData
	var targetYear uint

	// todo не могу понять, почему уникальные не получаются + возможно неоптимально
	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?)", Student.ID, []string{""}).
		Order("date_part('year', skips.date) desc").
		Select("DISTINCT date_part('year', skips.date), CASE WHEN (date_part('month', skips.date)) IN (2,3,4,5,6,7,8) THEN date_part('year', skips.date) - 1 ELSE date_part('year', skips.date) END AS value,  CASE WHEN (date_part('month', skips.date)) IN (2,3,4,5,6,7,8) THEN concat(date_part('year', skips.date) - 1, ' - ', date_part('year', skips.date)) ELSE concat(date_part('year', skips.date), ' - ', date_part('year', skips.date) + 1) END AS text").
		Scan(&years)

	// костыль
	seen := make(map[string]struct{}, len(years))
	j := 0
	for _, v := range years {
		if _, ok := seen[v.Text]; ok {
			continue
		}
		seen[v.Text] = struct{}{}
		years[j] = v
		j++
	}
	years = years[:j]

	if GETParams.Year != nil {
		targetYear = *GETParams.Year
	} else {
		targetYear = uint(years[0].Value.(float64))
	}

	for i := range years {
		if uint(years[i].Value.(float64)) == targetYear {
			years[i].IsSelected = true
			break
		}
	}

	var semesters []SelectOptionData
	var targetSemester string
	var skip Skip

	autumnSemesterStart := time.Date(int(targetYear), time.September, 1, 0, 0, 0, 0, time.UTC)
	autumnSemesterEnd := time.Date(int(targetYear)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
	springSemesterStart := time.Date(int(targetYear)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
	springSemesterEnd := time.Date(int(targetYear)+1, time.June, 31, 23, 59, 59, 59, time.UTC)

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", Student.ID, []string{""}, autumnSemesterStart, autumnSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Осенний"})
	}

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", Student.ID, []string{""}, springSemesterStart, springSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Весенний"})
	}

	if GETParams.Semester != nil {
		targetSemester = *GETParams.Semester
	} else {
		targetSemester = semesters[0].Text
	}

	for i := range semesters {
		if semesters[i].Text == targetSemester {
			semesters[i].IsSelected = true
			break
		}
	}

	groupInfo := GroupSemesterInfo{}
	sg.db.
		Where(`group_semester_infos.semester = ? AND group_semester_infos.year = ? AND group_semester_infos.group_id = ?`, targetSemester, targetYear, targetGroupID).
		First(&groupInfo)

	var rows []StudentSkipsRow

	for _, v := range groupInfo.DisciplinesIDs {
		row := StudentSkipsRow{}

		var skips []SkipRow
		var from time.Time
		var to time.Time

		if targetSemester == "Осенний" {
			// todo не уверен насчет utc (как он отличается от gmt + 7)
			//  + к тому, что всё ок здесь из документации к пакету time:
			//  "in the appropriate zone for that time in the given location."
			from = time.Date(int(targetYear), time.September, 1, 0, 0, 0, 0, time.UTC)
			to = time.Date(int(targetYear)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
		} else {
			from = time.Date(int(targetYear)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
			to = time.Date(int(targetYear)+1, time.June, 31, 23, 59, 59, 59, time.UTC)
		}

		sg.db.
			Model(Skip{}).
			Where("skips.discipline_id = ? AND skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", v, Student.ID, []string{""}, from, to).
			Order("skips.date").
			Select("to_char(skips.date, 'DD.MM.YYYY') as skip_date, skips.reason as skip_reason").
			Scan(&skips)

		if len(skips) == 0 {
			continue
		}

		sg.db.
			Model(Discipline{}).
			Where(`disciplines.id = ?`, v).
			Select("disciplines.name as discipline_name").
			Scan(&row)

		for _, v := range skips {
			if v.SkipReason == "У" {
				row.NumberOfUHours += 2
			} else {
				row.NumberOfNHours += 2
			}
			row.TotalHours += 2
		}

		row.Skips = skips

		rows = append(rows, row)
	}

	r := SkipsResult{
		StudentSkipsResult{
			StudentSkipsFilterData{
				Groups:    groups,
				Years:     years,
				Semesters: semesters,
			},
			StudentSkipsResultData{
				rows,
			},
		},
		StarostaSkipsResult{},
	}

	return &r, nil
}

func (sg *skipGorm) Update(skip *Skip) error {
	return sg.db.Model(&Skip{}).Update(skip).Error
}

func (sg *skipGorm) ByUserFilterData(user *User, GETParams interface{}) (interface{}, error) {

	var studentIDs []uint
	var Student Student

	sg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	sg.db.Where("id = ?", studentIDs[0]).Preload("Groups.Starosta").First(&Student)

	switch gParams := GETParams.(type) {

	case *SkipsStudentQueryParams:
		// если прилетят group id и year (как минимум), но без student id  - распарится в структуру студента
		result, err := sg.skipsFilterStudent(&Student, gParams)
		return result, err

	case *SkipsStarostaQueryParams:
		// если прилетят как минимум group id и student id (кт обязательно как минимум прилетят) - распарится в структуру старосты
		result, err := sg.skipsFilterStarosta(&Student, gParams)
		return result, err
	}

	// todo возвращать собственую ошибку по типу ErrЮзерСНедопустимымКлассом, хотя это отсекается RequireClassMW
	return nil, nil
}

func (sg *skipGorm) skipsFilterStudent(Student *Student, GETParams *SkipsStudentQueryParams) (*StudentSkipsFilterData, error) {
	// todo обработка ошибок

	var semesters []SelectOptionData
	var skip Skip

	autumnSemesterStart := time.Date(int(*GETParams.Year), time.September, 1, 0, 0, 0, 0, time.UTC)
	autumnSemesterEnd := time.Date(int(*GETParams.Year)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
	springSemesterStart := time.Date(int(*GETParams.Year)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
	springSemesterEnd := time.Date(int(*GETParams.Year)+1, time.June, 31, 23, 59, 59, 59, time.UTC)

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", Student.ID, []string{""}, autumnSemesterStart, autumnSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Осенний"})
	}

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", Student.ID, []string{""}, springSemesterStart, springSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Весенний"})
	}

	var r StudentSkipsFilterData
	if len(semesters) == 0 {
		r = StudentSkipsFilterData{
			Semesters: nil,
		}
		return &r, nil
	}

	r = StudentSkipsFilterData{
		Semesters: semesters,
	}

	return &r, nil
}

func (sg *skipGorm) skipsFilterStarosta(Starosta *Student, GETParams *SkipsStarostaQueryParams) (*StarostaSkipsFilterData, error) {
	// todo обработка ошибок

	var years []SelectOptionData
	var targetYear uint
	var r StarostaSkipsFilterData

	if GETParams.Year == nil {

		// todo не могу понять, почему уникальные не получаются + возможно неоптимально
		sg.db.
			Model(Skip{}).
			Where("skips.student_id = ? AND skips.reason NOT in (?)", *GETParams.StudentID, []string{""}).
			Order("date_part('year', skips.date) desc").
			Select("DISTINCT date_part('year', skips.date), CASE WHEN (date_part('month', skips.date)) IN (2,3,4,5,6,7,8) THEN date_part('year', skips.date) - 1 ELSE date_part('year', skips.date) END AS value,  CASE WHEN (date_part('month', skips.date)) IN (2,3,4,5,6,7,8) THEN concat(date_part('year', skips.date) - 1, ' - ', date_part('year', skips.date)) ELSE concat(date_part('year', skips.date), ' - ', date_part('year', skips.date) + 1) END AS text").
			Scan(&years)

		if len(years) == 0 {
			r = StarostaSkipsFilterData{
				Years:     nil,
				Semesters: nil,
			}
			return &r, nil
		}

		// костыль
		seen := make(map[string]struct{}, len(years))
		j := 0
		for _, v := range years {
			if _, ok := seen[v.Text]; ok {
				continue
			}
			seen[v.Text] = struct{}{}
			years[j] = v
			j++
		}
		years = years[:j]

		targetYear = uint(years[0].Value.(float64))

	} else {
		targetYear = *GETParams.Year
	}

	var semesters []SelectOptionData
	var skip Skip

	autumnSemesterStart := time.Date(int(targetYear), time.September, 1, 0, 0, 0, 0, time.UTC)
	autumnSemesterEnd := time.Date(int(targetYear)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
	springSemesterStart := time.Date(int(targetYear)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
	springSemesterEnd := time.Date(int(targetYear)+1, time.June, 31, 23, 59, 59, 59, time.UTC)

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", *GETParams.StudentID, []string{""}, autumnSemesterStart, autumnSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Осенний"})
	}

	sg.db.
		Model(Skip{}).
		Where("skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ?", *GETParams.StudentID, []string{""}, springSemesterStart, springSemesterEnd).
		First(&skip)

	if skip != (Skip{}) {
		semesters = append(semesters, SelectOptionData{Text: "Весенний"})
	}

	if len(semesters) == 0 {

		r = StarostaSkipsFilterData{
			Years:     years,
			Semesters: nil,
		}

		return &r, nil
	}

	r = StarostaSkipsFilterData{
		Years:     years,
		Semesters: semesters,
	}

	return &r, nil

}

func (sg *skipGorm) ByUserEditFilterData(user *User, GETParams *SkipsEditStarostaQueryParams) (*StarostaSkipsEditResult, error) {
	// todo про обработку ошибок и валидацию не забудь

	var studentIDs []uint
	var Starosta Student

	sg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	sg.db.Where("id = ?", studentIDs[0]).Preload("Groups").First(&Starosta)

	targetGroupID := *GETParams.GroupID
	date, _ := time.Parse("2006-01-02", *GETParams.Date)

	// тут возможно рассогласование с датами начала и конца в структуре GroupSemesterInfo ниже
	var semester string
	if date.Month() == time.January || date.Month() >= time.September && date.Month() <= time.December {
		semester = "Осенний"
	} else {
		semester = "Весенний"
	}

	var year int
	if semester == "Осенний" {
		year = date.Year()
	} else {
		year = date.Year() - 1
	}

	groupInfo := []GroupSemesterInfo{}
	sg.db.
		Where(`group_semester_infos.semester = ? AND group_semester_infos.year = ? AND group_semester_infos.group_id = ?`, semester, year, targetGroupID).
		Find(&groupInfo)

	if len(groupInfo) == 0 {
		return nil, nil
	}

	if date.After(groupInfo[0].EducationEndDate) && date.Before(groupInfo[0].EducationStartDate) {
		return nil, nil
	}

	_, currentWeekNumber := date.ISOWeek()
	_, startWeekNumber := groupInfo[0].EducationStartDate.ISOWeek()
	isEven := (currentWeekNumber-startWeekNumber+1)%2 == 0

	schelude := []Schedule{}
	sg.db.
		Where(`schedules.day = ? AND schedules.semester = ? AND schedules.year = ? AND schedules.group_id = ? AND schedules.is_even_week = ?`, date.Weekday(), semester, year, targetGroupID, isEven).
		Find(&schelude)

	if len(schelude) == 0 {
		return nil, nil
	}

	var Disciplines []SelectOptionData

	counters := map[int64]int{}
	counters2 := map[int64]int{}
	for _, v := range schelude[0].DisciplinesIDs {
		counters[v]++
	}

	for _, v := range schelude[0].DisciplinesIDs {
		d := Discipline{}
		sg.db.Where("disciplines.id = ?", v).First(&d)
		if counters[v] > 1 {
			counters2[v]++
			Disciplines = append(Disciplines, SelectOptionData{Text: d.Name + " " + strconv.Itoa(counters2[v]), Value: strconv.Itoa(int(v)) + "_" + strconv.Itoa(counters2[v])})
		} else {
			Disciplines = append(Disciplines, SelectOptionData{Text: d.Name, Value: strconv.Itoa(int(v)) + "_1"})
		}
	}

	var r StarostaSkipsEditResult
	r = StarostaSkipsEditResult{
		StarostaSkipsEditFilterData{
			Groups:      nil,
			Date:        "",
			Disciplines: Disciplines,
		},
		StarostaSkipsEditResultData{},
	}

	return &r, nil
}

type Skip struct {
	gorm.Model          // дата содержится в Created_at - отменено использование этого поля в кач-ве даты, выделено отдельное
	Reason       string `sql:"type:reason"`
	DisciplineID uint
	StudentID    uint
	Date         time.Time // нужна потому, что в методе skipsC.Show отсутсвующие пропуски за конкретную дату создаются на ходу
	LessonNumber int
}
