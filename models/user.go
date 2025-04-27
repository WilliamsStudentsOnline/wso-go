package models

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/ldap"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	ldap_lib "github.com/go-ldap/ldap/v3"
)

// User Model
type UserModel struct {
	*BaseModel
}

func NewUserModel(db *gorm.DB, log *zap.SugaredLogger) *UserModel {
	return &UserModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllUsersOptions struct {
	Start *string `json:"start" form:"start"`
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: dorm (with dorm room), tags, department, and office
	Preload []string `json:"preload" form:"preload[]"`
}

// Preload specifically allowed parts if requested
func (o *GetAllUsersOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "dorm") {
		db = db.Preload("DormRoom").Preload("DormRoom.Dorm")
	}
	if lib.StringsContains(o.Preload, "tags") {
		db = db.Preload("Tags")
	}
	if lib.StringsContains(o.Preload, "department") {
		db = db.Preload("Department")
	}
	if lib.StringsContains(o.Preload, "office") {
		db = db.Preload("Office")
	}

	return db
}

func (o *GetAllUsersOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("users.name ASC", true)
}

func (o *GetAllUsersOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Start != nil {
		db = db.Where("users.name > ?", *o.Start)
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}

	return db
}

func (o *GetAllUsersOptions) Run(db *gorm.DB) *gorm.DB {
	return o.Paginate(o.Preloader(db))
}

func (m *UserModel) GetAllUsers(u *[]*User, opts Options) (err error) {
	db := m.DB
	db = m.scopeVisible(db)
	db = m.scopeAtWilliams(db)
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(u).Error
	return
}

func (m *UserModel) CountAllUsers() (count int, err error) {
	db := m.DB.Model(&User{})
	db = m.scopeVisible(db)
	db = m.scopeAtWilliams(db)
	err = db.Count(&count).Error
	return
}

func (m *UserModel) GetAllUsersByType(u *[]User, userType string) (err error) {
	err = m.DB.Where("users.type = ?", userType).Find(u).Error
	return
}

// Returns only users currently at williams (excluding those taking gap year / study away / ...)
func (m *UserModel) GetAtWilliamsUsersByType(u *[]User, userType string) (err error) {
	err = m.DB.Where("users.type = ?", userType).Where("users.at_williams = true").Find(u).Error
	return
}

// Finds students who are on leave. They are marked as type `alum` in our database
func (m *UserModel) GetStudentsOnLeave(u *[]User) (err error) {
	seniorYear := (&StudentModel{}).SeniorYear()
	err = m.DB.Where("users.type = 'alum'").Where("users.class_year >= ?", seniorYear).Find(u).Error
	return
}

func (m *UserModel) GetUserByID(id uint, u *User) (err error) {
	err = m.DB.Where(NewUserWithID(id)).
		Preload("DormRoom").Preload("DormRoom.Dorm").
		Preload("Tags").
		Preload("Department").
		Preload("Office").
		First(u).Error
	return
}

func (m *UserModel) GetUserByUnixID(unix string, u *User) (err error) {
	err = m.DB.Where(&User{UnixID: unix}).
		Preload("DormRoom").Preload("DormRoom.Dorm").
		Preload("Tags").
		Preload("Department").
		Preload("Office").
		First(u).Error
	return
}

func (m *UserModel) DoesUserExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&User{}).Scopes(m.scopeVisible, m.scopeAtWilliams).Where("users.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// Update the user. Only allow specific keys to be passed
func (m *UserModel) UpdateUser(user *User) (err error) {
	// Generate search fields here.
	searchFields, err := m.generateSearchFields(user.ID)
	if err != nil {
		return
	}
	user.SearchFields = searchFields

	err = m.DB.Save(user).Error
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
				return lib.ErrorInvalidUserTag
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

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return
	}

	// Update search fields
	err = m.PopulateSearchFields(user.ID)
	return
}

// Populates the search field. This is an expensive function, so call is sparingly.
func (m *UserModel) PopulateSearchFields(id uint) (err error) {
	searchFields, err := m.generateSearchFields(id)
	if err != nil {
		return
	}

	err = m.DB.Model(NewUserWithID(id)).Update("search_fields", searchFields).Error
	return
}

