package main

import (
	"flag"
	"fmt"
	"sort"
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/dcadenas/pagerank"
	"github.com/emicklei/dot"
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
	uModel := models.NewUserModel(db, log)

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
		Find(&likes).Error
	if err != nil {
		log.Fatal(err)
	}

	graph := pagerank.New()

	for _, like := range likes {
		graph.Link(int(like.UserID), int(like.LikedID))
	}

	rankMap := make(map[uint]float64)
	ids := []uint{}

	graph.Rank(0.85, 0.00001, func(id int, rank float64) {
		rankMap[uint(id)] = rank
		ids = append(ids, uint(id))
	})

	sort.Slice(ids, func(i, j int) bool {
		return rankMap[ids[i]] < rankMap[ids[j]]
	})

	g := dot.NewGraph(dot.Directed)

	for _, id := range ids {
		user := models.User{}
		err = uModel.GetUserByID(id, &user)
		if err != nil {
			log.Fatal(err)
		}

		g.Node(strconv.Itoa(int(id))).Label(user.Name)

		var usrLikes []*models.EphmatchLike
		err = db.Model(&models.EphmatchLike{}).
			Joins("INNER JOIN ephmatch_profiles AS bp ON bp.user_id = ephmatch_likes.liked_id").
			Joins("INNER JOIN users AS bu ON bu.id = ephmatch_likes.liked_id").
			Where("bu.type = ? AND bu.at_williams = ?", "student", true).
			Where("bp.deleted_at IS NULL").
			Where(&models.EphmatchLike{
				UserID: user.ID,
			}).
			Find(&usrLikes).Error
		if err != nil {
			log.Fatal(err)
		}

		var usrAdmi []*models.EphmatchLike
		err = db.Model(&models.EphmatchLike{}).
			Joins("INNER JOIN ephmatch_profiles AS ap ON ap.user_id = ephmatch_likes.user_id").
			Joins("INNER JOIN users AS au ON au.id = ephmatch_likes.user_id").
			Where("au.type = ? AND au.at_williams = ?", "student", true).
			Where("ap.deleted_at IS NULL").
			Where(&models.EphmatchLike{
				LikedID: user.ID,
			}).
			Find(&usrAdmi).Error
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Printf("%s (L:%d, A:%d) - %f\n", user.Name, len(usrLikes), len(usrAdmi), rankMap[id])
	}

	for _, like := range likes {
		g.Edge(g.Node(strconv.Itoa(int(like.UserID))), g.Node(strconv.Itoa(int(like.LikedID))))
	}

	fmt.Println(g.String())
}
