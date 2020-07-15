package models

import (
	"database/sql"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"strconv"
	"time"
)

// Указатели используются потому, что иначе будет инициализация 0-ми значениями
type FilterScoresQueryParams struct {
	GroupID  *uint   `schema:"group_id"`
	Year     *int    `schema:"year"`
	Semester *string `schema:"semester"`
}

type ScoresStudentQueryParams struct {
	GroupID        *uint   `schema:"group_id"`
	Year           *uint   `schema:"year"`
	Semester       *string `schema:"semester"`
	DisciplinesIDs []uint  `schema:"disciplines_ids"`
}

type ScoresStarostaQueryParams struct {
	GroupID      *uint   `schema:"group_id"`
	Year         *uint   `schema:"year"`
	Semester     *string `schema:"semester"`
	DisciplineID *uint   `schema:"discipline_id"`
}

// Структура нужна, дабы не менять всякий раз сигнатуру методу
type ScoresResult struct {
	StudentScoresResult
	StarostaScoresResult
}

// Просто назвать её Result и поля filterData и resultdata, а в структуре выше сделать поле Student с типом Result
type StudentScoresResult struct {
	StudentScoresFilterData
	StudentScoresResultData
}

type StarostaScoresResult struct {
	StarostaScoresFilterData
	StarostaScoresResultData
}

type StudentScoresFilterData struct {
	Groups      []SelectOptionData
	Years       []SelectOptionData
	Semesters   []SelectOptionData
	Disciplines []SelectOptionData
	Teachers    []SelectOptionData
}

type StarostaScoresFilterData struct {
	Groups      []SelectOptionData
	Years       []SelectOptionData
	Semesters   []SelectOptionData
	Disciplines []SelectOptionData
}

type SelectOptionData struct {
	Value      interface{}
	Text       string
	IsSelected bool
}

type StudentScoresResultData struct {
	Result []StudentScoresRows
}

type StarostaScoresResultData struct {
	Result StarostaScoresRows
}

type StudentScoresRows struct {
	DisciplineName  string
	TeacherName     string
	Cathedra        string
	AttOne          string
	AttTwo          string
	AttThree        string
	AttIntermediate string
	Skips           string
}

type StarostaScoresRows struct {
	DisciplineID uint   //исп-ся в методе ReportByStarosta - больше так не делать, отдельную структуру на каждый метод
	Discipline   string //исп-ся в методе ReportByStarosta
	Teacher      string
	Cathedra     string
	StudentsInfo []StarostaScoresStudentInfo
}

type StarostaScoresStudentInfo struct {
	ScoreID         uint
	StudentName     string
	AttOne          string
	AttTwo          string
	AttThree        string
	AttIntermediate string
	Skips           string
}

type ScoreService interface {
	ScoreDB
}

type ScoreDB interface {
	LastUpdated() (string, error)
	ByUser(user *User, GETParams interface{}) (*ScoresResult, []string, error)
	ByUserFilterData(user *User, GETParams *FilterScoresQueryParams) (*ScoresResult, error)
	Update(score *Score) error
	ReportByUser(user *User, GETParams *ScoresStarostaQueryParams) (*excelize.File, error)
}

var _ ScoreService = &scoreService{}

type scoreService struct {
	ScoreDB
}

func NewScoreService(db *gorm.DB) ScoreService {
	sg := &scoreGorm{db}

	return &scoreService{
		ScoreDB: sg,
	}
}

var _ ScoreDB = &scoreGorm{}

type scoreGorm struct {
	db *gorm.DB
}

