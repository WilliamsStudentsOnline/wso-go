package models

import (
	"sort"
	"strconv"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// EphmatchProfile Model
type EphmatchProfileModel struct {
	*BaseModel
}

func NewEphmatchProfileModel(db *gorm.DB, log *zap.SugaredLogger) *EphmatchProfileModel {
	return &EphmatchProfileModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Get all ephmatch profiles.
func (m *EphmatchProfileModel) GetAllProfiles(p *[]*EphmatchProfile, opts Options) (err error) {
	// Get profiles
	db := m.DB.Model(&EphmatchProfile{}).Preload("User")

	db = m.scopeDefault(db)
	if opts != nil {
		db = opts.Run(db)
	}

	return db.Find(p).Error
}

// Get all ephmatch profiles.
func (m *EphmatchProfileModel) GetAllProfilesNoSelf(p *[]*EphmatchProfile, selfID uint, opts *GetAllProfilesOptions) (err error) {
	// Get profiles
	db := m.DB.Model(&EphmatchProfile{}).Preload("User")

	db = m.scopeDefault(db)
	if opts != nil {
		db = opts.Run(db)
	}

	db = db.Not(EphmatchProfile{UserID: selfID})

	err = db.Find(p).Error
	if err != nil {
		return
	}

	if opts != nil {
		// Do recommended sorting algorithm
		if opts.Sort != nil && *opts.Sort == "recommended" {
			err = m.SortProfilesByLiked(p, selfID, opts)
			if err != nil {
				return
			}
		}

		// Populate liked field
		if lib.StringsContains(opts.Preload, "liked") {
			err = m.PopulateLiked(p, selfID)
			if err != nil {
				return
			}
		}

		// Populate matched field
		if lib.StringsContains(opts.Preload, "matched") {
			err = m.PopulateMatched(p, selfID)
			if err != nil {
				return
			}
		}
	}

	return
}

func (m *EphmatchProfileModel) SortProfilesByLiked(p *[]*EphmatchProfile, selfID uint, opts *GetAllProfilesOptions) error {
	suggestedUsers, err := m.SuggestUsers(selfID)
	if err != nil {
		return err
	}

	usersToIdx := make(map[uint]int, len(suggestedUsers))
	for i, u := range suggestedUsers {
		usersToIdx[u] = i
	}

	// Sort by order if both users are suggested, then by the suggested user if there is only one, and then by
	// the most active user if both users are not suggested
	sort.Slice(*p, func(i, j int) bool {
		iIdx, iOk := usersToIdx[(*p)[i].UserID]
		jIdx, jOk := usersToIdx[(*p)[j].UserID]
		if iOk && jOk {
			return iIdx > jIdx
		} else if iOk && !jOk {
			return true
		} else if !iOk && jOk {
			return false
		} else {
			return (*p)[i].UpdatedAt.After((*p)[j].UpdatedAt)
		}
	})

	// Do manual offset, limit as we cant do it in SQL
	if opts.Offset != nil && *opts.Offset < uint(len(*p))-1 { // Ensure bounds
		*p = (*p)[*opts.Offset:]
	}
	if opts.Limit != nil && *opts.Limit < uint(len(*p)) {
		*p = (*p)[:*opts.Limit]
	}

	return nil
}

// Populate liked field on set of profiles
func (m *EphmatchProfileModel) PopulateLiked(p *[]*EphmatchProfile, selfID uint) (err error) {
	// Get ephmatch likes made by user self (selfID) to populate "liked" field. This is significantly faster
	// than a SQL query by several magnitudes.
	var likes []*EphmatchLike
	lm := NewEphmatchLikeModel(m.DB, m.log)
	err = lm.GetUserLikes(selfID, &likes)
	if err != nil {
		return
	}

	// Create a map of users we liked.
	likedUserMap := make(map[uint]bool)
	for _, like := range likes {
		likedUserMap[like.LikedID] = true
	}

	// If a profile is in the likedUserMap, set it to liked
	for _, profile := range *p {
		if _, ok := likedUserMap[profile.UserID]; ok {
			profile.Liked = lib.TruePtr()
		} else {
			profile.Liked = lib.FalsePtr()
		}
	}

	return
}

// Populate matched field on set of profiles
func (m *EphmatchProfileModel) PopulateMatched(p *[]*EphmatchProfile, selfID uint) (err error) {
	// Get ephmatch matches made by user self (selfID) to populate "matched" field. This is significantly faster
	// than a SQL query by several magnitudes.
	var matches []*EphmatchMatch
	mm := NewEphmatchMatchesModel(m.DB, m.log)
	err = mm.GetRawUserMatches(selfID, &matches)
	if err != nil {
		return
	}

	// Create a map of users we matched with.
	matchedUserMap := make(map[uint]bool)
	for _, match := range matches {
		matchedUserMap[match.MatchedUserID] = true
	}

	// If a profile is in the matchedUserMap, set it to matched
	for _, profile := range *p {
		if _, ok := matchedUserMap[profile.UserID]; ok {
			profile.Matched = lib.TruePtr()
		} else {
			profile.Matched = lib.FalsePtr()
		}
	}

	return
}

type GetAllProfilesOptions struct {
	// Can be new, updated, alphabetical, or recommended
	Sort *string `json:"sort" form:"sort"`

	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: tags, liked, matched
	Preload []string `json:"preload" form:"preload[]"`
	// Note: [liked, matched] preloads are done not in the preload step
}

func (p *GetAllProfilesOptions) Order(db *gorm.DB) *gorm.DB {
	if p.Sort != nil {
		if *p.Sort == "new" {
			return db.Order("ephmatch_profiles.created_at DESC", true)
		} else if *p.Sort == "updated" {
			return db.Order("ephmatch_profiles.updated_at DESC", true)
		} else if *p.Sort == "recommended" {
			return db
		}
	}

	return db.Order("users.name ASC", true)
}

func (p *GetAllProfilesOptions) Paginate(db *gorm.DB) *gorm.DB {
	// Do server-side limit/offset if recommended due to complex sorting algorithm
	if p.Sort != nil && *p.Sort == "recommended" {
		return db
	}

	db = p.Order(db)
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
		if p.Offset != nil {
			db = db.Offset(*p.Offset)
		}
	}

	return db
}

// Preload specifically allowed parts if requested
func (o *GetAllProfilesOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "tags") {
		db = db.Preload("User.Tags")
	}

	return db
}

