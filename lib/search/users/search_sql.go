package users

import (
	"sort"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"github.com/m1ome/leven"
)

var FieldNamesStd = map[string]string{
	"name":         "name",
	"class":        "class_year",
	"year":         "class_year",
	"class_year":   "class_year",
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
	"type":         "type",
	"pronoun":      "pronoun",
	"pronouns":     "pronoun",
}

type SearchUsersMySQL struct {
	DB *gorm.DB
}

// Search. Sorts by distance iff opts.Start is not specified.
func (s *SearchUsersMySQL) Search(query string, opts SearchOptions) (users []*models.User, totalResults int, err error) {
	ast, err := search.ParseSearchQuery(query)
	if err != nil {
		return
	}

	at := astTraverser{
		ast:           ast,
		query:         new(strings.Builder),
		vals:          []interface{}{},
		simpleQueries: []string{},
		db:            s.DB,
		errors:        []error{},
	}
	at.Traverse()

	if len(at.errors) > 0 {
		return users, totalResults, at.errors[len(at.errors)-1]
	}

	tx := s.DB.Model(&models.User{})
	tx = at.ConstructSQL(tx)
	tx = tx.Where("users.at_williams = ? AND users.visible = ?", true, true)
	// Run options
	if opts != nil {
		// TODO: either add counting here or no pagination for search
		tx = opts.Preloader(tx)
		// TODO: set ordering to be default for all calls or make ordering customizable
		// Order if we are having options
		tx = opts.Order(tx)
	}

	err = tx.Find(&users).Error
	if err != nil {
		return
	}

	// Post order sort by distance query
	bigQuery := new(strings.Builder)
	for _, simpleQuery := range at.simpleQueries {
		bigQuery.WriteString(simpleQuery)
	}

	// Total results before pagination
	totalResults = len(users)

	// Paginate server-side, rather than database-side
	if opts != nil {
		opts.PostOrder(bigQuery.String(), users)
		users = opts.Paginate(users)
	}
	return
}

type astTraverser struct {
	ast           *search.Query
	query         *strings.Builder
	vals          []interface{}
	simpleQueries []string
	db            *gorm.DB
	traversed     bool
	errors        []error
}

func (t *astTraverser) Traverse() {
	t.constructOr(t.ast.Or)
	t.traversed = true
}

func (t *astTraverser) ConstructSQL(tx *gorm.DB) *gorm.DB {
	if !t.traversed {
		t.Traverse()
	}
	return tx.Where(t.query.String(), t.vals...)
}

// Construct OR expressions together in the SQL.
func (t *astTraverser) constructOr(exps []*search.Expression) {
	t.query.WriteRune('(')
	for i, exp := range exps {
		t.constructAnd(exp.And)
		if i+1 < len(exps) {
			t.query.WriteString(" OR ")
		}
	}
	t.query.WriteRune(')')
}

// Construct AND expressions together in the SQL.
func (t *astTraverser) constructAnd(conds []*search.Condition) {
	t.query.WriteRune('(')
	for i, cond := range conds {
		ignore := t.parseCondition(cond)
		// Write AND iff it isn't the last one and we didn't ignore the current condition
		if i+1 < len(conds) && !ignore {
			t.query.WriteString(" AND ")
		}
	}
	t.query.WriteRune(')')
}

// Parse a specific individual condition (eg a field, value, or sub-expression)
func (t *astTraverser) parseCondition(cond *search.Condition) (ignore bool) {

	if cond.Or != nil {
		t.constructOr(cond.Or)
		//vals = append(vals, recVals...)
	} else if cond.Value != nil {
		t.parseBasicValue(cond.Value.ToString())
		//vals = append(vals, recVals...)
	} else if cond.Field != nil {
		ignore = t.parseField(cond.Field)
		//vals = append(vals, recVals...)
	} else {
		ignore = true
	}

	return
}

// If the condition is just a regular value, just parse that and look it up via the default search fields.
func (t *astTraverser) parseBasicValue(val string) {
	t.query.WriteString("users.search_fields LIKE ?")
	lower := strings.ToLower(val)
	t.vals = append(t.vals, "%"+lower+"%")
	t.simpleQueries = append(t.simpleQueries, lower)
}

