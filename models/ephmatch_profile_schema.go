package models

const (
	EphmatchProfileGenderMale   = "he/him/his"
	EphmatchProfileGenderFemale = "she/her/hers"
	EphmatchProfileGenderNB     = "they/them/theirs"
	EphmatchProfileGenderOther  = "<OTHER>"
	EphmatchProfileGenderNone   = ""
)

func EphmatchProfileGenderType(gender string) string {
	switch gender {
	case EphmatchProfileGenderMale:
		return EphmatchProfileGenderMale
	case EphmatchProfileGenderFemale:
		return EphmatchProfileGenderFemale
	case EphmatchProfileGenderNB:
		return EphmatchProfileGenderNB
	case "":
		return EphmatchProfileGenderNone
	default:
		return EphmatchProfileGenderOther
	}
}

type EphmatchProfile struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_ephmatch_profiles_on_user_id;not null;" json:"userID"`
	User   *User `json:"user,omitempty"`

	// These are optional
	Gender       *string `json:"gender"`
	Description  *string `json:"description"`
	MatchMessage *string `json:"matchMessage"`

	Liked bool `gorm:"-" json:"liked"` // If me (user) has an ephmatch entry where ephmatch.other_id=users.id and ephmatch.user_id=myID

	// Non db entry that acts as a flag for deleted_at column
	Deleted bool `gorm:"-" json:"deleted"`
}

func (*EphmatchProfile) TableName() string {
	return "ephmatch_profiles"
}

// Again, I hate hooks but this is the best way.
// This populates the deleted flag on the profile
func (p *EphmatchProfile) AfterFind() (err error) {
	p.Deleted = p.DeletedAt != nil
	return
}
