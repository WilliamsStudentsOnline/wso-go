package models

import (
	"errors"
	"strconv"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// User Model
type UserModel struct {
	BaseModel
}

func (m *UserModel) GetAllUsers(u *[]User) (err error) {
	err = m.DB.Find(u).Error
	return
}

func (m *UserModel) GetUserByID(id uint, u *User) (err error) {
	err = m.DB.Where(NewUserWithID(id)).First(u).Error
	return
}

// Update the user. Only allow specific keys to be passed
func (m *UserModel) UpdateUser(id uint, update map[string]interface{}) (err error) {
	MapPermit(update, "visible", "dorm_visible", "home_visible", "pronoun", "off_cycle")
	err = m.DB.Model(NewUserWithID(id)).Updates(update).Error
	return
}

func (m *UserModel) FirstOrCreateFromUnixID(unixID string, config *config.Config) (*User, error) {
	user := new(User)
	err := m.DB.Where(&User{
		UnixID: unixID,
	}).First(user).Error
	// If it's a legit error, throw an error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	// If no error, we found the user
	if err == nil {
		return user, nil
	}

	// Create user
	users, err := m.LDAPLookup(unixID, config)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("user not found in LDAP")
	}

	user = users[0]
	err = m.DB.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil

}

func (m *UserModel) Students() ([]*Student, error) {
	rows, err := m.DB.Model(&User{}).Where("type = ?", UserTypeStudent).Rows() // (*sql.Rows, error)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*Student

	for rows.Next() {
		var student Student
		// ScanRows scan a row into student
		err = m.DB.ScanRows(rows, &student)
		if err != nil {
			return nil, err
		}

		students = append(students, &student)
	}
	return students, nil
}

