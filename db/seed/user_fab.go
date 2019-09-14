// None of these generator generate any relations (e.g. GenerateUser does not generate a dorm or dorm ID).
package seed

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/brianvoe/gofakeit"
)

func GenerateUser() *models.User {
	switch gofakeit.Number(1, 4) {
	case 1:
		return GenerateStudent()
	case 2:
		return GenerateAlum()
	case 3:
		return GenerateProfessor()
	case 4:
		return GenerateStaff()
	}

	return nil
}

func generateBaseUser() *models.User {
	name := gofakeit.Name()
	cellPhone := gofakeit.Phone()
	campusPhoneExt := strconv.Itoa(gofakeit.Number(100, 9999))
	unix := generateUnix()
	title := gofakeit.JobTitle()
	visible := true
	dormVisible := gofakeit.Bool()
	homeTown := gofakeit.City()
	homeZip := gofakeit.Zip()
	homePhone := gofakeit.Phone()
	homeState := gofakeit.State()
	homeCountry := gofakeit.Country()
	homeVisible := gofakeit.Bool()
	suBox := strconv.Itoa(gofakeit.Number(1000, 5000))
	pronoun := gofakeit.RandString([]string{"they/them", "he/him", "she/her/hers"})

	return &models.User{
		Name:           name,
		CellPhone:      &cellPhone,
		CampusPhoneExt: &campusPhoneExt,
		UnixID:         unix,
		WilliamsEmail:  unix + "@williams.edu",
		Title:          &title,
		Visible:        &visible,
		DormVisible:    &dormVisible,
		HomeTown:       &homeTown,
		HomeZip:        &homeZip,
		HomePhone:      &homePhone,
		HomeState:      &homeState,
		HomeCountry:    &homeCountry,
		HomeVisible:    &homeVisible,
		SUBox:          &suBox,
		Admin:          lib.FalsePtr(),
		FactrakAdmin:   lib.FalsePtr(),
		Pronoun:        &pronoun,
		AtWilliams:     lib.TruePtr(),
	}
}

func GenerateStudent() *models.User {
	u := generateBaseUser()
	u.Type = models.UserTypeStudent

	classYear := gofakeit.Number(1975, time.Now().Year()+4)
	major := gofakeit.RandString([]string{"CSCI", "ECON", "PSCI"})
	entry := gofakeit.RandString([]string{"DAMP", "AP3", "MD4", "Willy AB", "Sage CD"})
	hasAcceptedFactrakPolicy := gofakeit.Bool()
	hasAcceptedDormtrakPolicy := gofakeit.Bool()

	u.ClassYear = &classYear
	u.Major = &major
	u.Entry = &entry
	u.HasAcceptedFactrakPolicy = &hasAcceptedFactrakPolicy
	u.HasAcceptedDormtrakPolicy = &hasAcceptedDormtrakPolicy
	u.AtWilliams = lib.TruePtr()

	return u
}

func GenerateAlum() *models.User {
	u := generateBaseUser()
	u.Type = models.UserTypeAlum

	classYear := gofakeit.Number(1975, time.Now().Year()-4)
	major := gofakeit.RandString([]string{"CSCI", "ECON", "PSCI"})
	entry := gofakeit.RandString([]string{"DAMP", "AP3", "MD4", "Willy AB", "Sage CD"})
	hasAcceptedFactrakPolicy := gofakeit.Bool()
	hasAcceptedDormtrakPolicy := gofakeit.Bool()

	u.ClassYear = &classYear
	u.Major = &major
	u.Entry = &entry
	u.HasAcceptedFactrakPolicy = &hasAcceptedFactrakPolicy
	u.HasAcceptedDormtrakPolicy = &hasAcceptedDormtrakPolicy
	u.AtWilliams = lib.FalsePtr()

	return u
}

func GenerateProfessor() *models.User {
	u := generateBaseUser()
	u.Type = models.UserTypeProfessor

	return u
}

func GenerateStaff() *models.User {
	u := generateBaseUser()
	u.Type = models.UserTypeStaff

	return u
}

func generateUnix() string {
	n := gofakeit.Number(2, 3)
	unix := ""
	for i := 0; i < n; i++ {
		unix += gofakeit.Letter()
	}

	return unix + strconv.Itoa(gofakeit.Number(1, 99))
}

func randomSliceStr(s []string) string {
	return s[rand.Intn(len(s))]
}
