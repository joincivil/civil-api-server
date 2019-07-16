package testutils

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"

	// load postgres specific dialect
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DBCreds struct {
	Port     int
	Dbname   string
	User     string
	Password string
	Host     string
}

// GetTestDBCreds returns the credentials for the local docker instance
// dependent on env vars.
func GetTestDBCreds() DBCreds {
	var creds DBCreds
	if os.Getenv("CI") == "true" {
		creds = DBCreds{
			Port:     5432,
			Dbname:   "circle_test",
			User:     "root",
			Password: "root",
			Host:     "localhost",
		}
	} else {
		creds = DBCreds{
			Port:     5432,
			Dbname:   "civil_crawler",
			User:     "docker",
			Password: "docker",
			Host:     "localhost",
		}
	}
	return creds
}

// GetTestDBConnection returns a new gorm Database connection for the local docker instance
func GetTestDBConnection() (*gorm.DB, error) {
	creds := GetTestDBCreds()

	connStr := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable",
		creds.Host, creds.Port, creds.User, creds.Dbname, creds.Password)
	fmt.Printf("Connecting to database: %v\n", connStr)
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error opening database connection:: err: %v", err)
	}

	db.LogMode(true)
	return db, err
}
