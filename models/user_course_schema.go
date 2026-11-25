package models

// Hard-deleted so the unique index allows re-adding; it ignores NULL terms, so one someday row per course is enforced in code
type UserCourse struct {
	BaseSchema
	UserID uint  `gorm:"unique_index:idx_user_courses_term;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	CourseCanonicalID  uint              `gorm:"index:idx_user_courses_course;not null" json:"courseCanonicalID"`
	CourseCanonical    *CourseCanonical  `json:"courseCanonical,omitempty"`
	CourseInSemesterID *uint             `gorm:"unique_index:idx_user_courses_term" json:"courseInSemesterID"`
	CourseInSemester   *CourseInSemester `json:"courseInSemester,omitempty"`

	Hidden bool `gorm:"not null;default:false" json:"hidden"`

	Offerings []*UserCourseOffering `json:"offerings,omitempty"`
}

func (*UserCourse) TableName() string {
	return "user_courses"
}

type UserCourseOffering struct {
	BaseSchema
	UserCourseID     uint            `gorm:"unique_index:idx_user_course_offerings_pair;not null" json:"userCourseID"`
	CourseOfferingID uint            `gorm:"unique_index:idx_user_course_offerings_pair;index:idx_user_course_offerings_course_offering;not null" json:"offeringID"`
	CourseOffering   *CourseOffering `json:"offering,omitempty"`
}

func (*UserCourseOffering) TableName() string {
	return "user_course_offerings"
}
