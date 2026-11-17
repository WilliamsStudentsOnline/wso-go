package models

type FactrakDeficitExemptionStatus string

const (
	FactrakDeficitExemptionStatusPending  FactrakDeficitExemptionStatus = "pending"
	FactrakDeficitExemptionStatusApproved FactrakDeficitExemptionStatus = "approved"
	FactrakDeficitExemptionStatusRejected FactrakDeficitExemptionStatus = "rejected"
)

const FactrakDeficitExemptionReasonMaxLength = 1000

type FactrakDeficitExemption struct {
	BaseSchema

	UserID uint  `gorm:"index:index_factrak_deficit_exemptions_on_user_id_semester;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// The semester it's for, stored like FactrakSurvey's semester fields
	SemesterSeason string `gorm:"size:32;index:index_factrak_deficit_exemptions_on_user_id_semester;not null" json:"semesterSeason" enums:"fall,spring"`
	SemesterYear   int    `gorm:"index:index_factrak_deficit_exemptions_on_user_id_semester;not null" json:"semesterYear"`

	Reason string                        `gorm:"size:65535;not null" json:"reason"`
	Status FactrakDeficitExemptionStatus `gorm:"size:16;index:index_factrak_deficit_exemptions_on_status;not null;default:'pending'" json:"status" enums:"pending,approved,rejected"`
}

func (*FactrakDeficitExemption) TableName() string {
	return "factrak_deficit_exemptions"
}
