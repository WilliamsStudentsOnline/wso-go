package main

import (
	"flag"
	"fmt"
	"math"
	"sort"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func main() {
	/* Flags */
	var configPath string
	var console bool

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in ")

	flag.Parse()

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "ephmatch-pagerank")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
		return
	}
	defer log.Sync()

	/* DATABASE */
	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

	/* Database Migrations */
	// NOTE: Job will not migrate anything; will fail if db is not updated on migrations
	/*dbUpToDate, err := migrate.MigrationUpToDate(migrate.MigrationGormOptions, db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}

	if !dbUpToDate {
		log.Fatal("Database migrations are not up to date")
	}*/

	// Do the actual stuff

	var likes []*models.EphmatchLike
	err = db.Model(models.EphmatchLike{}).
		Joins("INNER JOIN ephmatch_profiles AS ap ON ap.user_id = ephmatch_likes.user_id").
		Joins("INNER JOIN users AS au ON au.id = ephmatch_likes.user_id").
		Joins("INNER JOIN ephmatch_profiles AS bp ON bp.user_id = ephmatch_likes.liked_id").
		Joins("INNER JOIN users AS bu ON bu.id = ephmatch_likes.liked_id").
		Where("au.type = ? AND au.at_williams = ?", "student", true).
		Where("ap.deleted_at IS NULL").
		Where("bu.type = ? AND bu.at_williams = ?", "student", true).
		Where("bp.deleted_at IS NULL").
		Order("created_at ASC").
		Find(&likes).Error
	if err != nil {
		log.Fatal(err)
	}

	var profiles []*models.EphmatchProfile
	err = db.Model(models.EphmatchProfile{}).
		//Unscoped().
		Preload("User").
		Joins("INNER JOIN users AS u ON u.id = ephmatch_profiles.user_id").
		Where("u.type = ? AND u.at_williams = ?", "student", true).
		Find(&profiles).Error
	if err != nil {
		log.Fatal(err)
	}

	eloMap := make(map[uint]float64)
	for _, prof := range profiles {
		eloMap[prof.UserID] = 1000
	}

	for _, like := range likes {
		aScore := eloMap[like.UserID]
		bScore := eloMap[like.LikedID]

		// Because it is a like, we know that A likes B, so B>=A.
		aScore, bScore = elo(aScore, bScore, false)
		eloMap[like.UserID] = aScore
		eloMap[like.LikedID] = bScore
	}

	sort.Slice(profiles, func(i, j int) bool {
		return eloMap[profiles[i].UserID] < eloMap[profiles[j].UserID]
	})

	for _, prof := range profiles {
		fmt.Printf("%s - %f\n", prof.User.Name, eloMap[prof.UserID])
	}
}

func elo(aScore, bScore float64, aOverB bool) (aNewScore, bNewScore float64) {
	aExp := expVal(aScore, bScore)
	bExp := expVal(bScore, aScore)
	var k float64 = 20

	// A win means that A prefers themselves to B
	if aOverB {
		aNewScore = aScore + k*(1-aExp)
		bNewScore = bScore + k*(0-bExp)
	} else {
		aNewScore = aScore + k*(0-aExp)
		bNewScore = bScore + k*(1-bExp)
	}

	return
}

func expVal(x, y float64) (exp float64) {
	return 1 / (1 + math.Pow(10, (y-x)/100))
}
