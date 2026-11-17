package catalog_update

// Parsing is lenient since the HTML can change: missing values are left empty and never blank stored ones

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	CatalogSiteURL     = "https://catalog.williams.edu"
	catalogUnitIndex   = "/courses-and-programs/"
	unitIndexFile      = "index.html"
	unitFetchPause     = 300 * time.Millisecond // be gentle with the catalog site
	catalogUserAgent   = "wso-catalog-update (https://wso.williams.edu)"
	maxAreaAbbrevChars = 4 // areas_of_study.abbrev is varchar(4)
)

type CatalogUnit struct {
	Code             string
	Name             string
	Kind             string
	CatalogURL       string
	CurriculumPDFURL string
	WebsiteURL       string
	Description      string
	Attributes       []CatalogAttribute
}

type CatalogAttribute struct {
	Code string
	Name string
}

var distributionAttributes = map[string]string{
	"DIV_D1":  "Division I",
	"DIV_D2":  "Division II",
	"DIV_D3":  "Division III",
	"DPE_DPE": "Difference, Power, and Equity",
	"QFR_QFR": "Quantitative/Formal Reasoning",
	"WAC_WAC": "Writing Skills",
}

// Attributes courses carry that no unit page lists, named from older catalog pages
var UnlistedAttributes = map[string]string{
	"AFR_AFRELEC":     "AFR Interdepartmental Electives",
	"ARTH_ARTHPST16C": "ARTH Post-1600",
	"ASAM_ASAMREL":    "ASAM Related Courses",
	"DIV_D5":          "Winter Study Independent Study",
	"DIV_D6":          "No Divisional Credit",
	"ENGL_ENGL1900":   "ENGL Post-1900 Courses",
	"ENVI_ENVICORE":   "ENVI Core Courses",
	"ENVI_ENVIELAB":   "ENVI Electives Lab",
	"ENVI_ENVIHUM":    "ENVI Humanities, Arts + Social Science Electives",
	"ENVI_ENVINATW":   "ENVI Natural World Electives",
	"ENVI_ENVIPOL":    "ENVI Environmental Policy",
	"ENVI_ENVISENSEM": "ENVI Senior Seminar",
	"EVST_EVSTCORE":   "EVST Core Courses",
	"EVST_EVSTCULHUM": "EVST Culture/Humanities",
	"EVST_EVSTENVS":   "EVST Environmental Science",
	"EVST_EVSTSENSEM": "EVST Senior Seminar",
	"EVST_EVSTSOCSCI": "EVST Social Science/Policy",
	"FMST_FMSTCORE":   "FMST Core Courses",
	"FMST_FMSTREL":    "FMST Related Courses",
	"JLST_JLSTTHRY":   "JLST Theories of Justice/Law",
	"LING_LING":       "Linguistics",
	"LOG_LOG":         "On the Log",
	"MAST_MASTELEC":   "MAST Interdepartmental Electives",
	"MAST_MASTSENSEM": "MAST Senior Seminar",
	"MUS_MUSW":        "MUS World Music/Ethnomusicology",
	"POEC_POECAMER":   "POEC U.S. Political Economy + Public Policy Courses",
	"POEC_POECCOMP":   "POEC Comparative POEC/Public Policy Courses",
	"POEC_POECINTL":   "POEC International Political Economy Courses",
	"SLFX_SLFX":       "SLFX Winter Study Self-Expression",
	"STUX_STUX":       "STUX Winter Study Student Exploration",
	"TRVL_TRVL":       "TRVL Winter Study Travel",
	"WELL_WELL":       "WELL Winter Study Wellness",
	"WGSS_WGSSDIV":    "WGSS Racial Sexual + Cultural Diversity Courses",
	"WGSS_WGSSTHRY":   "WGSS Theory Courses",
}

var unitHref = regexp.MustCompile(`^(?:https://catalog\.williams\.edu)?/([a-z0-9-]+)/?$`)

