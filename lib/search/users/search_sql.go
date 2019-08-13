package users

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

var FieldNamesStd = map[string]string{
	"name":         "name",
	"class":        "class_year",
	"year":         "class_year",
	"neighborhood": "neighborhood",
	"cluster":      "neighborhood",
	"room":         "room",
	"entry":        "entry",
	"dorm":         "dorm",
	"building":     "building",
	"bldg":         "building",
	"unix":         "unix_id",
	"email":        "williams_email",
	"title":        "title",
	"department":   "department",
	"dept":         "department",
	"major":        "major",
	"zip":          "home_zip",
	"city":         "home_town",
	"town":         "home_town",
	"state":        "home_state",
	"country":      "home_country",
	"phone":        "campus_phone_ext",
	"ext":          "campus_phone_ext",
	"tag":          "tags",
	"tags":         "tags",
}

type SearchUsersMySQL struct {
	DB *gorm.DB
}

func (s *SearchUsersMySQL) Search(query string, users *[]*models.User, opts SearchOptions) (err error) {
	ast, err := search.ParseSearchQuery(query)
	if err != nil {
		return
	}

	tx := s.DB.Model(&models.User{})
	tx = s.recursiveConstruct(ast, tx)
	tx = tx.Where("users.at_williams = ? AND users.visible = ?", true, true)
	// Run options
	if opts != nil {
		tx = opts.Paginate(tx)
		tx = opts.Preloader(tx)
	}

	err = tx.Find(users).Error
	return
}

func (s *SearchUsersMySQL) recursiveConstruct(ast *search.Query, tx *gorm.DB) *gorm.DB {
	sb := new(strings.Builder)

	vals := s.constructOr(ast.Or, sb)

	return tx.Where(sb.String(), vals...)
}

// Construct OR expressions together in the SQL.
func (s *SearchUsersMySQL) constructOr(exps []*search.Expression, sb *strings.Builder) (vals []interface{}) {
	sb.WriteRune('(')
	for i, exp := range exps {
		recVals := s.constructAnd(exp.And, sb)
		vals = append(vals, recVals...)
		if i+1 < len(exps) {
			sb.WriteString(" OR ")
		}
	}
	sb.WriteRune(')')

	return vals
}

// Construct AND expressions together in the SQL.
func (s *SearchUsersMySQL) constructAnd(conds []*search.Condition, sb *strings.Builder) (vals []interface{}) {
	sb.WriteRune('(')
	for i, cond := range conds {
		recVals, ignore := s.parseCondition(cond, sb)
		vals = append(vals, recVals...)
		// Write AND iff it isn't the last one and we didn't ignore the current condition
		if i+1 < len(conds) && !ignore {
			sb.WriteString(" AND ")
		}
	}
	sb.WriteRune(')')

	return vals
}

// Parse a specific individual condition (eg a field, value, or sub-expression)
func (s *SearchUsersMySQL) parseCondition(cond *search.Condition, sb *strings.Builder) (vals []interface{}, ignore bool) {
	if cond.Or != nil {
		recVals := s.constructOr(cond.Or, sb)
		vals = append(vals, recVals...)
	} else if cond.Value != nil {
		recVals := s.parseBasicValue(cond.Value.ToString(), sb)
		vals = append(vals, recVals...)
	} else if cond.Field != nil {
		var recVals []interface{}
		recVals, ignore = s.parseField(cond.Field, sb)
		vals = append(vals, recVals...)
	} else {
		ignore = true
	}

	return
}

// If the condition is just a regular value, just parse that and look it up via the default search fields.
func (s *SearchUsersMySQL) parseBasicValue(val string, sb *strings.Builder) (vals []interface{}) {
	sb.WriteString("users.search_fields LIKE ?")
	vals = append(vals, "%"+strings.ToLower(val)+"%")
	return
}

