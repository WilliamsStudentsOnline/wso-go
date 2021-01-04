package models

import (
	"sort"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Ephmatch Matches Model
type EphmatchMatchesModel struct {
	*BaseModel
}

func NewEphmatchMatchesModel(db *gorm.DB, log *zap.SugaredLogger) *EphmatchMatchesModel {
	return &EphmatchMatchesModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func orderUIntPair(a uint, b uint) (smallest uint, largest uint) {
	smallest = a
	largest = b
	if smallest > largest {
		smallest = b
		largest = a
	}
	return
}

// Check if two users are matching by seeing if there is already an ephmatch from that user and with the other user.
func (m *EphmatchMatchesModel) IsMatching(userAID uint, userBID uint) (matching bool, err error) {
	smallest, largest := orderUIntPair(userAID, userBID)

	var count int
	// Search for both orders
	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserAID: smallest,
		UserBID: largest,
	}).Count(&count).Error
	matching = count > 0
	return
}

// Check if ephmatch already exists by seeing if there is already an ephmatch from that user and with the other user.
func (m *EphmatchMatchesModel) CheckDuplicateMatch(userAID uint, userBID uint) (duplicate bool, err error) {
	return m.IsMatching(userAID, userBID)
}

type GetMatchesOptions struct {
	// You can preload: tags
	Preload []string `json:"preload" form:"preload[]"`
}

// Preload specifically allowed parts if requested
func (o *GetMatchesOptions) Preloader(userB bool, db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "tags") {
		if userB {
			db = db.Preload("UserB.Tags")
		} else {
			db = db.Preload("UserA.Tags")
		}
	}

	return db
}

// Matched based on when the match was last made
// We use updated_at b/c it represents the latest match (eg a match may be deleted and recreated, thus it is new)
// If we use created_at, we would represent matches in order and ignore random noise that may push it to the front (many deletes and recreations)
func (o *GetMatchesOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("ephmatch_matches.updated_at desc", true)
}

func (o *GetMatchesOptions) Run(userB bool, db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	return o.Preloader(userB, db)
}

// Get ephmatch matches of user.
func (m *EphmatchMatchesModel) GetMatches(userID uint, opts *GetMatchesOptions, p *[]*EphmatchMatch) (err error) {

	// Get matches
	var matchesA []*EphmatchMatch
	var matchesB []*EphmatchMatch

	// Get all matches with userA being us, userB being other
	dbA := m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserAID: userID,
	}).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatch_matches.user_b_id").
		Where("u.type = ? AND u.at_williams = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = ephmatch_matches.user_b_id").
		Where("p.deleted_at IS NULL").
		// Preload other column and other's ephmatch profile
		Preload("UserB").
		Preload("UserB.EphmatchProfile")
	dbA = opts.Run(true, dbA)
	err = dbA.Find(&matchesA).Error
	if err != nil {
		return
	}

	// Get all matches with userB being us, userA being other
	dbB := m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserBID: userID,
	}).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatch_matches.user_a_id").
		Where("u.type = ? AND u.at_williams = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = ephmatch_matches.user_a_id").
		Where("p.deleted_at IS NULL").
		// Preload other column and other's ephmatch profile
		Preload("UserA").
		Preload("UserA.EphmatchProfile")
	dbB = opts.Run(false, dbB)
	err = dbB.Find(&matchesB).Error
	if err != nil {
		return
	}

	// Combine matches and fill in the other user as the matched user
	allMatches := make([]*EphmatchMatch, len(matchesA)+len(matchesB))
	for i, match := range matchesA {
		allMatches[i] = &EphmatchMatch{
			BaseSchema: BaseSchema{
				ID:        match.ID,
				CreatedAt: match.CreatedAt,
				UpdatedAt: match.UpdatedAt,
			},
			MatchedUser:       match.UserB,
			MatchedUserID:     match.UserBID,
			SeenByMatchedUser: match.UserBSeen,
			SeenBySelf:        match.UserASeen,
		}
	}
	for i, match := range matchesB {
		allMatches[i+len(matchesA)] = &EphmatchMatch{
			BaseSchema: BaseSchema{
				ID:        match.ID,
				CreatedAt: match.CreatedAt,
				UpdatedAt: match.UpdatedAt,
			},
			MatchedUser:       match.UserA,
			MatchedUserID:     match.UserAID,
			SeenByMatchedUser: match.UserASeen,
			SeenBySelf:        match.UserBSeen,
		}
	}
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].CreatedAt.After(allMatches[j].CreatedAt)
	})
	*p = allMatches

	return
}

func (m *EphmatchMatchesModel) SetMatchesAsSeen(userID uint) (err error) {
	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserAID:   userID,
		UserASeen: lib.FalsePtr(),
	}).Update("user_a_seen", lib.TruePtr()).Error
	if err != nil {
		return
	}

	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserBID:   userID,
		UserBSeen: lib.FalsePtr(),
	}).Update("user_b_seen", lib.TruePtr()).Error
	if err != nil {
		return
	}

	return
}

func (m *EphmatchMatchesModel) CountMatchesAndUnseen(userID uint) (unseen int, total int, err error) {
	var totalA, totalB, unseenA, unseenB int

	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserAID: userID,
	}).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatch_matches.user_b_id").
		Where("u.type = ? AND u.at_williams = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = ephmatch_matches.user_b_id").
		Where("p.deleted_at IS NULL").
		Count(&totalA).Error
	if err != nil {
		return
	}

	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserBID: userID,
	}).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatch_matches.user_a_id").
		Where("u.type = ? AND u.at_williams = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = ephmatch_matches.user_a_id").
		Where("p.deleted_at IS NULL").
		Count(&totalB).Error
	if err != nil {
		return
	}

	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserAID:   userID,
		UserASeen: lib.FalsePtr(),
	}).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatch_matches.user_b_id").
		Where("u.type = ? AND u.at_williams = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = ephmatch_matches.user_b_id").
		Where("p.deleted_at IS NULL").
		Count(&unseenA).Error
	if err != nil {
		return
	}

	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserBID:   userID,
		UserBSeen: lib.FalsePtr(),
	}).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatch_matches.user_a_id").
		Where("u.type = ? AND u.at_williams = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = ephmatch_matches.user_a_id").
		Where("p.deleted_at IS NULL").
		Count(&unseenB).Error
	if err != nil {
		return
	}

	unseen = unseenA + unseenB
	total = totalA + totalB

	return
}

// Get raw ephmatch matches of user. This will just have IDs, no users or profiles.
func (m *EphmatchMatchesModel) GetRawUserMatches(userID uint, p *[]*EphmatchMatch) (err error) {
	// Get matches

	var matchesA []*EphmatchMatch
	var matchesB []*EphmatchMatch

	// Get all matches with userA being us, userB being other
	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserAID: userID,
	}).Find(&matchesA).Error
	if err != nil {
		return
	}

	// Get all matches with userB being us, userA being other
	err = m.DB.Model(&EphmatchMatch{}).Where(&EphmatchMatch{
		UserBID: userID,
	}).Find(&matchesB).Error
	if err != nil {
		return
	}

	// Combine matches and fill in the other user as the matched user
	allMatches := make([]*EphmatchMatch, len(matchesA)+len(matchesB))
	for i := range matchesA {
		allMatches[i] = &EphmatchMatch{
			MatchedUserID: matchesA[i].UserBID,
		}
	}
	for i := range matchesB {
		allMatches[i+len(matchesA)] = &EphmatchMatch{
			MatchedUserID: matchesB[i].UserAID,
		}
	}
	*p = allMatches

	return
}