func (p *GetAllProfilesOptions) Run(db *gorm.DB) *gorm.DB {
	return p.Paginate(p.Preloader(db))
}

func (m *EphmatchProfileModel) CountProfiles() (count int, err error) {
	db := m.scopeDefault(m.DB.Model(&EphmatchProfile{}))
	err = db.Count(&count).Error
	return
}

// Create a new survey.
func (m *EphmatchProfileModel) CreateProfile(p *EphmatchProfile) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}

	err = m.DB.
		Preload("User").
		First(p).Error
	return
}

// Create or update a new profile unscoped about deleted. Updating a deleted profile will make it be undeleted
func (m *EphmatchProfileModel) CreateOrUpdateProfileUnscoped(userID uint, newProfile EphmatchProfile, p *EphmatchProfile) (err error) {

	var tempProf EphmatchProfile
	err = m.DB.Unscoped().Where(EphmatchProfile{UserID: userID}).First(&tempProf).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	// If our profile was deleted, undelete it
	if tempProf.DeletedAt != nil {
		query := m.DB.
			Unscoped().
			Model(&EphmatchProfile{}).
			Where(EphmatchProfile{UserID: userID})
		if newProfile.Description != nil {
			query = query.Update("description", newProfile.Description)
		}
		if newProfile.MatchMessage != nil {
			query = query.Update("match_message", newProfile.MatchMessage)
		}
		if newProfile.LocationVisible != nil {
			query = query.Update("location_visible", newProfile.LocationVisible)
		}
		if newProfile.LocationTown != nil {
			query = query.Update("location_town", newProfile.LocationTown)
		}
		if newProfile.LocationState != nil {
			query = query.Update("location_state", newProfile.LocationState)
		}
		if newProfile.LocationCountry != nil {
			query = query.Update("location_country", newProfile.LocationCountry)
		}
		if newProfile.MessagingPlatform != nil {
			query = query.Update("messaging_platform", newProfile.MessagingPlatform)
		}
		if newProfile.MessagingUsername != nil {
			query = query.Update("messaging_username", newProfile.MessagingUsername)
		}
		err = query.UpdateColumn("deleted_at", nil).
			Preload("User").
			First(p).
			Error

		return err
	}

	// Otherwise, create an ephmatch profile

	// Get a user profile to load in the location data if we don't have it
	var user User
	err = NewUserModel(m.DB, m.log).GetUserByID(userID, &user)
	if err != nil {
		return err
	}

	locationTown := lib.StrPtrDefaults(newProfile.LocationTown, user.HomeTown)
	locationState := lib.StrPtrDefaults(newProfile.LocationState, user.HomeState)
	locationCountry := lib.StrPtrDefaults(newProfile.LocationCountry, user.HomeCountry)

	err = m.DB.
		Model(&EphmatchProfile{}).
		Where(EphmatchProfile{UserID: userID}).
		Assign(EphmatchProfile{
			Description:       newProfile.Description,
			MatchMessage:      newProfile.MatchMessage,
			LocationVisible:   newProfile.LocationVisible,
			LocationTown:      locationTown,
			LocationState:     locationState,
			LocationCountry:   locationCountry,
			MessagingPlatform: newProfile.MessagingPlatform,
			MessagingUsername: newProfile.MessagingUsername,
		}).
		Preload("User").
		FirstOrCreate(p).Error
	return err
}

