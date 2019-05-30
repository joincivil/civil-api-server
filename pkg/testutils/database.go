package testutils

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/posts"
	// load postgres specific dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type dbcreds struct {
	Port     int
	Dbname   string
	User     string
	Password string
	Host     string
}

// GetTestDBConnection returns a new gorm Database connection for the local docker instance
func GetTestDBConnection() (*gorm.DB, error) {

	var creds dbcreds
	if os.Getenv("CI") == "true" {
		creds = dbcreds{
			Port:     5432,
			Dbname:   "circle_test",
			User:     "root",
			Password: "root",
			Host:     "localhost",
		}
	} else {
		creds = dbcreds{
			Port:     5432,
			Dbname:   "civil_crawler",
			User:     "docker",
			Password: "docker",
			Host:     "localhost",
		}
	}

	connStr := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", creds.Host, creds.Port, creds.User, creds.Dbname, creds.Password)
	fmt.Printf("Connecting to database: %v\n", connStr)
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error opening database connection:: err: %v", err)
	}

	db.LogMode(true)
	db.AutoMigrate(&posts.Base{})

	return db, err
}
