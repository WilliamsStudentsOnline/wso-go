package models

const (
	EphmatchMessagingPlatformPhone     = "Phone"
	EphmatchMessagingPlatformSnapchat  = "Snapchat"
	EphmatchMessagingPlatformInstagram = "Instagram"
)

const (
	EphmatchLookingForFriends = "friends"
	EphmatchLookingForFun     = "fun"
	EphmatchLookingForCasual  = "casual" // something casual
	EphmatchLookingForLove    = "love"
	EphmatchLookingForOpen    = "open" // open for whatever
)

func ValidateEphmatchMessagingPlatform(str string) bool {
	switch str {
	case EphmatchMessagingPlatformPhone, EphmatchMessagingPlatformSnapchat, EphmatchMessagingPlatformInstagram:
		return true
	default:
		return false
	}
}

func ValidateEphmatchLookingFor(str string) bool {
	switch str {
	case EphmatchLookingForFriends, EphmatchLookingForFun, EphmatchLookingForCasual, EphmatchLookingForLove, EphmatchLookingForOpen:
		return true
	default:
		return false
	}
}

type EphmatchProfile struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_ephmatch_profiles_on_user_id;not null;" json:"userID"`
	User   *User `json:"user,omitempty"`

	// These are optional
	Description  *string `json:"description"`
	MatchMessage *string `json:"matchMessage"`

	Relation *string `gorm:"-" json:"relation,omitempty"` // If self has an out-relation with this profile (user)
	Matched  *bool   `gorm:"-" json:"matched,omitempty"`  // If user and self are matched

	// Current location columns
	LocationVisible *bool   `gorm:"DEFAULT:true;not null" json:"locationVisible"`
	LocationTown    *string `json:"locationTown"`
	LocationState   *string `json:"locationState"`
	LocationCountry *string `json:"locationCountry"`

	// Messaging platform columns
	MessagingPlatform *string `json:"messagingPlatform"`
	MessagingUsername *string `json:"messagingUsername"`

	// Looking for:
	LookingFor *string `json:"lookingFor"`

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