// Generates the search field, but does not change the database.
func (m *UserModel) generateSearchFields(id uint) (searchFields string, err error) {
	var user User
	err = m.DB.Preload("Tags").
		Preload("DormRoom").
		Preload("DormRoom.Dorm").
		Preload("Office").
		First(&user, id).Error
	if err != nil {
		return
	}

	searchFields = user.GenerateSearchFields()
	return
}

// Generates the search field, but does not change the database. This works by looking up associated elements of the
// user and using those as context, rather than preloading them via the user. (Use case is generating fields of a user
// that doesn't exist yet.)
func (m *UserModel) generateSearchFieldsByUser(user User) (searchFields string, err error) {
	// Populate required preloads
	if len(user.Tags) == 0 {
		var tags []*Tag
		err = m.DB.Model(&user).Related(&tags, "Tags").Error
		if err != nil {
			return
		}
		user.Tags = tags
	}

	if len(user.Tags) == 0 {
		var tags []*Tag
		err = m.DB.Model(&user).Related(&tags, "Tags").Error
		if err != nil {
			return
		}
		user.Tags = tags
	}

	if user.DormRoomID != nil {
		if user.DormRoom == nil {
			var dormRoom DormRoom
			err = m.DB.Preload("Dorm").First(&dormRoom, user.DormRoomID).Error
			if err != nil {
				return
			}
			user.DormRoom = &dormRoom
		}
		if user.DormRoom.Dorm == nil {
			var dorm Dorm
			err = m.DB.First(&dorm, user.DormRoom.DormID).Error
			if err != nil {
				return
			}
			user.DormRoom.Dorm = &dorm
		}
	}

	if user.OfficeID != nil && user.Office == nil {
		var office Office
		err = m.DB.First(&office, user.OfficeID).Error
		if err != nil {
			return
		}
		user.Office = &office
	}

	userPtr := &user
	searchFields = userPtr.GenerateSearchFields()
	return
}

