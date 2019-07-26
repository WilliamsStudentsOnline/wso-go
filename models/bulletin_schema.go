package models

import "time"

const (
	BulletinTypeLostAndFound = "lostAndFound"
	BulletinTypeJob          = "job"
	BulletinTypeRide         = "ride"
	BulletinTypeExchange     = "exchange"
	BulletinTypeAnnouncement = "announcement"
)

// Bulletin Model Schema
type Bulletin struct {
	BaseSchema
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `gorm:"size:65535" json:"body"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`

	// Author information
	UserID uint  `json:"userID"`
	User   *User `json:"user,omitempty"`
}

func (*Bulletin) TableName() string {
	return "bulletin"
}

func NewBulletinWithID(bulletinID uint) Bulletin {
	return Bulletin{
		BaseSchema: BaseSchema{
			ID: bulletinID,
		},
	}
}

func (b *Bulletin) IsLostAndFound() bool {
	return b.Type == BulletinTypeLostAndFound
}

func (b *Bulletin) IsJob() bool {
	return b.Type == BulletinTypeJob
}

func (b *Bulletin) IsRide() bool {
	return b.Type == BulletinTypeRide
}

func (b *Bulletin) IsExchange() bool {
	return b.Type == BulletinTypeExchange
}

func (b *Bulletin) IsAnnouncement() bool {
	return b.Type == BulletinTypeAnnouncement
}