func (sg *scoreGorm) ByUserFilterData(user *User, GETParams *FilterScoresQueryParams) (*ScoresResult, error) {

	var studentIDs []uint
	var Student Student

	sg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	sg.db.
		Where("id = ?", studentIDs[0]).
		Preload("Groups.Starosta").
		First(&Student)

	// возможно неоптимально
	for _, v := range Student.Groups {
		if v.ID == *GETParams.GroupID {
			if v.Starosta.ID == Student.ID {
				return sg.scoresFilterStarosta(&Student, GETParams)
			}
		}
	}

	return sg.scoresFilterStudent(&Student, GETParams)

	// todo возвращать собственую ошибку по типу ErrЮзерСНедопустимымКлассом, хотя это отсекается RequireClassMW
	return nil, nil
}

func (sg *scoreGorm) ByUser(user *User, GETParams interface{}) (*ScoresResult, []string, error) {

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

	case *ScoresStudentQueryParams:

		if gParams.GroupID == nil {
			if Student.Groups[0].Starosta.ID == Student.ID {
				result, err := sg.scoresStarosta(&Student, &ScoresStarostaQueryParams{GroupID: gParams.GroupID})
				return result, []string{"scores/starosta.js"}, err
			}

			result, err := sg.scoresStudent(&Student, gParams)
			return result, []string{"scores/student.js"}, err

		} else {

			// возможно неоптимально
			for _, v := range Student.Groups {
				if v.ID == *gParams.GroupID {
					// выполняется ли хоть раз это условие? - да (случай с только параметром - группой от старосты)
					if v.Starosta.ID == Student.ID {
						result, err := sg.scoresStarosta(&Student, &ScoresStarostaQueryParams{GroupID: gParams.GroupID})
						return result, []string{"scores/starosta.js"}, err
					}
				}
			}

			result, err := sg.scoresStudent(&Student, gParams)
			return result, []string{"scores/student.js"}, err
		}

	case *ScoresStarostaQueryParams:

		// рассмотрение условия if gParams.GroupID == nil здесь не нужно, т.к. парсится будет без параметров
		// в любом случае в структуру студента, это выше учтено

		result, err := sg.scoresStarosta(&Student, gParams)
		return result, []string{"scores/starosta.js"}, err
	}

	// todo возвращать собственую ошибку по типу ErrЮзерСНедопустимымКлассом
	return nil, nil, nil
}

func (sg *scoreGorm) scoresStudent(Student *Student, GETParams *ScoresStudentQueryParams) (*ScoresResult, error) {
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

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Where("scores.student_id = ? AND groups.id = ?", Student.ID, targetGroupID).
		Order("scores.year desc").
		Select("DISTINCT scores.year as value, concat(scores.year, ' ', '-', ' ', scores.year + 1) as text").
		Scan(&years)

	if GETParams.Year != nil {
		targetYear = *GETParams.Year
	} else {
		targetYear = uint(years[0].Value.(int64))
	}

	for i := range years {
		if uint(years[i].Value.(int64)) == targetYear {
			years[i].IsSelected = true
			break
		}
	}

	var semesters []SelectOptionData
	var targetSemester string

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Where("scores.student_id = ? AND groups.id = ? AND scores.year = ?", Student.ID, targetGroupID, targetYear).
		Order("scores.semester desc").
		Select("DISTINCT scores.semester as text").
		Scan(&semesters)

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

	var disciplines []SelectOptionData
	var targetDisciplinesIDs []uint

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND groups_teachers.group_id = ?`, Student.ID, targetGroupID, targetYear, targetSemester, targetGroupID).
		Select("DISTINCT disciplines.id as value, disciplines.name as text").
		Scan(&disciplines)

	if GETParams.DisciplinesIDs != nil {
		for i := range GETParams.DisciplinesIDs {
			targetDisciplinesIDs = append(targetDisciplinesIDs, GETParams.DisciplinesIDs[i])
		}
	} else {
		for i := range disciplines {
			targetDisciplinesIDs = append(targetDisciplinesIDs, uint(disciplines[i].Value.(int64)))
		}
	}

	for i := range disciplines {
		// не оптимально
		// можно переписать так, чтобы цикл каждый раз начинался не с начала, а со следующей опцией за совпавшей
		// в предыдущей проверке
		for _, vv := range targetDisciplinesIDs {
			if uint(disciplines[i].Value.(int64)) == vv {
				disciplines[i].IsSelected = true
			}
		}
	}

	var result []StudentScoresRows
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
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN (SELECT skips.discipline_id, COUNT(*)*2 hours FROM skips WHERE skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ? GROUP BY skips.discipline_id) foo ON foo.discipline_id = disciplines.id", Student.ID, []string{""}, from, to).
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND groups_teachers.group_id = ? AND disciplines.id in (?)`, Student.ID, targetGroupID, targetYear, targetSemester, targetGroupID, targetDisciplinesIDs).
		Select("disciplines.name as discipline_name, concat(teachers.last_name, ' ', teachers.first_name, ' ', teachers.middle_name) as teacher_name, disciplines.cathedra as cathedra, scores.attestation_one as att_one, scores.attestation_two as att_two, scores.attestation_three as att_three, scores.intermediate_attestation as att_intermediate, foo.hours as skips").
		Scan(&result)

	r := ScoresResult{
		StudentScoresResult{
			StudentScoresFilterData{
				Groups:      groups,
				Years:       years,
				Semesters:   semesters,
				Disciplines: disciplines,
			},
			StudentScoresResultData{
				Result: result,
			},
		},
		StarostaScoresResult{},
	}

	return &r, nil
}