func (m *UserModel) updateUserUnsafe(dbUser *User, toUser *User) (err error) {
	if toUser.Type != dbUser.Type {
		m.log.Infof("Changing type of user %s from %s to %s", dbUser.UnixID, dbUser.Type, toUser.Type)
		dbUser.Type = toUser.Type
	}

	// 100% Could do this is a less verbose way, but this way is much more secure
	dbUser.Name = toUser.Name
	// Ye Shu note Feb 2024: all home, phone, and dorm info are absent in ODIR LDAP, hence we do not overwrite them in the DB.
	// dbUser.CellPhone = toUser.CellPhone
	dbUser.CampusPhoneExt = toUser.CampusPhoneExt
	dbUser.WilliamsEmail = toUser.WilliamsEmail
	dbUser.Title = toUser.Title
	dbUser.Visible = toUser.Visible
	dbUser.ClassYear = toUser.ClassYear
	dbUser.DepartmentID = toUser.DepartmentID
	dbUser.Department = toUser.Department
	// dbUser.HomeTown = toUser.HomeTown
	// dbUser.HomeZip = toUser.HomeZip
	// dbUser.HomePhone = toUser.HomePhone
	// dbUser.HomeState = toUser.HomeState
	// dbUser.HomeCountry = toUser.HomeCountry
	dbUser.Major = toUser.Major
	dbUser.SUBox = toUser.SUBox
	dbUser.OfficeID = toUser.OfficeID
	dbUser.Office = toUser.Office
	// dbUser.DormRoomID = toUser.DormRoomID
	// dbUser.DormRoom = toUser.DormRoom
	dbUser.AtWilliams = toUser.AtWilliams
	dbUser.WilliamsID = toUser.WilliamsID

	// Don't remove entry, but can update it
	if toUser.Entry != nil {
		dbUser.Entry = toUser.Entry
	}

	// Update the search fields
	dbUser.SearchFields, err = m.generateSearchFieldsByUser(*dbUser)
	if err != nil {
		return
	}

	// Update the user
	err = m.DB.Save(dbUser).Error
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
	rows, err := m.DB.Model(&User{}).Where("users.type = ?", UserTypeStudent).Rows() // (*sql.Rows, error)
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
	odirLdap := ldap.NewODIRLDAP()
	adLdap := ldap.NewADLDAP()

	// We need the Willy LDAP credentials for this
	err := config.Secrets.RequireLDAPAuth()
	if err != nil {
		return nil, err
	}

	// Connect & bind the Willy LDAP
	err = odirLdap.ConnectWithBind(config.Secrets.ODIRLDAPDN, config.Secrets.ODIRLDAPPassword)
	if err != nil {
		return nil, err
	}
	defer odirLdap.Close()

	m.log.Info("Start Williams LDAP each")
	// Get all matching users from ODIR LDAP
	userEntries, err := odirLdap.Each("cn", unixSearch)
	if err != nil {
		return nil, err
	}
	m.log.Info("End ODIR LDAP each. found: ", len(userEntries))

	// Connect the AD LDAP
	err = adLdap.ConnectWithBind(config.Secrets.ADLDAPDn, config.Secrets.ADLDAPPassword)
	if err != nil {
		return nil, err
	}
	defer adLdap.Close()

	// This is our result
	var users []*User

	// Go through every returned entry from LDAP
	for _, entry := range userEntries {
		user := &User{
			// We get last attribute value here due to OIT sometimes reassigning unixes and forgetting to
			// remove the old unix.
			UnixID:        getLDAPLastAttributeValue(entry, "cn"),
			Name:          entry.GetAttributeValue("displayName"),
			WilliamsEmail: entry.GetAttributeValue("mail"),
			// TODO: also support "wmsSuppressPhoto"
			Visible: lib.BoolToPtr(!parseBool(entry.GetAttributeValue("wmsSuppressAll"))),
			// User is in Ldap, therefore is at williams.
			// Explicitly set this to handle alums who return as fac/staff
			AtWilliams: lib.BoolToPtr(true),
			// By default we set everyone as unknown, as there are many edge case affiliations.
			Type: UserTypeUnknown,
		}

		// Get the AD info of the user
		adUser, adldapError := adLdap.Get("cn", user.UnixID)
		if adldapError != nil {
			m.log.Error("adldap error getting ", user.UnixID, ": ", adldapError)
		}
		// If we didn't find anything in search, just continue.
		if adUser == nil {
			m.log.Info("user not found in LDAP: ", user.UnixID)
			continue
		}

		// Use the AD user groups to determine the user type
		// Ye Shu note Feb 2024: AD groups are less accurate than ODIR LDAP wmsAffiliation attributes
		// TODO: parse the wmsAffiliation attribute in ODIR LDAP instead of using AD groups
		memberGroups := adUser.GetAttributeValues("memberOf")
		ua := lib.NewUserAssociation(memberGroups)
		adUserDN := strings.ToLower(adUser.DN)

		// Add Williams ID Number
		if adUser.GetAttributeValue("employeeID") != "" {
			user.WilliamsID = adUser.GetAttributeValue("employeeID")
		}

		/*
			Order of types:
			1. unknown
			2. faculty (if group has faculty)
			3. staff (if group has staff)
			4. student (if group has student)
			5. alum (if qualifies for student but class year is after this year AND is not in LDAP servers)
		*/

		if ua.IsFaculty() {
			user.Type = UserTypeProfessor
		} else if ua.IsStaff() {
			user.Type = UserTypeStaff
			if ua.IsStudent() {
				user.Type = UserTypeStudent // trying to fix bug reported by Asa Shepard '27 where some students are reported as staff
			}
		} else if ua.IsStudent() {
			user.Type = UserTypeStudent
		}

		// If the user was a student and is now unknown, they're probably off-cycle alum
		if strings.Contains(adUserDN, "ou=student") && user.Type == UserTypeUnknown {
			user.Type = UserTypeAlum
		}

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
				// We will need to change this in 2045.
				if year < 50 || user.IsStudent() {
					year += 2000
				} else {
					year += 1900
				}
				user.ClassYear = &year
			} else {
				// If the class contains letters (grad students), put no class year, as
				// they're not a normal student

				// TODO: handle special class years (e.g. G1/G2/GE/SV, respectively meaning Grad Art year 1/2, CDE, Bennington College/MCLA visiting students)
				// perhaps by adding a new field to the user schema?

				user.ClassYear = nil
			}
		}

		// If user is not visible, skip parsing details for that user
		if !*user.Visible {
			users = append(users, user)
			continue
		}

		// Personal info parsing
		// Ye Shu note Feb 2024: all home and cell phone info are absent in ODIR LDAP
		// user.HomeCountry = parseStrToPtr(entry.GetAttributeValue("wmsHomeCountry"))
		// user.HomeState = parseStrToPtr(entry.GetAttributeValue("wmsHomeState"))
		// user.HomeTown = parseStrToPtr(entry.GetAttributeValue("wmsHomeCity"))
		// user.HomeZip = parseStrToPtr(entry.GetAttributeValue("wmsHomePostal"))
		// user.CellPhone = parseStrToPtr(entry.GetAttributeValue("wmsCellPhone"))
		user.Title = parseStrToPtr(entry.GetAttributeValue("title"))

		// campus phone extension: parse from whole phone number
		campusPhone := entry.GetAttributeValue("telephoneNumber")
		if len(campusPhone) >= 4 {
			campusPhone = campusPhone[len(campusPhone)-4:]
			user.CampusPhoneExt = &campusPhone
		}

		// Student personal parsing
		if user.IsStudent() || user.IsAlum() {
			// Ye Shu note Feb 2024: all dorm info are absent in ODIR LDAP
			dormName := ""
			// dormName := entry.GetAttributeValue("wmsDormAddr1")

			if !skipDorm && dormName != "" {
				var dorm Dorm
				err = m.DB.Where(&Dorm{
					Name: dormName,
				}).First(&dorm).Error

				if err != nil {
					if gorm.IsRecordNotFoundError(err) {
						// If user is off-campus at an unknown location, create/use that dorm
						m.log.With("unixID", user.UnixID).Warnf("Encountered unknown dorm: %s. Creating it to resolve", dormName)
						err = m.DB.Where(&Dorm{
							Name: dormName,
						}).FirstOrCreate(&dorm).Error
						if err != nil {
							return nil, err
						}
					} else if err != nil {
						return nil, err
					}
				}

				roomNum := entry.GetAttributeValue("wmsDormAddr2")
				var dormRoom DormRoom
				// If no room num (likely off campus, ARTH Grad, CDE), assign them to room 0.
				if roomNum == "" {
					err = m.DB.Where(&DormRoom{
						DormID: dorm.ID,
						Number: "0",
					}).FirstOrCreate(&dormRoom).Error
					if err != nil {
						return nil, err
					}
				} else {
					err = m.DB.Where(&DormRoom{
						DormID: dorm.ID,
						Number: roomNum,
					}).FirstOrCreate(&dormRoom).Error
					if err != nil {
						return nil, err
					}
				}

				user.DormRoomID = &dormRoom.ID
				user.DormRoom = &dormRoom

				// Get entry
				student := user.Student()
				if entry.GetAttributeValue("wmsDormAddr3") != "" && (student.Prefrosh() || student.Frosh()) {
					user.Entry = parseStrToPtr(entry.GetAttributeValue("wmsDormAddr3"))
				}
			} else {
				user.DormRoom = nil
				user.DormRoomID = nil
			}

			// Ye Shu note Feb 2024: for grad students, the field oktaRoomNumber may also be set
			suBox := entry.GetAttributeValue("oktaBuildingName")
			if entry.GetAttributeValue("oktaRoomNumber") != "" {
				suBox = suBox + ", " + entry.GetAttributeValue("oktaRoomNumber")
			}
			user.SUBox = parseStrToPtr(suBox)
		} else if user.IsProfessor() || user.IsStaff() {
			// Office Number: if oktaRoomNumber is empty, ignore it
			number := entry.GetAttributeValue("oktaBuildingName")
			if entry.GetAttributeValue("oktaRoomNumber") != "" {
				number = number + " " + entry.GetAttributeValue("oktaRoomNumber")
			}

			var office Office
			err = m.DB.Where(&Office{
				Number: number,
			}).FirstOrCreate(&office).Error
			if err != nil {
				return nil, err
			}
			user.Office = &office
			user.OfficeID = &office.ID

			departmentName := entry.GetAttributeValue("department")
			if departmentName != "" {
				// remove " Department" from the end of the department name
				// note: this only applies to faculties
				// staff members' department name does not have " Department" at the end
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

		users = append(users, user)
	}

	return users, nil
}

