package catalog_update_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Trimmed from catalog.williams.edu (2026-11)
const unitIndexHTML = `<nav class="flex-wrap"><ul class="dept-nav">
<li class="nav-main-item dept-wrapper"><a class="nav-main-link maindept" href="/afr/">Africana Studies <span>(major)</span> </a></li>
<li class="nav-main-item dept-wrapper"><a class="nav-main-link maindept" href="/anso/">Anthropology &amp; Sociology</a>
  <ul><li class="nav-drop-item subdept"><a href="/anth/" class="catalog-link">Anthropology <span>(major)</span> </a></li></ul></li>
<li class="nav-main-item dept-wrapper"><a class="nav-main-link maindept" href="/envi/">Environmental Studies <span>(major &amp; concentration)</span></a></li>
<li class="nav-main-item dept-wrapper"><a class="nav-main-link maindept" href="/dpe/">Difference, Power, &amp; Equity (DPE) Courses</a></li>
</ul></nav>`

const afrPageHTML = `<div class="dept-pdf"><a href="https://catalog.williams.edu/pdf/afr.pdf" class="cc-pdf">Department/Program Curriculum PDF</a></div>
<div class='card'><div class='card-title'><h2 class="h5">AFR Core Electives</h2></div>
  <ul class='card-actions'><li><a class='card-action' href='https://catalog.williams.edu/afr/list/?strm=1271&sbattr=AFR_AFRCORE&subj=NA'>FALL</a></li></ul></div>
<div class='card'><div class='card-title'><h2 class="h5">Unit Course Lists</h2></div>
  <ul class='card-actions'><li><a class='card-action' href='https://catalog.williams.edu/afr/list/?strm=1271&sbattr=NA&subj=AFR'>FALL</a></li></ul></div>
<div class='card'><div class='card-title'><h2 class="h5">AMST Comparative Ethnic Studies</h2></div>
  <ul class='card-actions'><li><a class='card-action' href='https://catalog.williams.edu/afr/list/?strm=1271&sbattr=AMST_AMSTCOMP&subj=NA'>FALL</a></li></ul></div>
<p>More information can be found on the <a href="http://africana.williams.edu/">More information can be found on the Africana Studies site</a>.</p>`

func TestParseUnitIndex(t *testing.T) {
	units, err := ParseUnitIndex([]byte(unitIndexHTML))
	require.NoError(t, err)
	assert := testify.New(t)
	require.Len(t, units, 5)
	assert.Equal(CatalogUnit{Code: "afr", Name: "Africana Studies", Kind: "major", CatalogURL: "https://catalog.williams.edu/afr/"}, units[0])
	assert.Equal("Anthropology & Sociology", units[1].Name)
	assert.Equal("", units[1].Kind)
	assert.Equal("anth", units[2].Code)
	assert.Equal("major,concentration", units[3].Kind)
	assert.Equal("Difference, Power, & Equity (DPE) Courses", units[4].Name)

	_, err = ParseUnitIndex([]byte("<html>redesigned</html>"))
	assert.Error(err) // a layout change shouldn't look like "no units"
}

func TestParseUnitPage(t *testing.T) {
	u := CatalogUnit{Code: "afr"}
	require.NoError(t, ParseUnitPage([]byte(afrPageHTML), &u))
	assert := testify.New(t)
	assert.Equal("https://catalog.williams.edu/pdf/afr.pdf", u.CurriculumPDFURL)
	assert.Equal("http://africana.williams.edu/", u.WebsiteURL)
	assert.Equal([]CatalogAttribute{
		{Code: "AFR_AFRCORE", Name: "AFR Core Electives"},
		{Code: "AMST_AMSTCOMP", Name: "AMST Comparative Ethnic Studies"},
	}, u.Attributes)
}

const chemPageHTML = `<aside><div class='card'><div class='card-title'><h2 class="h5">Unit Course Lists</h2></div></div></aside>
<article id="post-20" class="cf dept-desc post-20 post type-post status-publish format-standard hentry category-uncategorized ">
  <div class="post-content ">
    <p>Through a variety of individual courses and sequential programs, the department provides an opportunity for students to explore the nature and significance of chemistry.</p>
<p>A major in chemistry can be achieved in several ways, preferably beginning in the student&#8217;s first year at Williams. <a href="http://chemistry.williams.edu/" target="_blank" rel="noopener noreferrer">More information can be found on the Chemistry site.</a></p>
<p>&nbsp;</p>
<ul>
<li>Aqua Fitness</li>
<li>Squash</li>
</ul>
  </div><!-- .post-content -->
  <footer class="entry-footer"><a href="http://www.printfriendly.com/print/v2?url=https://catalog.williams.edu/chem/">Print this page</a></footer>
</article>`