// Возвращает время последнего изменения записей в таблице Scores в формате 2006-01-02
// В случае любых ошибок возвращает "", err, где err - ошибка
// Если в таблице Scores нет записей, то err = ErrNotFound
// Если возникла любая другая ошибка, то err != ErrNotFound и будет возвращена непосредственно
func (sg *scoreGorm) LastUpdated() (string, error) {
	var result time.Time

	err := sg.db.Table("scores").Select("min(updated_at)").Row().Scan(&result)

	switch err {
	case nil:
		return result.Format("2006-01-02"), nil
	case sql.ErrNoRows:
		return "", ErrNotFound
	default:
		return "", err
	}
}

func (sg *scoreGorm) scoresStarosta(Starosta *Student, GETParams *ScoresStarostaQueryParams) (*ScoresResult, error) {
	// todo про обработку ошибок и валидацию не забудь

	var groups []SelectOptionData
	var targetGroupID uint

	if GETParams.GroupID != nil {
		targetGroupID = *GETParams.GroupID
	} else {
		targetGroupID = Starosta.Groups[0].ID
	}

	for _, v := range Starosta.Groups {
		if v.ID == targetGroupID {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID, IsSelected: true})
		} else {
			groups = append(groups, SelectOptionData{Text: v.Name, Value: v.ID})
		}
	}

	var Years []SelectOptionData
	var targetYear uint

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Where("scores.student_id = ? AND groups.id = ?", Starosta.ID, targetGroupID).
		Order("scores.year desc").
		Select("DISTINCT scores.year as value, concat(scores.year, ' ', '-', ' ', scores.year + 1) as text").
		Scan(&Years)

	if GETParams.Year != nil {
		targetYear = *GETParams.Year
	} else {
		targetYear = uint(Years[0].Value.(int64))
	}

	for i := range Years {
		if uint(Years[i].Value.(int64)) == targetYear {
			Years[i].IsSelected = true
			break
		}
	}

	var Semesters []SelectOptionData
	var targetSemester string

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Where("scores.student_id = ? AND groups.id = ? AND scores.year = ?", Starosta.ID, targetGroupID, targetYear).
		Order("scores.semester desc").
		Select("DISTINCT scores.semester as text").
		Scan(&Semesters)

	if GETParams.Semester != nil {
		targetSemester = *GETParams.Semester
	} else {
		targetSemester = Semesters[0].Text
	}

	for i := range Semesters {
		if Semesters[i].Text == targetSemester {
			Semesters[i].IsSelected = true
			break
		}
	}

	var Disciplines []SelectOptionData
	var targetDisciplineID uint

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND groups_teachers.group_id = ?`, Starosta.ID, targetGroupID, targetYear, targetSemester, targetGroupID).
		Select("DISTINCT disciplines.id as value, disciplines.name as text").
		Scan(&Disciplines)

	if GETParams.DisciplineID != nil {
		targetDisciplineID = *GETParams.DisciplineID
	} else {
		targetDisciplineID = uint(Disciplines[0].Value.(int64))
	}

	for i := range Disciplines {
		if uint(Disciplines[i].Value.(int64)) == targetDisciplineID {
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

	var starostaScoresRows StarostaScoresRows

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND disciplines.id = ? AND groups_teachers.group_id = ?`, Starosta.ID, targetGroupID, targetYear, targetSemester, targetDisciplineID, targetGroupID).
		Select("concat(teachers.last_name, ' ', teachers.first_name, ' ', teachers.middle_name) as teacher, disciplines.cathedra as cathedra").
		Scan(&starostaScoresRows)

	for _, student := range group.Students {
		studentInfo := StarostaScoresStudentInfo{}
		score := Score{}

		studentInfo.StudentName = student.LastName + " " + student.FirstName + " " + student.MiddleName

		sg.db.
			Joins("inner join disciplines on disciplines.id = scores.discipline_id").
			Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
			Joins("inner join groups on disciplines_groups.group_id = groups.id").
			Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
			Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
			Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
			Where(`groups.id = ? AND groups_teachers.group_id = ?`, group.ID, group.ID).
			FirstOrCreate(&score, Score{StudentID: student.ID, Year: targetYear, Semester: targetSemester, DisciplineID: targetDisciplineID})

		studentInfo.ScoreID = score.ID
		studentInfo.AttOne = score.AttestationOne
		studentInfo.AttTwo = score.AttestationTwo
		studentInfo.AttThree = score.AttestationThree
		studentInfo.AttIntermediate = score.IntermediateAttestation

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
			Where(`skips.student_id = ? AND skips.reason NOT in (?) AND skips.discipline_id = ? AND skips.date BETWEEN ? AND ?`, student.ID, []string{""}, targetDisciplineID, from, to).
			Select("COUNT(*)*2 as skips").
			Group("skips.discipline_id").
			Scan(&studentInfo)

		starostaScoresRows.StudentsInfo = append(starostaScoresRows.StudentsInfo, studentInfo)
	}

	r := ScoresResult{
		StudentScoresResult{},
		StarostaScoresResult{
			StarostaScoresFilterData{
				Groups:      groups,
				Years:       Years,
				Semesters:   Semesters,
				Disciplines: Disciplines,
			},
			StarostaScoresResultData{
				Result: starostaScoresRows,
			},
		},
	}

	return &r, nil
}

