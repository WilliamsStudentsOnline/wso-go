package models

const (
	EphmatchProfileGenderMale   = "he/him/his"
	EphmatchProfileGenderFemale = "she/her/hers"
	EphmatchProfileGenderNB     = "they/them/theirs"
	EphmatchProfileGenderOther  = ""
)

func EphmatchProfileGenderType(gender string) string {
	switch gender {
	case EphmatchProfileGenderMale:
		return EphmatchProfileGenderMale
	case EphmatchProfileGenderFemale:
		return EphmatchProfileGenderFemale
	case EphmatchProfileGenderNB:
		return EphmatchProfileGenderNB
	default:
		return EphmatchProfileGenderOther
	}
}

type EphmatchProfile struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_ephmatch_profiles_on_user_id;not null;" json:"userID"`
	User   *User `json:"user,omitempty"`

	Gender      string `json:"gender"`
	Description string `json:"description"`

	Liked bool `gorm:"-" json:"liked"` // If me (user) has an ephmatch entry where ephmatch.other_id=users.id and ephmatch.user_id=myID
}

func (*EphmatchProfile) TableName() string {
	return "ephmatch_profiles"
}
