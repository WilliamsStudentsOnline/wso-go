package db

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

// SQLite tests + `make atlas-schema` desired-state dump
func AutoMigrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.Neighborhood{},
		&models.Dorm{},
		&models.DormRoom{},
		&models.Office{},
		&models.Bulletin{},
		&models.Tag{},
		&models.AreaOfStudy{},
		&models.Course{},
		&models.FactrakAgreement{},
		&models.FactrakSurvey{},
		&models.DormtrakReview{},
		&models.Ephcatch{},
		&models.BulletinRide{},
		&models.Discussion{},
		&models.Post{},
		&models.EphmatchProfile{},
		&models.EphmatchMatch{},
		&models.EphmatchLike{},
		&models.NotificationSettings{},
		&models.NotificationToken{},
		&models.GoodrichMenuItem{},
		&models.GoodrichOrder{},
		&models.EphmatchRelation{},
		&models.BannedUser{},
		&models.BookListing{},
		&models.Book{},
	).Error
}
