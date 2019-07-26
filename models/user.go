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
	*BaseModel
}

func NewUserModel(db *gorm.DB) *UserModel {
	return &UserModel{
		BaseModel: NewBaseModel(db),
	}
}

func (m *UserModel) GetAllUsers(u *[]User) (err error) {
	err = m.DB.Scopes(m.scopeVisible, m.scopeAtWilliams).Find(u).Error
	return
}

func (m *UserModel) GetUserByID(id uint, u *User) (err error) {
	err = m.DB.Where(NewUserWithID(id)).Preload("Tags").First(u).Error
	return
}

type UpdateUserParams struct {
	Visible     *bool   `json:"visible"`
	DormVisible *bool   `json:"dormVisible"`
	HomeVisible *bool   `json:"homeVisible"`
	Pronoun     *string `json:"pronoun"`
	OffCycle    *bool   `json:"offCycle"`
}

// Update the user. Only allow specific keys to be passed
func (m *UserModel) UpdateUser(id uint, update *UpdateUserParams) (err error) {
	dbUpdate := map[string]interface{}{
		"visible":      update.Visible,
		"dorm_visible": update.DormVisible,
		"home_visible": update.HomeVisible,
		"pronoun":      update.Pronoun,
		"off_cycle":    update.OffCycle,
	}
	DeleteNilFields(dbUpdate)

	err = m.DB.Model(NewUserWithID(id)).Updates(dbUpdate).Error
	return
}