func (m *EphmatchProfileModel) UpdateProfile(p *EphmatchProfile) (err error) {
	err = m.DB.Save(&p).Error
	if err != nil {
		return
	}

	err = m.DB.Preload("User").First(p, p.ID).Error
	return
}

func (m *EphmatchProfileModel) DeleteProfile(p *EphmatchProfile) (err error) {
	err = m.DB.Delete(p).Error
	return
}

func (m *EphmatchProfileModel) GetSelfProfileByID(userID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Unscoped().
		Preload("User").
		Preload("User.Tags").
		Where("ephmatch_profiles.user_id = ?", userID).
		First(p).Error
	return
}

func (m *EphmatchProfileModel) GetSelfProfileByIDScopedNoDefault(userID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Preload("User").
		Preload("User.Tags").
		Where("ephmatch_profiles.user_id = ?", userID).
		First(p).Error
	return
}

func (m *EphmatchProfileModel) GetProfileByID(profileUserID uint, p *EphmatchProfile) (err error) {
	err = m.DB.Model(&EphmatchProfile{}).
		Scopes(m.scopeDefault).
		Where("ephmatch_profiles.user_id = ?", profileUserID).
		Preload("User").
		Preload("User.Tags").
		First(p).Error
	if err != nil {
		return
	}

	return
}

func (m *EphmatchProfileModel) DoesProfileExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&EphmatchProfile{}).Scopes(m.scopeDefault).Where("ephmatch_profiles.user_id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// Default scope: is student and is visible. Make sure to enumerate users on join
func (m *EphmatchProfileModel) scopeDefault(db *gorm.DB) *gorm.DB {
	db = db.Joins("INNER JOIN users ON users.id = ephmatch_profiles.user_id").
		Where("users.type = ?", UserTypeStudent)
	return db
}

func (m *EphmatchProfileModel) SuggestUsers(userID uint) ([]uint, error) {
	/*
		Algorithm:
		  1. find similar users based on similar likes
		  2. find what else the similar users liked
	*/

	// Get who user liked
	rows, err := m.DB.Table("ephmatch_likes").
		Select("liked_id").
		Where("user_id = ?", userID).
		Rows()
	if err != nil {
		return nil, err
	}

	var likedUsers []uint
	var likedUsersStr []string
	likedUsersExists := make(map[uint]struct{}) // to check if a given userID was liked
	for rows.Next() {
		var likedUserID uint
		err = rows.Scan(&likedUserID)
		if err != nil {
			return nil, err
		}

		likedUsers = append(likedUsers, likedUserID)
		likedUsersStr = append(likedUsersStr, strconv.Itoa(int(likedUserID)))
		likedUsersExists[likedUserID] = struct{}{}
	}

	rows.Close()

	// Get users with similar like preferences (admirers of people user has liked)
	rows, err = m.DB.Table("ephmatch_likes").
		Select("user_id").
		Where("liked_id IN (" + strings.Join(likedUsersStr, ",") + ")").
		Rows()
	if err != nil {
		return nil, err
	}

	var admirers []uint // users similar to us
	var admirersStr []string
	for rows.Next() {
		var admirerID uint
		err = rows.Scan(&admirerID)
		if err != nil {
			return nil, err
		}

		admirers = append(admirers, admirerID)
		admirersStr = append(admirersStr, strconv.Itoa(int(admirerID)))
	}

	rows.Close()

	// Get other users that admirers liked

	rows, err = m.DB.Table("ephmatch_likes").
		Select("liked_id, count(*)").
		Where("user_id IN (" + strings.Join(admirersStr, ",") + ")").
		Group("liked_id").Rows()
	if err != nil {
		return nil, err
	}

	suggestedUsersToLikes := make(map[uint]int) // users suggested: mapped userID to number of times suggested
	var suggestedUsers []uint
	for rows.Next() {
		var likedID uint
		var count int
		err = rows.Scan(&likedID, &count)
		if err != nil {
			return nil, err
		}

		// Scale down the count if this user has already been liked. This way, liked users aren't all at the start
		if _, ok := likedUsersExists[likedID]; ok {
			// Simulate floats with 1 digit of accuracy as a faster way to do math
			count = (10 * count) / 15
		}

		suggestedUsersToLikes[likedID] = count
		suggestedUsers = append(suggestedUsers, likedID)
	}

	rows.Close()

	sort.Slice(suggestedUsers, func(i, j int) bool {
		return suggestedUsersToLikes[suggestedUsers[i]] < suggestedUsersToLikes[suggestedUsers[j]]
	})

	return suggestedUsers, nil
}
