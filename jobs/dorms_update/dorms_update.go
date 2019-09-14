package dorms_update

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// name,neighborhood_name,number_singles,number_doubles,number_flex,description,number_bathrooms,number_washers
type Dorm struct {
	Name             string
	NeighborhoodName string
	NumberSingles    int
	NumberDoubles    int
	NumberFlex       int
	Description      string
	NumberBathrooms  int
	NumberWashers    int
}

// number,floor,type,faces,area,common_room_access
type Room struct {
	Number           string
	Floor            int
	Type             string
	Faces            string
	Area             int
	CommonRoomAccess bool
}

func UpdateDorms(db *gorm.DB, log *logrus.Logger, dorms []*Dorm) (err error) {
	dormNames := make(map[string]bool)

	// Add/update dorms from data to db
	for _, dorm := range dorms {
		// Must get neighborhood
		var hood models.Neighborhood
		err = db.Where("neighborhoods.name = ?", dorm.NeighborhoodName).First(&hood).Error
		if err != nil {
			return
		}

		// Get dorm in DB
		var dbDorm models.Dorm
		dbErr := db.Where("dorms.neighborhood_id = ? AND dorms.name = ?", hood.ID, dorm.Name).First(&dbDorm).Error
		// Either we create a new dorm or update one
		if dbErr == nil {
			// If we must update the dorm, we will have no error
			log.Infof("Updating dorm %s/%s", hood.Name, dbDorm.Name)

			// Do the update
			dbDorm.NumberSingles = &dorm.NumberSingles
			dbDorm.NumberDoubles = &dorm.NumberDoubles
			dbDorm.NumberFlex = &dorm.NumberFlex
			dbDorm.Description = &dorm.Description
			dbDorm.NumberBathrooms = &dorm.NumberBathrooms
			dbDorm.NumberWashers = &dorm.NumberWashers

			// Save the updated dorm
			err = db.Save(&dbDorm).Error
			if err != nil {
				return
			}
		} else if gorm.IsRecordNotFoundError(dbErr) {
			// If this is a new dorm, we will get a record not found error
			log.Infof("Creating dorm %s/%s", hood.Name, dorm.Name)

			// Fill in create data
			dbDorm.Name = dorm.Name
			dbDorm.NeighborhoodID = hood.ID
			dbDorm.NumberSingles = &dorm.NumberSingles
			dbDorm.NumberDoubles = &dorm.NumberDoubles
			dbDorm.NumberFlex = &dorm.NumberFlex
			dbDorm.Description = &dorm.Description
			dbDorm.NumberBathrooms = &dorm.NumberBathrooms
			dbDorm.NumberWashers = &dorm.NumberWashers

			// Create the new dorm
			err = db.Create(&dbDorm).Error
			if err != nil {
				return
			}
		} else {
			// This is if we have a real db error
			return dbErr
		}

		// Add dorm to the name list
		dormNames[fmt.Sprintf("%s/%s", dorm.NeighborhoodName, dorm.Name)] = true
	}

	// Remove dorms from db that don't appear in data
	var dbDorms []*models.Dorm
	err = db.Preload("Neighborhood").Find(&dbDorms).Error
	if err != nil {
		return
	}

	for _, dbDorm := range dbDorms {
		// If it is in the data dorm list, ignore it
		if _, ok := dormNames[fmt.Sprintf("%s/%s", dbDorm.Neighborhood.Name, dbDorm.Name)]; ok {
			continue
		}

		log.Warnf("Deleting dorm %s/%s", dbDorm.Neighborhood.Name, dbDorm.Name)

		// Get the associated dorm rooms
		var dbDormRooms []*models.DormRoom
		err = db.Where("dorm_rooms.dorm_id = ?", dbDorm.ID).Find(&dbDormRooms).Error
		if err != nil {
			return
		}

		// Go through each dorm room and delete the associated dormtrak reviews
		for _, dbDormRoom := range dbDormRooms {
			log.Debugf("Deleting dormtrak reviews of dorm room %d", dbDormRoom.ID)
			err = db.Unscoped().Delete(&models.DormtrakReview{DormRoomID: dbDormRoom.ID}).Error
			if err != nil {
				return
			}
		}

		// Delete dorm rooms
		log.Debugf("Deleting dorm rooms of dorm %d", dbDorm.ID)
		err = db.Unscoped().Delete(&models.DormRoom{DormID: dbDorm.ID}).Error
		if err != nil {
			return
		}

		// Delete the dorm
		log.Debugf("Deleting dorm %d", dbDorm.ID)
		err = db.Unscoped().Delete(dbDorm).Error
		if err != nil {
			return
		}
	}

	return
}