func TestParseUnitPageDescription(t *testing.T) {
	u := CatalogUnit{Code: "chem"}
	require.NoError(t, ParseUnitPage([]byte(chemPageHTML), &u))
	testify.New(t).Equal("Through a variety of individual courses and sequential programs, the department provides an opportunity for students to explore the nature and significance of chemistry.\n\n"+
		"A major in chemistry can be achieved in several ways, preferably beginning in the student’s first year at Williams. More information can be found on the Chemistry site.\n\n"+
		"Aqua Fitness\nSquash", u.Description)
}

func TestLoadCatalogUnitsFromSavedPages(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte(unitIndexHTML), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "afr.html"), []byte(afrPageHTML), 0644))
	units, err := LoadCatalogUnits(dir, "", zap.NewNop().Sugar())
	require.NoError(t, err)
	require.Len(t, units, 5)
	testify.New(t).Len(units[0].Attributes, 2)
}

func TestUpdateAreasAndAttributes(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)
	log := zap.NewNop().Sugar()

	dept := models.Department{Name: "Africana Studies"}
	require.NoError(t, db.Create(&dept).Error)
	afr := models.AreaOfStudy{Name: "Africana Studies", Abbreviation: "AFR", DepartmentID: &dept.ID}
	amst := models.AreaOfStudy{Name: "American Studies", Abbreviation: "AMST", DepartmentID: &dept.ID}
	require.NoError(t, db.Create(&afr).Error)
	require.NoError(t, db.Create(&amst).Error)
	wgst := models.AreaOfStudy{Name: "Women's and Gender Studies", Abbreviation: "WGST", DepartmentID: &dept.ID}
	require.NoError(t, db.Create(&wgst).Error)

	u := CatalogUnit{Code: "afr", Name: "Africana Studies", Kind: "major", CatalogURL: "https://catalog.williams.edu/afr/"}
	require.NoError(t, ParseUnitPage([]byte(afrPageHTML), &u))
	units := []CatalogUnit{u,
		{Code: "dpe", Name: "DPE Courses", CatalogURL: "https://catalog.williams.edu/dpe/"},
		{Code: "wgss", Name: "Women's, Gender & Sexuality Studies", Kind: "major", CatalogURL: "https://catalog.williams.edu/wgss/"},
	}

	stats, err := UpdateAreasAndAttributes(db, units, log)
	require.NoError(t, err)
	assert.Equal(2, stats.AreasUpdated)
	assert.Equal([]string{"dpe"}, stats.Unmatched)

	var got models.AreaOfStudy
	require.NoError(t, db.First(&got, afr.ID).Error)
	assert.Equal("major", got.CatalogKind)
	assert.Equal("https://catalog.williams.edu/pdf/afr.pdf", got.CurriculumPDFURL)
	assert.Equal("http://africana.williams.edu/", got.WebsiteURL)

	var gotWGST models.AreaOfStudy
	require.NoError(t, db.First(&gotWGST, wgst.ID).Error)
	assert.Equal("https://catalog.williams.edu/wgss/", gotWGST.CatalogURL)

	var core, comp, div models.CourseAttribute
	require.NoError(t, db.Where("code = ?", "AFR_AFRCORE").First(&core).Error)
	assert.Equal("AFR Core Electives", core.Name)
	assert.Equal(models.CourseAttributeKindProgram, core.Kind)
	require.NoError(t, db.Where("code = ?", "AMST_AMSTCOMP").First(&comp).Error)
	require.NotNil(t, comp.AreaOfStudyID)
	assert.Equal(amst.ID, *comp.AreaOfStudyID) // owned by its code prefix, not the page it was on
	require.NoError(t, db.Where("code = ?", "DIV_D3").First(&div).Error)
	assert.Equal("Division III", div.Name)
	assert.Equal(models.CourseAttributeKindDistribution, div.Kind)

	// A run that parsed less never blanks stored values
	_, err = UpdateAreasAndAttributes(db, []CatalogUnit{{Code: "afr", Name: "Africana Studies"}}, log)
	require.NoError(t, err)
	got = models.AreaOfStudy{}
	require.NoError(t, db.First(&got, afr.ID).Error)
	assert.Equal("major", got.CatalogKind)
	assert.Equal("http://africana.williams.edu/", got.WebsiteURL)

	var count int
	require.NoError(t, db.Model(&models.CourseAttribute{}).Count(&count).Error)
	assert.Equal(8+len(UnlistedAttributes), count)
}
