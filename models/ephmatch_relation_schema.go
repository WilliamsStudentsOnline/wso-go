package models

const (
	EphmatchRelationLike    = "like"
	EphmatchRelationDislike = "dislike"
)

func ValidateEphmatchRelation(str string) bool {
	switch str {
	case EphmatchRelationLike, EphmatchRelationDislike:
		return true
	default:
		return false
	}
}

// EphmatchRelation Schema
type EphmatchRelation struct {
	BaseSchema

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `gorm:"index:index_ephmatch_relations_on_user_id;not null;" json:"userID"`

	// Belongs to the user that the User has a (asymmetrical) relation with
	Other   *User `json:"other"`
	OtherID uint  `gorm:"index:index_ephmatch_relations_on_other_id;not null;" json:"otherID"`

	Relation string `json:"relation"`
}

func (*EphmatchRelation) TableName() string {
	return "ephmatch_relations"
}
