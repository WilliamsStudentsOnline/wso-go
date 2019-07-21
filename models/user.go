package models

import (
	"log"
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
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

// TODO: AIDAN ENSURE THAT NIL FIELDS DON'T OVERWRITE CURRENT USER FIELDS (AT LEAST FOR ENTRY)
func (m *UserModel) LDAPLookup(unixSearch string, config *config.Config) error {
	// Initialize LDAPs
	willyLdap := lib.NewWilliamsLDAP()
	ndsLdap := lib.NewNDSLDAP()

	// We need the Willy LDAP credentials for this
	err := config.Secrets.RequireLDAPAuth()
	if err != nil {
		return err
	}

	// Connect & bind the Willy LDAP
	err = willyLdap.ConnectWithBind(config.Secrets.WsoLdapDN, config.Secrets.WsoLdapPassword)
	if err != nil {
		return err
	}
	defer willyLdap.Close()

	// Get all users from Willy LDAP
	userEntries, err := willyLdap.Each("uid", unixSearch)
	if err != nil {
		return err
	}

	// Connect the NDS LDAP
	err = ndsLdap.Connect()
	if err != nil {
		return err
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
			return err
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
						log.Println("Encountered unknown dorm:", dormName)
						user.DormRoomID = nil
						user.DormRoom = nil
					} else if err != nil {
						return err
					}
				} else {
					var dormRoom DormRoom
					err = m.DB.Where(&DormRoom{
						Dorm:   dorm,
						Number: entry.GetAttributeValue("wmsDormAddr2"),
					}).FirstOrCreate(&dormRoom).Error
					if err != nil {
						return err
					}

					user.DormRoom = &dormRoom
				}
			}
		}

		log.Printf("User: %+v\n", user)
	}

	return nil
}

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
