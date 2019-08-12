package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dorms_update"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func main() {
	/* Flags */
	var configPath string
	var dormPath string
	var roomsPath string

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&dormPath, "dorm", "data/dormtrak/dorms.csv", "path to dorm info csv file")
	flag.StringVar(&roomsPath, "rooms", "data/dormtrak/rooms", "path to room info directory of csv files")

	flag.Parse()

	/* Logging */
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Config Error: " + err.Error())
		return
	}

	/* Set Logging Level */
	log.SetLevel(cfg.LogLevelParsed)

	/* Parse params */
	dormPath, err = filepath.Abs(dormPath)
	if err != nil {
		log.Fatal("Dorm Path Error: " + err.Error())
		return
	}

	roomsPath, err = filepath.Abs(roomsPath)
	if err != nil {
		log.Fatal("Rooms Path Error: " + err.Error())
		return
	}

	/* DATABASE */
	db := config.LoadDatabase(cfg)
	defer config.CloseDatabase(db)

	/* Database Migrations */
	// NOTE: Job will not migrate anything; will fail if DB is not updated on migrations
	lastMigrationID, err := migrate.LastMigration(migrate.MigrationGormOptions, db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}

	if lastMigrationID != migrate.Migrations[len(migrate.Migrations)-1].ID {
		log.Fatal("Database migrations are not up to date")
	}

	/* Actual logic of code */

	/* Parse Dorm CSV */
	dorms, err := ReadDorms(dormPath)
	if err != nil {
		log.Fatal("Dorm Parse Error: " + err.Error())
		return
	}

	/* Run update dorms */
	logger := log.New()
	err = dorms_update.UpdateDorms(db, logger, dorms)
	if err != nil {
		log.Fatal("Update Dorms Error: " + err.Error())
		return
	}

	/* Parse Room CSV */
	// Get trakked dorms in database
	var dbDorms []*models.Dorm
	err = db.Scopes(dormScopeTrakked).Find(&dbDorms).Error
	if err != nil {
		log.Fatal("Get Trakked Dorms Error: " + err.Error())
		return
	}

	for _, dbDorm := range dbDorms {
		dormRoomPath := filepath.Join(roomsPath, dbDorm.Name+".csv")
		rooms, err := ReadRooms(dormRoomPath)
		if err != nil {
			log.Fatal("Read Rooms Error: " + err.Error())
			return
		}

		err = dorms_update.UpdateRooms(db, logger, rooms, dbDorm)
		if err != nil {
			log.Fatal("Update Rooms Error: " + err.Error())
			return
		}
	}

	log.Info("Finished")
}

func ReadDorms(file string) ([]*dorms_update.Dorm, error) {
	// Open file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read file into matrix of strings
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var dorms []*dorms_update.Dorm

	// Go through lines and turn them into structs. Ignore first line as it is for header
	for _, line := range lines[1:] {
		numSing := 0
		if line[2] != "" {
			numSing, err = strconv.Atoi(line[2])
			if err != nil {
				return nil, err
			}
		}

		numDoub := 0
		if line[3] != "" {
			numDoub, err = strconv.Atoi(line[3])
			if err != nil {
				return nil, err
			}
		}

		numFlex := 0
		if line[4] != "" {
			numFlex, err = strconv.Atoi(line[4])
			if err != nil {
				return nil, err
			}
		}

		numBath := 0
		if line[6] != "" {
			numBath, err = strconv.Atoi(line[6])
			if err != nil {
				return nil, err
			}
		}

		numWash := 0
		if line[7] != "" {
			numWash, err = strconv.Atoi(line[7])
			if err != nil {
				return nil, err
			}
		}

		dorms = append(dorms, &dorms_update.Dorm{
			Name:             line[0],
			NeighborhoodName: line[1],
			NumberSingles:    numSing,
			NumberDoubles:    numDoub,
			NumberFlex:       numFlex,
			Description:      line[5],
			NumberBathrooms:  numBath,
			NumberWashers:    numWash,
		})
	}

	return dorms, nil
}

func ReadRooms(file string) ([]*dorms_update.Room, error) {
	// Open file
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read file into matrix of strings
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var rooms []*dorms_update.Room

	// Go through lines and turn them into structs. Ignore first line as it is for header
	for _, line := range lines[1:] {
		floor, err := strconv.Atoi(line[1])
		if err != nil {
			return nil, err
		}

		area, err := strconv.Atoi(line[4])
		if err != nil {
			return nil, err
		}

		roomType := line[2]
		if roomType != models.DormRoomTypeSingle &&
			roomType != models.DormRoomTypeDouble &&
			roomType != models.DormRoomTypeFlex {
			return nil, fmt.Errorf("unknown room type \"%s\"", roomType)
		}

		croomAccess, err := strconv.ParseBool(line[5])
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, &dorms_update.Room{
			Number:           line[0],
			Floor:            floor,
			Type:             roomType,
			Faces:            line[3],
			Area:             area,
			CommonRoomAccess: croomAccess,
		})
	}

	return rooms, nil
}

func dormScopeTrakked(db *gorm.DB) *gorm.DB {
	return db.
		Joins("JOIN neighborhoods ON neighborhoods.id = dorms.neighborhood_id").
		Where("neighborhoods.trakked = ?", true)
}
