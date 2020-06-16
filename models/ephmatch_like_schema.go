package models

// EphmatchLike Schema
type EphmatchLike struct {
	BaseSchema

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `gorm:"index:index_ephmatch_likes_on_user_id;not null;" json:"userID"`

	// Belongs to the user that the User liked
	Liked   *User `json:"liked"`
	LikedID uint  `gorm:"index:index_ephmatch_likes_on_liked_id;not null;" json:"likedID"`
}

func (*EphmatchLike) TableName() string {
	return "ephmatch_likes"
}