func (m *UserModel) UpdateAllFromLDAP(cfg *config.Config) error {
	m.log.Info("Begin user update from LDAP")
	unixesAtWilliams := make(map[string]bool)

	m.log.Info("Start LDAP lookup")
	toUsers, err := m.LDAPLookup("*", cfg)
	if err != nil {
		return err
	}
	m.log.Info("End LDAP lookup")

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

		// Intercept people from Taiwan and ensure they are ROC not PRC
		if toUser.HomeCountry != nil && *toUser.HomeCountry == "Taiwan, Province of China" {
			toUser.HomeCountry = lib.StrToPtr("Taiwan")
		}

		// Either update or create the LDAP user
		if !query.RecordNotFound() {
			// If the user exists
			// TODO: Figure out a way to batch these calls
			err = m.updateUserUnsafe(dbUser, toUser)
			if err != nil {
				m.log.With(zap.Error(err)).Errorf("Could not save user %s with %#+v", toUser.UnixID, toUser)
			}
		} else {
			// If the user does not exist
			m.log.Infof("Creating new user %s type %s", toUser.UnixID, toUser.Type)
			err = m.DB.Create(toUser).Error
			if err != nil {
				m.log.With(zap.Error(err)).Errorf("Could not create user %s with %#+v", toUser.UnixID, toUser)
			}
			// Update user with search field
			err = m.PopulateSearchFields(toUser.ID)
			if err != nil {
				m.log.With(zap.Error(err)).Errorf("Could not update user (%s) search field", toUser.UnixID)
			}
		}
	}

	// Get database users
	var users []User
	if err = m.DB.Find(&users).Error; err != nil {
		return err
	}

	for _, user := range users {
		// If user's unix was not found in the LDAP
		if _, ok := unixesAtWilliams[user.UnixID]; !ok {
			// Update not in LDAP:
			// TODO: Figure out a way to batch these calls
			err = m.updateNotInLDAP(&user)
			if err != nil {
				m.log.With(zap.Error(err)).Errorf("Could not update (not in LDAP) user %s with %#+v", user.UnixID, user)
			}
		}
	}

	m.log.Info("Finished user update from LDAP")
	return nil
}