func (m *UserModel) UpdateUserTags(id uint, tags []string) (err error) {
	// Get user
	user := new(User)
	err = m.DB.First(&user, id).Error
	if err != nil {
		return err
	}

	// Start transaction
	tx := m.DB.Begin()
	if err = tx.Error; err != nil {
		return err
	}

	// Clear previous tags
	err = tx.Model(&user).Association("Tags").Clear().Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Add new tags
	for _, tagName := range tags {
		// Get tag
		tag := new(Tag)
		err = m.DB.Where(&Tag{
			Name: tagName,
		}).First(tag).Error

		// If we don't have the tag, error
		if err != nil {
			tx.Rollback()
			if gorm.IsRecordNotFoundError(err) {
				return errors.New("invalid user tag")
			}
			return err
		}

		// If we do have the tag, add it
		err = tx.Model(&user).Association("Tags").Append(tag).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (m *UserModel) updateUserUnsafe(dbUser *User, toUser *User) error {
	if toUser.Type != dbUser.Type {
		log.Infof("Changing type of user %s from %s to %s", dbUser.UnixID, dbUser.Type, toUser.Type)
		dbUser.Type = toUser.Type
	}

	// 100% Could do this is a less verbose way, but this way is much more secure
	dbUser.Name = toUser.Name
	dbUser.CellPhone = toUser.CellPhone
	dbUser.CampusPhoneExt = toUser.CampusPhoneExt
	dbUser.WilliamsEmail = toUser.WilliamsEmail
	dbUser.Title = toUser.Title
	dbUser.Visible = toUser.Visible
	dbUser.ClassYear = toUser.ClassYear
	dbUser.DepartmentID = toUser.DepartmentID
	dbUser.Department = toUser.Department
	dbUser.HomeTown = toUser.HomeTown
	dbUser.HomeZip = toUser.HomeZip
	dbUser.HomePhone = toUser.HomePhone
	dbUser.HomeState = toUser.HomeState
	dbUser.HomeCountry = toUser.HomeCountry
	dbUser.Major = toUser.Major
	dbUser.SUBox = toUser.SUBox
	dbUser.OfficeID = toUser.OfficeID
	dbUser.Office = toUser.Office
	dbUser.DormRoomID = toUser.DormRoomID
	dbUser.DormRoom = toUser.DormRoom
	dbUser.AtWilliams = toUser.AtWilliams

	// Don't remove entry, but can update it
	if toUser.Entry != nil {
		dbUser.Entry = toUser.Entry
	}

	return m.DB.Save(dbUser).Error
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
	if config.DisableLDAP {
		return nil, errors.New("LDAP is disabled in config")
	}

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

	log.Info("Start Willy LDAP each")
	// Get all users from Willy LDAP
	userEntries, err := willyLdap.Each("uid", unixSearch)
	if err != nil {
		return nil, err
	}
	log.Info("End Willy LDAP each")

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
			// Explicitly set this to handle alums who return as fac/staff
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
						log.Warn("Encountered unknown dorm:", dormName)
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

func (m *UserModel) UpdateAllFromLDAP(cfg *config.Config) error {
	log.Info("Begin user update from LDAP")
	unixesAtWilliams := make(map[string]bool)

	log.Info("Start LDAP lookup")
	toUsers, err := m.LDAPLookup("*", cfg)
	if err != nil {
		return err
	}
	log.Info("End LDAP lookup")

	for _, toUser := range toUsers {
		unixesAtWilliams[toUser.UnixID] = true
		dbUser := new(User)

		query := m.DB.Where(&User{
			UnixID: toUser.UnixID,
		}).First(dbUser)
		// Error if it is a real error
		if query.Error != nil && !query.RecordNotFound() {
			return query.Error
		}

		// Either update or create the LDAP user
		if !query.RecordNotFound() {
			// If the user exists
			// TODO: Figure out a way to batch these calls
			err = m.updateUserUnsafe(dbUser, toUser)
			if err != nil {
				log.WithError(err).Errorf("Could not save user %s with %#+v", toUser.UnixID, toUser)
			}
		} else {
			// If the user does not exist
			log.Infof("Creating new user %s type %s", toUser.UnixID, toUser.Type)
			err = m.DB.Create(toUser).Error
			if err != nil {
				log.WithError(err).Errorf("Could not create user %s with %#+v", toUser.UnixID, toUser)
			}
		}
	}

	// Get database users
	var users []User
	if err = m.DB.Find(&users).Error; err != nil {
		return err
	}

	// Connect to NDS
	ndsLdap := lib.NewNDSLDAP()
	err = ndsLdap.Connect()
	if err != nil {
		return err
	}
	defer ndsLdap.Close()

	for _, user := range users {
		// If user's unix was not found in the LDAP
		if _, ok := unixesAtWilliams[user.UnixID]; !ok {
			// Update not in LDAP:
			// TODO: Figure out a way to batch these calls
			err = m.updateNotInLDAP(&user, ndsLdap)
			if err != nil {
				log.WithError(err).Errorf("Could not update (not in LDAP) user %s with %#+v", user.UnixID, user)
			}
		}
	}

	log.Info("Finished user update from LDAP")
	return nil
}

// TODO: Add this once Factrak is done
func (m *UserModel) UpdateServerDeficit(user *User) error {
	if !user.IsStudent() {
		return errors.New("user must be student")
	}
	return nil
}

func (*UserModel) scopeVisible(db *gorm.DB) *gorm.DB {
	return db.Where("visible = ?", true)
}

func (*UserModel) scopeAtWilliams(db *gorm.DB) *gorm.DB {
	return db.Where("at_williams = ?", true)
}

func (*UserModel) scopeAlphabetical(db *gorm.DB) *gorm.DB {
	return db.Order("name DESC")
}

// Updates users that are not found in LDAP anymore (alumni usually) by searching for them on NDS,
// finding their type, and then updating the user.
func (m *UserModel) updateNotInLDAP(user *User, ndsLdap *lib.LDAP) error {
	user.AtWilliams = false
	entry, err := ndsLdap.Get("uid", user.UnixID)
	if err != nil {
		return err
	}

	// if LDAP returned anything
	if entry != nil {
		assocType := userAssociationType(entry)
		if user.Type != assocType {
			log.Infof("Changing type of missing user %s from %s to %s", user.UnixID, user.Type, assocType)
		}
		user.Type = assocType
	}

	user.DormRoom = nil
	user.DormRoomID = nil
	user.Office = nil
	user.OfficeID = nil

	return m.DB.Save(user).Error
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
