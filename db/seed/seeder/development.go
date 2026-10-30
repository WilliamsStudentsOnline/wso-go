package seeder

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/db/seed"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Development fills a fresh migrated DB with a small, synthetic dataset for local
// devenv. Idempotent: no-ops when users already exist.
func Development(db *gorm.DB) error {
	n, err := UserCount(db)
	if err != nil {
		return err
	}
	if n > 0 {
		zap.S().Infof("development seed: skipping (%d users already present)", n)
		return nil
	}

	zap.S().Info("development seed: users")
	if err := Users(0, db); err != nil {
		return fmt.Errorf("users: %w", err)
	}

	zap.S().Info("development seed: dorms")
	if err := Dorms(4, db); err != nil {
		return fmt.Errorf("dorms: %w", err)
	}

	zap.S().Info("development seed: board threads")
	if err := BoardThreads(12, db); err != nil {
		return fmt.Errorf("board: %w", err)
	}

	zap.S().Info("development seed: ephmatch profiles")
	if err := EphmatchProfiles(16, db); err != nil {
		return fmt.Errorf("ephmatch profiles: %w", err)
	}

	zap.S().Info("development seed: done")
	return nil
}

// Dorms inserts n synthetic dorms (no rooms).
func Dorms(n int, db *gorm.DB) error {
	if n <= 0 {
		n = 4
	}
	for i := 0; i < n; i++ {
		d := seed.GenerateDorm()
		d.Name = fmt.Sprintf("Seed Dorm %d", i+1)
		if err := db.Create(d).Error; err != nil {
			return err
		}
	}
	return nil
}

// BoardThreadCount is used by devenv-oriented checks.
func BoardThreadCount(db *gorm.DB) (int, error) {
	var count int
	err := db.Model(&models.BoardThread{}).Count(&count).Error
	return count, err
}
