package seeder

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/db/seed"
	"github.com/jinzhu/gorm"
)

// Dorms inserts n synthetic dorms (no rooms). Optional; local devenv normally
// loads db/seed/development.sql.gz instead.
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