// This constructs the search by field part of the query. Because of the complexities of relational databases,
// some of these flags require multiple subqueries or complex joins.
// If you want performance, don't use MySQL as a search engine.
// TODO: research and see if joins would make this faster
func (s *SearchUsersMySQL) parseField(field *search.Field, sb *strings.Builder) (vals []interface{}, ignore bool) {
	fieldName, ok := FieldNamesStd[strings.ToLower(field.Key)]
	if !ok {
		ignore = true
		return
	}
	fieldValue := strings.ToLower(field.Value.ToString())
	valueSearch := "%" + fieldValue + "%"

	switch fieldName {
	case "building":
		sb.WriteString("users.office_id IN (?)")
		vals = append(vals,
			s.DB.Model(&models.Office{}).
				Select("offices.id").
				Where("lower(offices.number) LIKE ?", valueSearch).QueryExpr())
	case "dorm":
		sb.WriteString("users.dorm_visible = ? AND users.dorm_room_id IN (?)")
		vals = append(vals, true,
			s.DB.Model(&models.DormRoom{}).
				Select("dorm_rooms.id").
				Where("dorm_rooms.dorm_id IN (?)",
					s.DB.Model(&models.Dorm{}).
						Select("dorms.id").
						Where("lower(dorms.name) LIKE ?", valueSearch).QueryExpr(),
				).QueryExpr())
	case "room":
		sb.WriteString("users.dorm_visible = ? AND users.dorm_room_id IN (?)")
		vals = append(vals, true,
			s.DB.Model(&models.DormRoom{}).
				Select("dorm_rooms.id").
				Where("lower(dorm_rooms.number) LIKE ?", valueSearch).QueryExpr())
	case "department":
		// Have to get department or area of study abbreviation
		sb.WriteString("users.department_id IN (?) OR users.department_id IN (?)")
		vals = append(vals,
			s.DB.Model(&models.Department{}).
				Select("departments.id").
				Where("lower(departments.name) LIKE ?", valueSearch).QueryExpr(),
			s.DB.Model(&models.AreaOfStudy{}).
				Select("areas_of_study.department_id").
				Where("areas_of_study.abbrev LIKE ?", strings.ToUpper(valueSearch)).QueryExpr())
	case "neighborhood":
		// Get neighborhood by getting all dorm rooms associated with a specific neighborhood.
		// Essentially, it is a join on users, dorm rooms, dorms, and neighborhoods (but as subqueries)
		// This in particular is massively inefficient.
		sb.WriteString("users.dorm_visible = ? AND users.dorm_room_id IN (?)")
		vals = append(vals, true,
			s.DB.Model(&models.DormRoom{}).
				Select("dorm_rooms.id").
				Where("dorm_rooms.dorm_id IN (?)",
					s.DB.Model(&models.Dorm{}).
						Select("dorms.id").
						Where("dorms.neighborhood_id IN (?)",
							s.DB.Model(&models.Neighborhood{}).
								Select("neighborhoods.id").
								Where("lower(neighborhoods.name) LIKE ?", valueSearch).QueryExpr(),
						).QueryExpr(),
				).QueryExpr())
	case "tags":
		sb.WriteString("users.id IN (?)")
		// Get tags matching name, then look at the many to many join table for user ids with that tag id, then
		// look up all users with those ids.
		vals = append(vals,
			s.DB.Table("tags_users").
				Select("tags_users.user_id").
				Where("tags_users.tag_id IN (?)",
					s.DB.Model(&models.Tag{}).
						Select("tags.id").
						Where("lower(tags.name) LIKE ?", valueSearch).QueryExpr(),
				).QueryExpr(),
		)
	case "home_zip", "home_town", "home_state", "home_country":
		sb.WriteString("users.home_visible = ? AND lower(users." + fieldName + ") LIKE ?")
		vals = append(vals, true, valueSearch)
	case "name", "class_year", "entry", "unix_id", "williams_email", "title", "major", "campus_phone_ext":
		sb.WriteString("lower(users." + fieldName + ") LIKE ?")
		vals = append(vals, valueSearch)
	default:
		ignore = true
	}

	return
}

func (*SearchUsersMySQL) NewOptions(offset *uint, limit *uint, preload *[]string) SearchOptions {
	return &models.GetAllUsersOptions{
		Offset:  offset,
		Limit:   limit,
		Preload: preload,
	}
}