// TODO: AIDAN ENSURE THAT NIL FIELDS DON'T OVERWRITE CURRENT USER FIELDS (AT LEAST FOR ENTRY)
func (m *UserModel) LDAPLookup(unixSearch string, config *config.Config) ([]*User, error) {
	// Initialize LDAPs
	willyLdap := lib.NewWilliamsLDAP()
	ndsLdap := lib.NewNDSLDAP()

	// We need the Willy LDAP credentials for this
	err := config.Secrets.RequireLDAPAuth()
	if err != nil {
		return nil, err
	}

	// Connect & bind the Willy LDAP
	err = willyLdap.ConnectWithBind(config.Secrets.WsoLdapDN, config.Secrets.WsoLdapPassword)
	if err != nil {
		return nil, err
	}
	defer willyLdap.Close()

	// Get all users from Willy LDAP
	userEntries, err := willyLdap.Each("uid", unixSearch)
	if err != nil {
		return nil, err
	}

	// Connect the NDS LDAP
	err = ndsLdap.Connect()
	if err != nil {
		return nil, err
	}
	defer ndsLdap.Close()

	// This is our result
	users := make([]*User, len(userEntries))

	// Go through every returned entry from LDAP
	for idx, entry := range userEntries {
		user := &User{
			UnixID:        entry.GetAttributeValue("uid"),
			Name:          entry.GetAttributeValue("cn"),
			WilliamsEmail: entry.GetAttributeValue("mail"),
			Visible:       parseBool(entry.GetAttributeValue("visible")),
			// User is in Ldap, therefore is at williams.
			//  Explicitly set this to handle alums who return as fac/staff
			AtWilliams: true,
			// By default we set everyone as staff, as there are many edge case affiliations.
			Type: UserTypeStaff,
		}

		// Get the NDS info of the user
		ndsUser, err := ndsLdap.Get("uid", user.UnixID)
		if err != nil {
			return nil, err
		}

		user.Type = userAssociationType(ndsUser)

		// Skip adding dorm
		skipDorm := false

		wmsClass := entry.GetAttributeValue("wmsClass")
		if wmsClass != "" {
			if wmsClass == "SV" {
				// Handles Bennington College/MCLA people
				user.ClassYear = nil
				// We set this to nil for now, as we don't want Bennington College in Dormtrak
				skipDorm = true
			} else if year, err := strconv.Atoi(wmsClass); err == nil {
				// Handle Alum fac/staff from pre-2000.
				// The year cutoff will need to be changed when we get to 2050!
				// If the user is a student, we know it will be in the 21st century
				// We will need to change this in 2100.
				if year < 50 || user.IsStudent() {
					year += 2000
				} else {
					year += 1900
				}
				user.ClassYear = &year
			} else {
				// If the class contains letters (grad students), put no class year, as
				// they're not a normal student
				user.ClassYear = nil
			}
		}

		// If user is not visible, finish parsing here.
		if !user.Visible {
			users[idx] = user
			break
		}

		// Personal info parsing
		user.HomeCountry = parseStrToPtr(entry.GetAttributeValue("wmsHomeCountry"))
		user.HomeState = parseStrToPtr(entry.GetAttributeValue("wmsHomeState"))
		user.HomeTown = parseStrToPtr(entry.GetAttributeValue("wmsHomeCity"))
		user.HomeZip = parseStrToPtr(entry.GetAttributeValue("wmsHomePostal"))
		user.CellPhone = parseStrToPtr(entry.GetAttributeValue("wmsCellPhone"))
		user.Title = parseStrToPtr(entry.GetAttributeValue("title"))

		// Parse ext from whole phone number
		campusPhone := entry.GetAttributeValue("telephoneNumber")
		if len(campusPhone) >= 4 {
			campusPhone = campusPhone[len(campusPhone)-4:]
			user.CampusPhoneExt = &campusPhone
		}

		// Student personal parsing
		if user.IsStudent() || user.IsAlum() {
			dormName := entry.GetAttributeValue("wmsDormAddr1")
			if dormName != "" && !skipDorm {
				var dorm Dorm
				err = m.DB.Where(&Dorm{
					Name: dormName,
				}).First(&dorm).Error

				if err != nil {
					if gorm.IsRecordNotFoundError(err) {
						log.Warnln("Encountered unknown dorm:", dormName)
						user.DormRoomID = nil
						user.DormRoom = nil
					} else if err != nil {
						return nil, err
					}
				} else {
					var dormRoom DormRoom
					err = m.DB.Where(&DormRoom{
						Dorm:   dorm,
						Number: entry.GetAttributeValue("wmsDormAddr2"),
					}).FirstOrCreate(&dormRoom).Error
					if err != nil {
						return nil, err
					}

					user.DormRoomID = &dormRoom.ID
					user.DormRoom = &dormRoom
				}

				student := user.Student()
				if entry.GetAttributeValue("wmsDormAddr3") != "" && (student.Prefrosh() || student.Frosh()) {
					user.Entry = parseStrToPtr(entry.GetAttributeValue("wmsDormAddr3"))
				}
			} else {
				user.DormRoom = nil
				user.DormRoomID = nil
			}

			user.SUBox = parseStrToPtr(entry.GetAttributeValue("wmsCampusAddr1"))
		} else if user.IsProfessor() || user.IsStaff() {
			number := entry.GetAttributeValue("wmsCampusAddr1") + " " + entry.GetAttributeValue("wmsCampusAddr2")

			var office Office
			err = m.DB.Where(&Office{
				Number: number,
			}).FirstOrCreate(&office).Error
			if err != nil {
				return nil, err
			}
			user.Office = &office
			user.OfficeID = &office.ID

			departmentName := entry.GetAttributeValue("ou")
			if departmentName != "" {
				deptIdx := strings.Index(departmentName, " Department")
				if deptIdx == -1 {
					deptIdx = len(departmentName)
				}
				departmentName = departmentName[:deptIdx]

				var dept Department
				err = m.DB.Where(&Department{
					Name: departmentName,
				}).FirstOrCreate(&dept).Error
				if err != nil {
					return nil, err
				}
				user.Department = &dept
				user.DepartmentID = &dept.ID

			}
		}

		users[idx] = user
	}

	return users, nil
}

//func (m *UserModel) UpdateAllFromLDAP

func userAssociationType(ndsUser *ldap.Entry) string {
	ua := lib.NewUserAssociation(
		ndsUser.GetAttributeValue("wmsAffiliation"),
		ndsUser.GetAttributeValue("wmsAffiliationFamily"),
	)

	if ua.IsClassStudent() {
		if ua.IsAlum() {
			return UserTypeAlum
		} else if ua.IsCurrentStudent() {
			return UserTypeStudent
		}
	} else if ua.IsClassEmployee() && ua.IsFaculty() {
		return UserTypeProfessor
	} else if ua.IsStaff() {
		return UserTypeStaff
	}

	return UserTypeStaff
}

func parseStrToPtr(str string) *string {
	if str == "" {
		return nil
	}

	return lib.StrToPtr(str)
}

// Parse boolean from string. Default is true.
func parseBool(str string) bool {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return true
	}

	return b
}