func UpdateRooms(db *gorm.DB, log *logrus.Logger, rooms []*Room, dorm *models.Dorm) (err error) {
	roomNumbers := make(map[string]bool)

	// Add/update dorm rooms from data to db
	for _, room := range rooms {
		// Get dorm in DB
		var dbRoom models.DormRoom
		dbErr := db.Where(
			"dorm_rooms.dorm_id = ? AND dorm_rooms.number = ?",
			dorm.ID, room.Number,
		).First(&dbRoom).Error
		// Either we create a new room or update one
		if dbErr == nil {
			// If we must update the room, we will have no error
			log.Infof("Updating room %s/%s", dorm.Name, room.Number)

			// Do the update
			dbRoom.FloorNumber = &room.Floor
			dbRoom.Faces = &room.Faces
			dbRoom.Area = &room.Area
			dbRoom.CommonRoomAccess = &room.CommonRoomAccess
			dbRoom.RoomType = room.Type

			// Save the updated room
			err = db.Save(&dbRoom).Error
			if err != nil {
				return
			}
		} else if gorm.IsRecordNotFoundError(dbErr) {
			// If this is a new room, we will get a record not found error
			log.Infof("Creating room %s/%s", dorm.Name, room.Number)

			// Fill in create data
			dbRoom.Number = room.Number
			dbRoom.DormID = dorm.ID
			dbRoom.FloorNumber = &room.Floor
			dbRoom.Faces = &room.Faces
			dbRoom.Area = &room.Area
			dbRoom.CommonRoomAccess = &room.CommonRoomAccess
			dbRoom.RoomType = room.Type

			// Create the new room
			err = db.Create(&dbRoom).Error
			if err != nil {
				return
			}
		} else {
			// This is if we have a real db error
			return dbErr
		}

		// Add room to the room numbers list
		roomNumbers[room.Number] = true
	}

	// Remove dorms from db that don't appear in data
	var dbRooms []*models.DormRoom
	err = db.Where("dorm_rooms.dorm_id = ?", dorm.ID).Find(&dbRooms).Error
	if err != nil {
		return
	}

	for _, dbRoom := range dbRooms {
		// If it is in the data room list, ignore it
		if _, ok := roomNumbers[dbRoom.Number]; ok {
			continue
		}

		log.Warnf("Deleting room %s/%s", dorm.Name, dbRoom.Number)

		// Delete dorm room reviews
		log.Debugf("Deleting reviews of dorm room %s", dbRoom.Number)
		err = db.Unscoped().Delete(&models.DormtrakReview{DormRoomID: dbRoom.ID}).Error
		if err != nil {
			return
		}

		// Delete the dorm
		log.Debugf("Deleting room %s", dbRoom.Number)
		err = db.Unscoped().Delete(dbRoom).Error
		if err != nil {
			return
		}
	}

	return nil
}

func UpdateDormsStatistics(db *gorm.DB, log *logrus.Logger) (err error) {
	dormModel := models.NewDormModel(db)

	// Remove dorms from db that don't appear in data
	var dorms []*models.Dorm
	err = db.Scopes(dormModel.ScopeTrakked).Find(&dorms).Error
	if err != nil {
		return
	}

	for _, dorm := range dorms {
		log.Infof("Updating dorm statistics %s", dorm.Name)
		err = dormModel.ReloadStatistics(dorm.ID)
		if err != nil {
			return
		}
	}

	return
}
