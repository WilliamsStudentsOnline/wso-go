package models

// EphmatchLike Schema
type EphmatchLike struct {
	BaseSchema

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `json:"userID"`

	// Belongs to the user that the User liked
	Liked   *User `json:"liked"`
	LikedID uint  `json:"likedID"`
}

func (*EphmatchLike) TableName() string {
	return "ephmatch_likes"
}
