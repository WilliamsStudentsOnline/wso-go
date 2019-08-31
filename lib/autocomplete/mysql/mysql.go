package mysql

import (
	"sort"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"github.com/m1ome/leven"
)

type Autocomplete struct {
	DB *gorm.DB
}

func (a *Autocomplete) AreaOfStudy(q string) ([]autocomplete.ACEntry, error) {
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
	entries := make([]autocomplete.ACEntry, len(areasName)+len(areasAbbrev))
	// Add in the names
	for i := range areasName {
		entries[i] = autocomplete.ACEntry{
			ID:    areas[i].ID,
			Value: areas[i].Name,
		}
	}
	// Add in the abbreviations
	lenAreasName := len(areasName)
	for i := range areasAbbrev {
		entries[lenAreasName+i] = autocomplete.ACEntry{
			ID:    areasAbbrev[i].ID,
			Value: areasAbbrev[i].Abbreviation,
		}
	}

	sortByDistance(q, entries)

	return entries, nil
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

const (
	ACTypeCourse = "course"
	ACTypeArea   = "area"
)

func (a *Autocomplete) Course(q string) ([]autocomplete.ACEntry, error) {
	q = strings.ToLower(q)
	parts := strings.Split(q, " ")

	// Get areas
	areas, err := a.dbGetAreaAbbrev(parts[0])
	if err != nil {
		return nil, err
	}

	// If no areas, return nothing
	if len(areas) == 0 {
		return []autocomplete.ACEntry{}, nil
	}

	// If multiple areas, return autocomplete for the areas (just abbreviation)
	if len(areas) > 1 {
		entries := make([]autocomplete.ACEntry, len(areas))
		for i := range areas {
			entries[i] = autocomplete.ACEntry{
				ID:    areas[i].ID,
				Value: areas[i].Abbreviation,
				Type:  ACTypeArea,
			}
		}
		sortByDistance(parts[0], entries)
		return entries, nil
	}

	// If the abbreviation does not equal our passed abbreviation (aka we haven't finished
	// typing it yet: eg "csc" for "csci"), return just the abbreviation.
	if strings.ToLower(areas[0].Abbreviation) != parts[0] {
		return []autocomplete.ACEntry{{ID: areas[0].ID, Value: areas[0].Abbreviation, Type: ACTypeArea}}, nil
	}

	// Once we have only one area, return autocomplete for that

	// Get courses
	var courses []models.Course
	tx := a.DB.Model(&models.Course{}).
		Select("id, number").
		Where("area_of_study_id = ?", areas[0].ID)
	if len(parts) > 1 {
		tx = tx.Where("lower(number) LIKE ?", parts[1]+"%")
	}
	err = tx.Find(&courses).Error
	if err != nil {
		return nil, err
	}

	// Return course names (with the abbreviation)
	entries := make([]autocomplete.ACEntry, len(courses))
	for i := range courses {
		entries[i] = autocomplete.ACEntry{
			ID:    courses[i].ID,
			Value: areas[0].Abbreviation + " " + courses[i].Number,
			Type:  ACTypeCourse,
		}
	}

	sortByDistance(q, entries)

	return entries, nil
}

func (a *Autocomplete) Professor(q string) ([]autocomplete.ACEntry, error) {
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

	entries := make([]autocomplete.ACEntry, len(profs))
	for i := range profs {
		entries[i] = autocomplete.ACEntry{
			ID:    profs[i].ID,
			Value: profs[i].Name,
		}
	}

	sortByDistance(q, entries)

	return entries, nil
}

func (a *Autocomplete) Tag(q string) ([]autocomplete.ACEntry, error) {
	q = strings.ToLower(q)

	var tags []models.Tag

	// Do SQL
	err := a.DB.Model(&models.Tag{}).
		Select("id, name").
		Where("lower(name) LIKE ?", "%"+q+"%").
		Find(&tags).Error
	if err != nil {
		return nil, err
	}

	entries := make([]autocomplete.ACEntry, len(tags))
	for i := range tags {
		entries[i] = autocomplete.ACEntry{ID: tags[i].ID, Value: tags[i].Name}
	}

	sortByDistance(q, entries)

	return entries, nil
}

type levDist struct {
	val  autocomplete.ACEntry
	dist int
}

// Sorts the resulting strings by levenshtein distance from the original query
func sortByDistance(q string, entries []autocomplete.ACEntry) {
	dists := make([]levDist, len(entries))

	// Calculate the distances
	for i, val := range entries {
		dists[i] = levDist{
			val:  val,
			dist: leven.Distance(q, strings.ToLower(val.Value)),
		}
	}

	// Sort
	sort.Slice(dists, func(i, j int) bool {
		return dists[i].dist < dists[j].dist
	})

	// Copy the sorted list into the original list of resulting strings
	for i := range dists {
		entries[i] = dists[i].val
	}
}
