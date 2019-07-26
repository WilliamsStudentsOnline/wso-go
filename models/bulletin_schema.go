package models

import "time"

// Bulletin Types
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

// TableName returns the name of the bulletins table
func (*Bulletin) TableName() string {
	return "bulletin"
}

// NewBulletinWithID creates a new bulletin with ID
func NewBulletinWithID(bulletinID uint) Bulletin {
	return Bulletin{
		BaseSchema: BaseSchema{
			ID: bulletinID,
		},
	}
}

// IsLostAndFound checks if the bulletin is a lost and found posting
func (b *Bulletin) IsLostAndFound() bool {
	return b.Type == BulletinTypeLostAndFound
}

// IsJob checks if the bulletin is a job posting
func (b *Bulletin) IsJob() bool {
	return b.Type == BulletinTypeJob
}

// IsRide checks if the bulletin is a ride offer/request
func (b *Bulletin) IsRide() bool {
	return b.Type == BulletinTypeRide
}

// IsExchange checks if the the bulletin is an exchange
func (b *Bulletin) IsExchange() bool {
	return b.Type == BulletinTypeExchange
}

// IsAnnouncement checks if the bulletin is an announcement
func (b *Bulletin) IsAnnouncement() bool {
	return b.Type == BulletinTypeAnnouncement
}
