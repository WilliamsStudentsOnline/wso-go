package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dorms_update"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

//go:generate go run gen.go

func main() {
	/* Flags */
	var configPath string
	var useLocal bool
	var dormPath string
	var roomsPath string
	var disableMigrationCheck bool

	// Command-line flags
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&useLocal, "local", true, "use local embedded data")
	flag.StringVar(&dormPath, "dorm", "jobs/dorms_update/data/dorms.csv", "path to dorm info csv file")
	flag.StringVar(&roomsPath, "rooms", "jobs/dorms_update/data/rooms", "path to room info directory of csv files")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")

	flag.Parse()

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	/* LOGGING */
	log, err := config.NewDefaultProductionLog(cfg)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

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
	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

	/* Database Migrations */
	// NOTE: Job will not migrate anything; will fail if db is not updated on migrations
	dbUpToDate, err := migrate.MigrationUpToDate(migrate.MigrationGormOptions, db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}

	if !dbUpToDate {
		if disableMigrationCheck {
			log.Warn("Database migrations are not up to date")
		} else {
			log.Fatal("Database migrations are not up to date")
		}
	}

	/* Actual logic of code */
	log.Info("Updating Dorms")

	/* Parse Dorm CSV */
	dorms, err := ReadDorms(dormPath, useLocal)
	if err != nil {
		log.Fatal("Dorm Parse Error: " + err.Error())
		return
	}

	/* Run update dorms */
	err = dorms_update.UpdateDorms(db, log, dorms)
	if err != nil {
		log.Fatal("Update Dorms Error: " + err.Error())
		return
	}

	/* Parse Room CSV */
	log.Info("Updating Dorm Rooms")

	// Get trakked dorms in database
	var dbDorms []*models.Dorm
	err = db.Scopes(models.NewDormModel(nil, log).ScopeTrakked).Find(&dbDorms).Error
	if err != nil {
		log.Fatal("Get Trakked Dorms Error: " + err.Error())
		return
	}

	for _, dbDorm := range dbDorms {
		dormRoomPath := filepath.Join(roomsPath, dbDorm.Name+".csv")
		rooms, err := ReadRooms(dormRoomPath, useLocal)
		if err != nil {
			log.Fatal("Read Rooms Error: " + err.Error())
			return
		}

		err = dorms_update.UpdateRooms(db, log, rooms, dbDorm)
		if err != nil {
			log.Fatal("Update Rooms Error: " + err.Error())
			return
		}
	}

	/* Update Dorm Statistics */
	// Dorm facts are automatically updated with the dorm room updates.
	log.Info("Updating Dorm Statistics")

	err = dorms_update.UpdateDormsStatistics(db, log)
	if err != nil {
		log.Fatal("Update Dorm Statistics Error: " + err.Error())
		return
	}

	log.Info("Finished")
}

func ReadDorms(file string, useLocal bool) ([]*dorms_update.Dorm, error) {
	var r io.Reader

	if useLocal {
		r = bytes.NewBufferString(dormsCSV)
	} else {
		// Open file
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer r.Close()
	}

	// Read file into matrix of strings
	lines, err := csv.NewReader(r).ReadAll()
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

func ReadRooms(file string, useLocal bool) ([]*dorms_update.Room, error) {
	var r io.Reader

	if useLocal {
		baseFile := filepath.Base(file)
		r = bytes.NewBufferString(roomCSVs[baseFile])
	} else {
		// Open file
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer r.Close()
	}

	// Read file into matrix of strings
	lines, err := csv.NewReader(r).ReadAll()
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
