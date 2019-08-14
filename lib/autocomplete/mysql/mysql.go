package mysql

import (
	"sort"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"github.com/m1ome/leven"
)

type Autocomplete struct {
	DB *gorm.DB
}

func (a *Autocomplete) AreaOfStudy(q string) ([]string, error) {
	q = strings.ToLower(q)

	// Get area of study by name
	var areasName []models.AreaOfStudy
	err := a.DB.Model(&models.AreaOfStudy{}).
		Select("name").
		Where("lower(name) LIKE ? or lower(name) LIKE ?", q+"%", "% "+q+"%").
		Find(&areasName).Error
	if err != nil {
		return nil, err
	}

	// Get areas by abbreviation
	areasAbbrev, err := a.dbGetAreaAbbrev(q)
	if err != nil {
		return nil, err
	}

	// Combine the area types
	areas := make([]models.AreaOfStudy, len(areasName)+len(areasAbbrev))
	copy(areas, areasName)
	copy(areas[len(areasName):], areasAbbrev)

	// Turn the areas into a combined list of strings
	strs := make([]string, len(areasName)+len(areasAbbrev))
	// Add in the names
	for i := range areasName {
		strs[i] = areas[i].Name
	}
	// Add in the abbreviations
	lenAreasName := len(areasName)
	for i := range areasAbbrev {
		strs[lenAreasName+i] = areasAbbrev[i].Abbreviation
	}

	sortByDistance(q, strs)

	return strs, nil
}

func (a *Autocomplete) dbGetAreaAbbrev(q string) ([]models.AreaOfStudy, error) {
	var areasAbbrev []models.AreaOfStudy
	abbrev := strings.ToUpper(q)

	err := a.DB.Model(&models.AreaOfStudy{}).
		Select("id, abbrev").
		Where("abbrev LIKE ?", abbrev+"%").
		Find(&areasAbbrev).Error
	if err != nil {
		return nil, err
	}

	return areasAbbrev, nil
}

func (a *Autocomplete) Course(q string) ([]string, error) {
	q = strings.ToLower(q)
	parts := strings.Split(q, " ")

	// Get areas
	areas, err := a.dbGetAreaAbbrev(parts[0])
	if err != nil {
		return nil, err
	}

	// If no areas, return nothing
	if len(areas) == 0 {
		return []string{}, nil
	}

	// If multiple areas, return autocomplete for the areas (just abbreviation)
	if len(areas) > 1 {
		strs := make([]string, len(areas))
		for i := range areas {
			strs[i] = areas[i].Abbreviation
		}
		sortByDistance(parts[0], strs)
		return strs, nil
	}

	// If the abbreviation does not equal our passed abbreviation (aka we haven't finished
	// typing it yet: eg "csc" for "csci"), return just the abbreviation.
	if strings.ToLower(areas[0].Abbreviation) != parts[0] {
		return []string{areas[0].Abbreviation}, nil
	}

	// Once we have only one area, return autocomplete for that

	// Get courses
	var courses []models.Course
	tx := a.DB.Model(&models.Course{}).
		Select("number").
		Where("area_of_study_id = ?", areas[0].ID)
	if len(parts) > 1 {
		tx = tx.Where("lower(number) LIKE ?", parts[1]+"%")
	}
	err = tx.Find(&courses).Error
	if err != nil {
		return nil, err
	}

	// Return course names (with the abbreviation)
	strs := make([]string, len(courses))
	for i := range courses {
		strs[i] = areas[0].Abbreviation + " " + courses[i].Number
	}

	sortByDistance(q, strs)

	return strs, nil
}

func (a *Autocomplete) Professor(q string) ([]string, error) {
	q = strings.ToLower(q)
	words := strings.Split(q, " ")

	// Generate an SQL query with each word being different
	sqlSB := new(strings.Builder)
	sqlParams := make([]interface{}, len(words))

	for i := range words {
		sqlSB.WriteString("lower(name) LIKE ?")
		sqlParams[i] = "%" + words[i] + "%"
		if i+1 < len(words) {
			sqlSB.WriteString(" OR ")
		}
	}

	var profs []models.User

	// Do SQL
	err := a.DB.Model(&models.User{}).
		Select("name").
		Where(sqlSB.String(), sqlParams...).
		Where("users.at_williams = ? AND users.type = ?", true, models.UserTypeProfessor).
		Find(&profs).Error
	if err != nil {
		return nil, err
	}

	strs := make([]string, len(profs))
	for i := range profs {
		strs[i] = profs[i].Name
	}

	sortByDistance(q, strs)

	return strs, nil
}

func (a *Autocomplete) Tag(q string) ([]string, error) {
	q = strings.ToLower(q)

	var tags []models.Tag

	// Do SQL
	err := a.DB.Model(&models.Tag{}).
		Select("name").
		Where("lower(name) LIKE ?", "%"+q+"%").
		Find(&tags).Error
	if err != nil {
		return nil, err
	}

	strs := make([]string, len(tags))
	for i := range tags {
		strs[i] = tags[i].Name
	}

	sortByDistance(q, strs)

	return strs, nil
}

type levDist struct {
	str  string
	dist int
}

// Sorts the resulting strings by levenshtein distance from the original query
func sortByDistance(q string, strs []string) {
	dists := make([]levDist, len(strs))

	// Calculate the distances
	for i, val := range strs {
		dists[i] = levDist{
			str:  val,
			dist: leven.Distance(q, strings.ToLower(val)),
		}
	}

	// Sort
	sort.Slice(dists, func(i, j int) bool {
		return dists[i].dist < dists[j].dist
	})

	// Copy the sorted list into the original list of resulting strings
	for i := range dists {
		strs[i] = dists[i].str
	}
}