func (m *UserModel) UpdateFactrakSurveyDeficit(user *User) error {
	if !user.IsStudent() {
		return errors.New("user must be student")
	}
	return NewStudentModel(m.DB, m.log).UpdateFactrakSurveyDeficit(user)
}

func (*UserModel) scopeVisible(db *gorm.DB) *gorm.DB {
	return db.Where("users.visible = ?", true)
}

func (*UserModel) scopeAtWilliams(db *gorm.DB) *gorm.DB {
	return db.Where("users.at_williams = ?", true)
}

func (*UserModel) scopeAlphabetical(db *gorm.DB) *gorm.DB {
	return db.Order("users.name DESC")
}

// Updates users that are not found in LDAP anymore (alumni usually) by searching for them on NDS,
// finding their type, and then updating the user.
func (m *UserModel) updateNotInLDAP(user *User) error {
	// TODO: remove at_williams and replace with deleted_at
	user.AtWilliams = lib.BoolToPtr(false)

	// If user is student and the class year is this year or less, set user type to be alum
	if user.Type == UserTypeStudent && user.ClassYear != nil && *user.ClassYear <= time.Now().Year() {
		m.log.Infof("Changing type of missing user %s from %s to alum", user.UnixID, user.Type)
		user.Type = UserTypeAlum
	}

	user.DormRoom = nil
	user.DormRoomID = nil
	user.Office = nil
	user.OfficeID = nil
	// We can do this without loading associations as we know that we deleted all of them.
	user.SearchFields = user.GenerateSearchFields()

	// truncate to 250 characters, since the column is only 255 characters long
	if len(user.SearchFields) > 250 {
		m.log.Infof("Truncating long search fields for user %s", user.UnixID)
		user.SearchFields = user.SearchFields[:250]
	}

	return m.DB.Save(user).Error
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

func getLDAPLastAttributeValue(entry *ldap_lib.Entry, attribute string) string {
	values := entry.GetAttributeValues(attribute)
	if len(values) == 0 {
		return ""
	}
	return values[len(values)-1]
}