// This constructs the search by field part of the query. Because of the complexities of relational databases,
// some of these flags require multiple subqueries or complex joins.
// If you want performance, don't use MySQL as a search engine.
// TODO: research and see if joins would make this faster
func (t *astTraverser) parseField(field *search.Field) (ignore bool) {
	fieldName, ok := FieldNamesStd[strings.ToLower(field.Key)]
	if !ok {
		t.errors = append(t.errors, lib.NewErrorUnknownSearchField(field.Key))
		ignore = true
		return
	}
	fieldValue := strings.ToLower(field.Value.ToString())
	valueSearch := "%" + fieldValue + "%"

	switch fieldName {
	case "building":
		t.query.WriteString("users.office_id IN (?)")
		t.vals = append(t.vals,
			t.db.Model(&models.Office{}).
				Select("offices.id").
				Where("lower(offices.number) LIKE ?", valueSearch).QueryExpr())
	case "dorm":
		t.query.WriteString("users.dorm_visible = ? AND users.dorm_room_id IN (?)")
		t.vals = append(t.vals, true,
			t.db.Model(&models.DormRoom{}).
				Select("dorm_rooms.id").
				Where("dorm_rooms.dorm_id IN (?)",
					t.db.Model(&models.Dorm{}).
						Select("dorms.id").
						Where("lower(dorms.name) LIKE ?", valueSearch).QueryExpr(),
				).QueryExpr())
	case "room":
		t.query.WriteString("users.dorm_visible = ? AND users.dorm_room_id IN (?)")
		t.vals = append(t.vals, true,
			t.db.Model(&models.DormRoom{}).
				Select("dorm_rooms.id").
				Where("lower(dorm_rooms.number) LIKE ?", valueSearch).QueryExpr())
	case "department":
		// Have to get department or area of study abbreviation
		t.query.WriteString("users.department_id IN (?) OR users.department_id IN (?)")
		t.vals = append(t.vals,
			t.db.Model(&models.Department{}).
				Select("departments.id").
				Where("lower(departments.name) LIKE ?", valueSearch).QueryExpr(),
			t.db.Model(&models.AreaOfStudy{}).
				Select("areas_of_study.department_id").
				Where("areas_of_study.abbrev LIKE ?", strings.ToUpper(valueSearch)).QueryExpr())
	case "neighborhood":
		// Get neighborhood by getting all dorm rooms associated with a specific neighborhood.
		// Essentially, it is a join on users, dorm rooms, dorms, and neighborhoods (but as subqueries)
		// This in particular is massively inefficient.
		t.query.WriteString("users.dorm_visible = ? AND users.dorm_room_id IN (?)")
		t.vals = append(t.vals, true,
			t.db.Model(&models.DormRoom{}).
				Select("dorm_rooms.id").
				Where("dorm_rooms.dorm_id IN (?)",
					t.db.Model(&models.Dorm{}).
						Select("dorms.id").
						Where("dorms.neighborhood_id IN (?)",
							t.db.Model(&models.Neighborhood{}).
								Select("neighborhoods.id").
								Where("lower(neighborhoods.name) LIKE ?", valueSearch).QueryExpr(),
						).QueryExpr(),
				).QueryExpr())
	case "tags":
		t.query.WriteString("users.id IN (?)")
		// Get tags matching name, then look t the many to many join table for user ids with that tag id, then
		// look up all users with those ids.
		t.vals = append(t.vals,
			t.db.Table("tags_users").
				Select("tags_users.user_id").
				Where("tags_users.tag_id IN (?)",
					t.db.Model(&models.Tag{}).
						Select("tags.id").
						Where("lower(tags.name) LIKE ?", valueSearch).QueryExpr(),
				).QueryExpr(),
		)
	case "home_zip", "home_town", "home_state", "home_country":
		t.query.WriteString("users.home_visible = ? AND lower(users." + fieldName + ") LIKE ?")
		t.vals = append(t.vals, true, valueSearch)
	case "name", "class_year", "entry", "unix_id", "williams_email", "title", "major", "campus_phone_ext":
		t.query.WriteString("lower(users." + fieldName + ") LIKE ?")
		t.vals = append(t.vals, valueSearch)
	case "type":
		t.query.WriteString("users.type = ?")
		t.vals = append(t.vals, fieldValue)
	case "pronoun":
		t.query.WriteString("users.pronoun = ?")
		t.vals = append(t.vals, fieldValue)
	default:
		ignore = true
		t.errors = append(t.errors, lib.NewErrorUnknownSearchField(fieldName))
	}

	return
}

type levDist struct {
	user *models.User
	dist int
}

// Sorts the resulting strings by levenshtein distance from the original query
func sortByDistance(q string, users []*models.User) {
	dists := make([]levDist, len(users))

	// Calculate the distances
	for i, val := range users {
		dists[i] = levDist{
			user: val,
			dist: leven.Distance(q, strings.ToLower(val.SearchFields)),
		}
	}

	// Sort
	sort.Slice(dists, func(i, j int) bool {
		return dists[i].dist < dists[j].dist
	})

	// Copy the sorted list into the original list of resulting strings
	for i := range dists {
		users[i] = dists[i].user
	}
}

type SearchUsersOptionsMysql struct {
	*models.GetAllUsersOptions
}

func (o *SearchUsersOptionsMysql) Preloader(db *gorm.DB) *gorm.DB {
	return o.GetAllUsersOptions.Preloader(db)
}

func (o *SearchUsersOptionsMysql) Order(db *gorm.DB) *gorm.DB {
	if o.Start == nil {
		return db
	}
	return o.GetAllUsersOptions.Order(db)
}

func (o *SearchUsersOptionsMysql) Paginate(users []*models.User) []*models.User {
	// Assume users slice is ordered by name!
	if o.Start != nil {
		paginateIdx := 0
		for i := range users {
			// Loop until we find a user that is after start. Then we use that index to cut all users before start
			if users[i].Name > *o.Start {
				paginateIdx = i
				break
			}
		}
		users = users[paginateIdx:]
	}

	if o.Limit != nil {
		// Only do offset if we have limit
		if o.Offset != nil {
			// Ensure bounds
			if *o.Offset < uint(len(users))-1 {
				users = users[*o.Offset:]
			}
		}

		// Ensure bounds
		if *o.Limit < uint(len(users)) {
			users = users[:*o.Limit]
		}
	}

	return users
}

func (o *SearchUsersOptionsMysql) PostOrder(q string, users []*models.User) {
	// Only sort by distance iff start isn't specified
	if o.Start == nil {
		sortByDistance(q, users)
	}
}

func (*SearchUsersMySQL) NewOptions(offset *uint, limit *uint, preload []string) SearchOptions {
	return &SearchUsersOptionsMysql{&models.GetAllUsersOptions{
		Offset:  offset,
		Limit:   limit,
		Preload: preload,
	}}
}
