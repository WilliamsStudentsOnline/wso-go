package users

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/jinzhu/gorm"
)

func SearchUsersMySQL(query string, db *gorm.DB) (*gorm.DB, error) {
	ast, err := search.ParseSearchQuery(query)
	if err != nil {
		return nil, err
	}

	db = recursiveSearchMySQL(ast, db)

	return db, nil
}

func recursiveSearchMySQL(ast *search.Query, db *gorm.DB) *gorm.DB {
	return constructOrMySQL(ast.Or, db)
}

func constructOrMySQL(exps []*search.Expression, db *gorm.DB) *gorm.DB {
	return db
}

func constructAndMySQL(conds []*search.Condition, db *gorm.DB) *gorm.DB {
	return db
}
