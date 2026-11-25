// Package areas adds areas of study the major plan needs, matching departments by name since ids differ between dev and prod
package areas

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Area struct {
	Name             string
	Abbreviation     string
	DepartmentName   string
	CatalogKind      string
	CatalogURL       string
	CurriculumPDFURL string
}

var Areas = []Area{
	{
		Name: "Asian American Studies", Abbreviation: "AAS", DepartmentName: "Asian American Studies",
		CatalogKind: "concentration", CatalogURL: "https://catalog.williams.edu/aas/",
		CurriculumPDFURL: "https://catalog.williams.edu/pdf/aas.pdf",
	},
	{
		Name: "Coastal and Ocean Studies", Abbreviation: "CAOS", DepartmentName: "Williams-Mystic Program",
		CatalogKind: "concentration", CatalogURL: "https://catalog.williams.edu/caos/",
		CurriculumPDFURL: "https://catalog.williams.edu/pdf/caos.pdf",
	},
	{
		Name: "East Asian Languages and Cultures", Abbreviation: "EALC", DepartmentName: "Asian Languages,Lit&Culture",
		CatalogKind: "major", CatalogURL: "https://catalog.williams.edu/ealc/",
		CurriculumPDFURL: "https://catalog.williams.edu/pdf/ealc.pdf",
	},
}

type Options struct {
	Apply bool
	Areas []Area
}

type Result struct {
	Created           []string
	Existing          []string
	MissingDepartment []string
}

func Run(db *gorm.DB, log *zap.SugaredLogger, opts Options) (*Result, error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	areas := opts.Areas
	if areas == nil {
		areas = Areas
	}
	res := &Result{}

	for _, a := range areas {
		// Unscoped: abbrev and name are unique even across soft-deleted rows
		var existing int
		if err := db.Unscoped().Model(&models.AreaOfStudy{}).
			Where("abbrev = ? OR name = ?", a.Abbreviation, a.Name).
			Count(&existing).Error; err != nil {
			return res, err
		}
		if existing > 0 {
			log.Infof("%s: area already exists, skipping", a.Abbreviation)
			res.Existing = append(res.Existing, a.Abbreviation)
			continue
		}

		var depts []models.Department
		if err := db.Where("name = ?", a.DepartmentName).Find(&depts).Error; err != nil {
			return res, err
		}
		if len(depts) != 1 {
			log.Warnf("%s: found %d departments named %q, skipping", a.Abbreviation, len(depts), a.DepartmentName)
			res.MissingDepartment = append(res.MissingDepartment, a.Abbreviation)
			continue
		}
		deptID := depts[0].ID

		if !opts.Apply {
			log.Infof("%s: would add %q in department %d %q (dry run)", a.Abbreviation, a.Name, deptID, a.DepartmentName)
			res.Created = append(res.Created, a.Abbreviation)
			continue
		}

		area := &models.AreaOfStudy{
			Name:             a.Name,
			Abbreviation:     a.Abbreviation,
			CatalogKind:      a.CatalogKind,
			CatalogURL:       a.CatalogURL,
			CurriculumPDFURL: a.CurriculumPDFURL,
			DepartmentID:     &deptID,
		}
		if err := db.Create(area).Error; err != nil {
			return res, err
		}
		log.Infof("%s: added %q (id %d) in department %d %q", a.Abbreviation, a.Name, area.ID, deptID, a.DepartmentName)
		res.Created = append(res.Created, a.Abbreviation)
	}
	return res, nil
}