func ParseUnitIndex(html []byte) ([]CatalogUnit, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}
	var units []CatalogUnit
	seen := map[string]bool{}
	doc.Find("ul.dept-nav a[href]").Each(func(_ int, a *goquery.Selection) {
		m := unitHref.FindStringSubmatch(strings.TrimSpace(a.AttrOr("href", "")))
		if m == nil || seen[m[1]] {
			return
		}
		kind := normalizeKind(a.Find("span").Text())
		a.Find("span").Remove()
		name := strings.Join(strings.Fields(a.Text()), " ")
		if name == "" {
			return
		}
		seen[m[1]] = true
		units = append(units, CatalogUnit{
			Code:       m[1],
			Name:       name,
			Kind:       kind,
			CatalogURL: fmt.Sprintf("%s/%s/", CatalogSiteURL, m[1]),
		})
	})
	if len(units) == 0 {
		return nil, fmt.Errorf("no units found on the catalog index page")
	}
	return units, nil
}

func normalizeKind(s string) string {
	s = strings.ToLower(s)
	var kinds []string
	for _, k := range []string{"major", "concentration"} {
		if strings.Contains(s, k) {
			kinds = append(kinds, k)
		}
	}
	return strings.Join(kinds, ",")
}

func ParseUnitPage(html []byte, unit *CatalogUnit) error {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return err
	}
	unit.CurriculumPDFURL = strings.TrimSpace(doc.Find("a.cc-pdf").First().AttrOr("href", ""))
	doc.Find("a[href]").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		if strings.HasPrefix(strings.TrimSpace(a.Text()), "More information can be found") {
			unit.WebsiteURL = strings.TrimSpace(a.AttrOr("href", ""))
			return false
		}
		return true
	})
	unit.Description = parseUnitDescription(doc)
	// Each attribute is a card: a title, then links to course lists filtered by ?sbattr=CODE
	unit.Attributes = nil
	doc.Find(".card-title").Each(func(_ int, title *goquery.Selection) {
		name := strings.Join(strings.Fields(title.Text()), " ")
		href := title.Parent().Find("a[href*='sbattr=']").First().AttrOr("href", "")
		code := queryParam(href, "sbattr")
		if name != "" && code != "" && code != "NA" {
			unit.Attributes = append(unit.Attributes, CatalogAttribute{Code: code, Name: name})
		}
	})
	return nil
}

func parseUnitDescription(doc *goquery.Document) string {
	var blocks []string
	doc.Find("article.dept-desc .post-content").First().Children().Each(func(_ int, el *goquery.Selection) {
		var lines []string
		if el.Is("ul, ol") {
			el.Find("li").Each(func(_ int, li *goquery.Selection) {
				lines = append(lines, strings.Join(strings.Fields(li.Text()), " "))
			})
		} else {
			lines = append(lines, strings.Join(strings.Fields(el.Text()), " "))
		}
		if text := strings.TrimSpace(strings.Join(lines, "\n")); text != "" {
			blocks = append(blocks, text)
		}
	})
	return strings.Join(blocks, "\n\n")
}

func queryParam(href, key string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

func LoadCatalogUnits(pagesDir, archiveDir string, log *zap.SugaredLogger) ([]CatalogUnit, error) {
	load := func(name, url string) ([]byte, error) {
		if pagesDir != "" {
			return os.ReadFile(filepath.Join(pagesDir, name))
		}
		body, err := fetchPage(url)
		if err == nil && archiveDir != "" {
			if err := os.MkdirAll(archiveDir, 0755); err == nil {
				_ = os.WriteFile(filepath.Join(archiveDir, name), body, 0644)
			}
		}
		time.Sleep(unitFetchPause)
		return body, err
	}

	index, err := load(unitIndexFile, CatalogSiteURL+catalogUnitIndex)
	if err != nil {
		return nil, fmt.Errorf("load catalog unit index: %w", err)
	}
	units, err := ParseUnitIndex(index)
	if err != nil {
		return nil, err
	}
	for i := range units {
		page, err := load(units[i].Code+".html", units[i].CatalogURL)
		if err == nil {
			err = ParseUnitPage(page, &units[i])
		}
		if err != nil {
			log.Warnf("catalog unit %s: %v", units[i].Code, err)
		}
	}
	return units, nil
}

func fetchPage(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", catalogUserAgent)
	res, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, res.Status)
	}
	return io.ReadAll(res.Body)
}