func (sg *scoreGorm) scoresFilterStudent(Student *Student, GETParams *FilterScoresQueryParams) (*ScoresResult, error) {

	var err error
	var disciplines, semesters, years []SelectOptionData

	if GETParams.Year == nil {

		err = sg.db.
			Model(Score{}).
			Joins("inner join disciplines on disciplines.id = scores.discipline_id").
			Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
			Joins("inner join groups on disciplines_groups.group_id = groups.id").
			Where("scores.student_id = ? AND groups.id = ?", Student.ID, *GETParams.GroupID).
			Order("scores.year desc").
			Select("DISTINCT scores.year as value, concat(scores.year, ' ', '-', ' ', scores.year + 1) as text").
			Scan(&years).
			Error
		if err != nil {
			return nil, err
		}

	} else {
		years = append(years, SelectOptionData{Value: *GETParams.Year})
	}

	if GETParams.Semester == nil {
		err = sg.db.
			Model(Score{}).
			Joins("inner join disciplines on disciplines.id = scores.discipline_id").
			Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
			Joins("inner join groups on disciplines_groups.group_id = groups.id").
			Where("scores.student_id = ? AND groups.id = ? AND scores.year = ?", Student.ID, *GETParams.GroupID, years[0].Value).
			Order("scores.semester desc").
			Select("DISTINCT scores.semester as text").
			Scan(&semesters).
			Error
		if err != nil {
			return nil, err
		}

	} else {
		semesters = append(semesters, SelectOptionData{Text: *GETParams.Semester})
	}

	// наверняка не оптимально. Найти бы способ, как сейвить результат запроса .. хотя а как, ведь без Select-а запрос
	// не выполнится, а значит и нечего будет сейвить. Еще как вариант: писать в select-е as value и as text как для
	// дисциплин, так и для преподов, но они не будут различимы для scan тогда. Можно было бы использовать для этого
	// структурные теги, как на 2:13 в 06_10-Working with Raw Result Rows, но структура SelectOptionData
	// используется не только для дисциплин. Поэтому было принято решение работать с raw result rows, тайминг
	// и видос тот же самый
	// todo обработать ошибку
	rows, err := sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND groups_teachers.group_id = ?`, Student.ID, *GETParams.GroupID, years[0].Value, semesters[0].Text, *GETParams.GroupID).
		Select("disciplines.id, disciplines.name").
		Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var d SelectOptionData
		rows.Scan(&d.Value, &d.Text)
		disciplines = append(disciplines, d)
	}

	r := ScoresResult{
		StudentScoresResult{
			StudentScoresFilterData{
				Years:       years,
				Semesters:   semesters,
				Disciplines: disciplines,
			},
			StudentScoresResultData{},
		},
		StarostaScoresResult{},
	}

	return &r, nil
}

func (sg *scoreGorm) scoresFilterStarosta(Starosta *Student, GETParams *FilterScoresQueryParams) (*ScoresResult, error) {

	var err error
	var disciplines, semesters, years []SelectOptionData

	if GETParams.Year == nil {

		err = sg.db.
			Model(Score{}).
			Joins("inner join disciplines on disciplines.id = scores.discipline_id").
			Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
			Joins("inner join groups on disciplines_groups.group_id = groups.id").
			Where("scores.student_id = ? AND groups.id = ?", Starosta.ID, *GETParams.GroupID).
			Order("scores.year desc").
			Select("DISTINCT scores.year as value, concat(scores.year, ' ', '-', ' ', scores.year + 1) as text").
			Scan(&years).
			Error
		if err != nil {
			return nil, err
		}

	} else {
		years = append(years, SelectOptionData{Value: *GETParams.Year})
	}

	if GETParams.Semester == nil {
		err = sg.db.
			Model(Score{}).
			Joins("inner join disciplines on disciplines.id = scores.discipline_id").
			Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
			Joins("inner join groups on disciplines_groups.group_id = groups.id").
			Where("scores.student_id = ? AND groups.id = ? AND scores.year = ?", Starosta.ID, *GETParams.GroupID, years[0].Value).
			Order("scores.semester desc").
			Select("DISTINCT scores.semester as text").
			Scan(&semesters).
			Error
		if err != nil {
			return nil, err
		}

	} else {
		semesters = append(semesters, SelectOptionData{Text: *GETParams.Semester})
	}

	// наверняка не оптимально. Найти бы способ, как сейвить результат запроса .. хотя а как, ведь без Select-а запрос
	// не выполнится, а значит и нечего будет сейвить. Еще как вариант: писать в select-е as value и as text как для
	// дисциплин, так и для преподов, но они не будут различимы для scan тогда. Можно было бы использовать для этого
	// структурные теги, как на 2:13 в 06_10-Working with Raw Result Rows, но структура SelectOptionData
	// используется не только для дисциплин. Поэтому было принято решение работать с raw result rows, тайминг
	// и видос тот же самый
	// todo обработать ошибку
	rows, err := sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND groups_teachers.group_id = ?`, Starosta.ID, *GETParams.GroupID, years[0].Value, semesters[0].Text, *GETParams.GroupID).
		Select("disciplines.id, disciplines.name").
		Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var d SelectOptionData
		rows.Scan(&d.Value, &d.Text)
		disciplines = append(disciplines, d)
	}

	r := ScoresResult{
		StudentScoresResult{
			StudentScoresFilterData{
				Years:       years,
				Semesters:   semesters,
				Disciplines: disciplines,
			},
			StudentScoresResultData{},
		},
		StarostaScoresResult{},
	}

	return &r, nil
}

