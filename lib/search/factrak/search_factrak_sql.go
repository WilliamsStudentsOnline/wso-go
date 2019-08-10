package factrak

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchFactrakMySQL struct {
	DB *gorm.DB
}

func NewSearchFactrakMySQL(db *gorm.DB) *SearchFactrakMySQL {
	return &SearchFactrakMySQL{
		DB: db,
	}
}

func (s *SearchFactrakMySQL) SearchProfessors(query string, users *[]*models.User, opts models.Options) (err error) {
	words := strings.Split(strings.ToLower(query), " ")

	// Generate an SQL query with each word being different
	sqlSB := new(strings.Builder)
	sqlParams := make([]interface{}, len(words))

	for i := range words {
		sqlSB.WriteString("lower(name) LIKE ?")
		sqlParams[i] = "%" + words[i] + "%"
		if i+1 < len(words) {
			sqlSB.WriteString(" AND ")
		}
	}

	// Do SQL
	tx := s.DB.Model(&models.User{})
	tx = tx.Where(sqlSB.String(), sqlParams...)
	tx = tx.Where("users.at_williams = ? AND users.type = ?", true, models.UserTypeProfessor)
	// Run options
	if opts != nil {
		tx = opts.Paginate(tx)
		tx = opts.Preloader(tx)
	}

	err = tx.Find(users).Error
	return
}

func (s *SearchFactrakMySQL) SearchCourses(query string, courses *[]*models.Course, opts models.Options) (err error) {
	words := strings.Split(strings.ToLower(query), " ")
	if len(words) == 0 {
		return
	}

	aos := words[0]

	// Search by area of study first
	tx := s.DB.Model(&models.Course{}).
		Where("courses.area_of_study_id IN (?)",
			s.DB.Model(&models.AreaOfStudy{}).
				Where("lower(areas_of_study.name) LIKE ? "+
					"OR lower(areas_of_study.name) LIKE ? "+
					"OR areas_of_study.abbrev LIKE ? ",
					aos+"%",
					"% "+aos+"%",
					strings.ToUpper(aos+"%"),
				).QueryExpr(),
		)

	if len(words) > 1 {
		tx = tx.Where("lower(courses.number) LIKE ?", words[1]+"%")
	}

	// Run options
	if opts != nil {
		tx = opts.Paginate(tx)
		tx = opts.Preloader(tx)
	}

	err = tx.Find(courses).Error
	return
}

func (*SearchFactrakMySQL) NewProfessorsOptions(offset *uint, limit *uint, preload *[]string) models.Options {
	return &models.GetAllProfessorsOptions{
		Offset:  offset,
		Limit:   limit,
		Preload: preload,
	}
}

func (*SearchFactrakMySQL) NewCoursesOptions(offset *uint, limit *uint, preload *[]string) models.Options {
	return &models.GetAllCoursesOptions{
		Offset:  offset,
		Limit:   limit,
		Preload: preload,
	}
}
