package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/disintegration/imaging"
)

func main() {
	fmt.Println("HHJHIAWfbneriuyvnjkwrb")
	var inDir string
	var configPath string
	var console bool

	flag.StringVar(&inDir, "in", "", "input directory")
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
	log, err := logging.SetupLog(cfg, "frosh-photos")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
		return
	}
	defer log.Sync()

	pb, err := pictures.NewPictureBackend(cfg, log)
	if err != nil {
		log.Fatal("New Picture Backend: " + err.Error())
	}

	in, err := filepath.Abs(inDir)
	if err != nil {
		log.Fatal("Filepath: " + err.Error())
	}

	files, err := os.ReadDir(in)
	if err != nil {
		log.Fatal(err)
	}

	// run for every photo
	for _, file := range files {
		unix := strings.TrimSuffix(file.Name(), ".jpg")

		// skip if photo already exists (e.g. if the user has already uploaded)
		exists, existErr := pb.DoesUserPhotoExists(unix)
		if exists {
			log.Warnf("Skipping %s, already has photo.", unix)
		}
		if existErr != nil {
			log.Warnf("%s error checking existence: %v", unix, existErr)
			continue
		}

		log.Infof("saving %s", unix)
		saveErr := savePhoto(filepath.Join(in, file.Name()), unix, pb)
		if saveErr != nil {
			log.Warnf("%s error saving: %v", unix, saveErr)
		}
	}
}

func savePhoto(path string, unix string, pb pictures.PictureBackend) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	var dualSaveWg sync.WaitGroup
	errors := make(chan error, 2)

	dualSaveWg.Add(1)
	go func(wg *sync.WaitGroup) {
		imgScaled := imaging.Fill(img, 300, 300, imaging.Center, imaging.Lanczos)
		err = pb.SaveUserPhotoLarge(unix, imgScaled)
		if err != nil {
			errors <- err
		}
		wg.Done()
	}(&dualSaveWg)

	dualSaveWg.Add(1)
	go func(wg *sync.WaitGroup) {
		imgThumb := imaging.Fill(img, 50, 50, imaging.Center, imaging.Lanczos)

		err = pb.SaveUserPhotoThumb(unix, imgThumb)
		if err != nil {
			errors <- err
		}
		wg.Done()
	}(&dualSaveWg)

	dualSaveWg.Wait()
	close(errors)

	err = <-errors
	if err != nil {
		return err
	}

	return nil
}