type UnitUpdateStats struct {
	AreasUpdated int
	Unmatched    []string
	Attributes   int
}

func UpdateAreasAndAttributes(db *gorm.DB, units []CatalogUnit, log *zap.SugaredLogger) (UnitUpdateStats, error) {
	var stats UnitUpdateStats
	var areas []models.AreaOfStudy
	if err := db.Find(&areas).Error; err != nil {
		return stats, err
	}
	areaByAbbrev := map[string]*models.AreaOfStudy{}
	for i := range areas {
		areaByAbbrev[strings.ToUpper(areas[i].Abbreviation)] = &areas[i]
	}
	for abbrev, catalogCode := range models.RenamedSubjects {
		if area := areaByAbbrev[abbrev]; area != nil && areaByAbbrev[catalogCode] == nil {
			areaByAbbrev[catalogCode] = area
		}
	}

	tx := db.Begin()
	if tx.Error != nil {
		return stats, tx.Error
	}
	defer tx.RollbackUnlessCommitted()

	for _, u := range units {
		area := areaByAbbrev[strings.ToUpper(u.Code)]
		if area == nil || len(u.Code) > maxAreaAbbrevChars {
			stats.Unmatched = append(stats.Unmatched, u.Code)
		} else {
			updates := map[string]interface{}{}
			setIfPresent(updates, "catalog_kind", u.Kind)
			setIfPresent(updates, "catalog_url", u.CatalogURL)
			setIfPresent(updates, "curriculum_pdf_url", u.CurriculumPDFURL)
			setIfPresent(updates, "website_url", u.WebsiteURL)
			setIfPresent(updates, "description", u.Description)
			if len(updates) > 0 {
				if err := tx.Model(area).Updates(updates).Error; err != nil {
					return stats, err
				}
				stats.AreasUpdated++
			}
		}

		for _, a := range u.Attributes {
			owner := area
			if prefix := strings.SplitN(a.Code, "_", 2)[0]; areaByAbbrev[prefix] != nil {
				owner = areaByAbbrev[prefix] // AFR_AFRCORE belongs to AFR even if listed elsewhere
			}
			if err := upsertAttribute(tx, a.Code, a.Name, owner); err != nil {
				return stats, err
			}
			stats.Attributes++
		}
	}
	for code, name := range distributionAttributes {
		if err := upsertAttribute(tx, code, name, nil); err != nil {
			return stats, err
		}
	}
	for code, name := range UnlistedAttributes {
		var n int
		if err := tx.Model(&models.CourseAttribute{}).Where("code = ?", code).Count(&n).Error; err != nil {
			return stats, err
		}
		if n > 0 {
			continue
		}
		if err := upsertAttribute(tx, code, name, areaByAbbrev[strings.SplitN(code, "_", 2)[0]]); err != nil {
			return stats, err
		}
		stats.Attributes++
	}

	if err := tx.Commit().Error; err != nil {
		return stats, err
	}
	log.Infof("Catalog units: updated %d areas, %d attributes; no matching area for %d units: %s",
		stats.AreasUpdated, stats.Attributes, len(stats.Unmatched), strings.Join(stats.Unmatched, " "))
	return stats, nil
}

func setIfPresent(updates map[string]interface{}, column, value string) {
	if value = strings.TrimSpace(value); value != "" {
		updates[column] = value
	}
}

func upsertAttribute(db *gorm.DB, code, name string, owner *models.AreaOfStudy) error {
	kind := models.CourseAttributeKindProgram
	if _, ok := distributionAttributes[code]; ok {
		kind = models.CourseAttributeKindDistribution
	}
	var attr models.CourseAttribute
	err := db.Where("code = ?", code).First(&attr).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	attr.Code, attr.Name, attr.Kind = code, name, kind
	if owner != nil {
		attr.AreaOfStudyID = &owner.ID
	}
	if attr.ID == 0 {
		return db.Create(&attr).Error
	}
	return db.Save(&attr).Error
}