func (sg *scoreGorm) Update(score *Score) error {
	return sg.db.Model(&Score{}).Update(score).Error
}

// todo проверка на то, что это действительно староста (автризация) пока не реализована
func (sg *scoreGorm) ReportByUser(user *User, GETParams *ScoresStarostaQueryParams) (*excelize.File, error) {
	// гет параметры, кт должны прилетать (всегда): Группа, Уч год, Семестр

	var studentIDs []uint
	var Starosta Student

	sg.db.
		Table("users_students").
		Where("users_students.user_id = ?", user.ID).
		Pluck("student_id", &studentIDs)

	// todo нельзя воспользоваться хелпером first, т.к. он без preload и where,
	//  поэтому вручную обрабатываем ошибки
	sg.db.Where("id = ?", studentIDs[0]).Preload("Groups.Starosta").First(&Starosta)

	var result []StarostaScoresRows

	sg.db.
		Model(Score{}).
		Joins("inner join disciplines on disciplines.id = scores.discipline_id").
		Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
		Joins("inner join groups on disciplines_groups.group_id = groups.id").
		Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
		Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
		Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
		Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND groups_teachers.group_id = ?`, Starosta.ID, *GETParams.GroupID, *GETParams.Year, *GETParams.Semester, *GETParams.GroupID).
		Select("disciplines.name as discipline, disciplines.id as discipline_id, concat(teachers.last_name, ' ', teachers.first_name, ' ', teachers.middle_name) as teacher").
		Scan(&result)

	// формирование результата

	f := excelize.NewFile()

	var group Group

	sg.db.
		Where("id = ?", *GETParams.GroupID).
		Preload("Students", func(db *gorm.DB) *gorm.DB { return db.Order("students.last_name") }).
		Find(&group)

	for i, row := range result {
		for _, student := range group.Students {
			studentInfo := StarostaScoresStudentInfo{}

			studentInfo.StudentName = student.LastName + " " + student.FirstName + " " + student.MiddleName

			var from time.Time
			var to time.Time

			if *GETParams.Semester == "Осенний" {
				// todo не уверен насчет utc (как он отличается от gmt + 7)
				//  + к тому, что всё ок здесь из документации к пакету time:
				//  "in the appropriate zone for that time in the given location."
				from = time.Date(int(*GETParams.Year), time.September, 1, 0, 0, 0, 0, time.UTC)
				to = time.Date(int(*GETParams.Year)+1, time.January, 31, 23, 59, 59, 59, time.UTC)
			} else {
				from = time.Date(int(*GETParams.Year)+1, time.February, 1, 0, 0, 0, 0, time.UTC)
				to = time.Date(int(*GETParams.Year)+1, time.June, 31, 23, 59, 59, 59, time.UTC)
			}

			sg.db.
				Model(Score{}).
				Joins("inner join disciplines on disciplines.id = scores.discipline_id").
				Joins("inner join disciplines_groups on disciplines_groups.discipline_id = disciplines.id").
				Joins("inner join groups on disciplines_groups.group_id = groups.id").
				Joins("INNER JOIN disciplines_teachers ON disciplines.id = disciplines_teachers.discipline_id").
				Joins("INNER JOIN teachers ON teachers.id = disciplines_teachers.teacher_id").
				Joins("INNER JOIN (SELECT skips.discipline_id, COUNT(*)*2 hours FROM skips WHERE skips.student_id = ? AND skips.reason NOT in (?) AND skips.date BETWEEN ? AND ? GROUP BY skips.discipline_id) foo ON foo.discipline_id = disciplines.id", student.ID, []string{""}, from, to).
				Joins("INNER JOIN groups_teachers ON teachers.id = groups_teachers.teacher_id").
				Where(`scores.student_id = ? AND groups.id = ? AND scores.year = ? AND scores.semester = ? AND disciplines.id = ? AND groups_teachers.group_id = ?`, student.ID, group.ID, *GETParams.Year, *GETParams.Semester, row.DisciplineID, group.ID).
				Select("scores.attestation_one as att_one, scores.attestation_two as att_two, scores.attestation_three as att_three, scores.intermediate_attestation as att_intermediate, foo.hours as skips").
				Scan(&studentInfo)

			result[i].StudentsInfo = append(result[i].StudentsInfo, studentInfo)
		}

	}

	style, _ := f.NewStyle(`{"alignment":{"horizontal":"center","vertical":"center"}}`)
	f.SetCellStyle("Sheet1", "A1", "A1", style)

	f.MergeCell("Sheet1", "A1", "A3")
	f.SetCellValue("Sheet1", "A1", "№")
	f.SetCellValue("Sheet1", "B1", "Дисциплина")
	f.SetCellValue("Sheet1", "B2", "Преподаватель")
	f.SetCellValue("Sheet1", "B3", `ФИО \ Аттестация`)

	// возможно не оптимально и это как-то можно объединить с циклом выше
	// + здесь пока не учитывается, что столбцы могут быть AA, AB .. BA, BB...
	counter := 1
	for _, studentInfo := range result[0].StudentsInfo {
		rowNumber := strconv.Itoa(counter + 3)
		f.SetCellValue("Sheet1", "A"+rowNumber, counter)
		f.SetCellValue("Sheet1", "B"+rowNumber, studentInfo.StudentName)

		counter++
	}

	column := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	disciplineNameColumnStart := 2

	f.SetColWidth("Sheet1", "B", "B", 31)
	f.SetColWidth("Sheet1", "A", "A", 3)

	for i := range result {
		f.MergeCell("Sheet1", string(column[disciplineNameColumnStart])+"1", string(column[disciplineNameColumnStart+3])+"1")
		f.MergeCell("Sheet1", string(column[disciplineNameColumnStart])+"2", string(column[disciplineNameColumnStart+3])+"2")

		f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart])+"1", result[i].Discipline)
		f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart])+"2", result[i].Teacher)
		f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart])+"3", 1)
		f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart+1])+"3", 2)
		f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart+2])+"3", 3)
		f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart+3])+"3", "Пропущенные часы")
		f.SetColWidth("Sheet1", string(column[disciplineNameColumnStart+3]), string(column[disciplineNameColumnStart+3]), 18)

		row := 4
		for j, v := range result[i].StudentsInfo {
			rowNumber := strconv.Itoa(row + j)

			f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart])+rowNumber, v.AttOne)
			f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart+1])+rowNumber, v.AttTwo)
			f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart+2])+rowNumber, v.AttThree)
			f.SetCellValue("Sheet1", string(column[disciplineNameColumnStart+3])+rowNumber, v.Skips)

		}
		disciplineNameColumnStart += 4
	}

	return f, nil
}

// todo rename to "Grade"
type Score struct {
	gorm.Model
	// todo лучше поменять uint на int
	Year                    uint   // год начала обучения, e.g. если учебный год 2016 - 2017, то Year = 2016
	AttestationOne          string `sql:"type:score_value"`
	AttestationTwo          string `sql:"type:score_value"`
	AttestationThree        string `sql:"type:score_value"`
	IntermediateAttestation string `sql:"type:score_value"`
	Semester                string `sql:"type:semester"`
	StudentID               uint
	DisciplineID            uint
}
