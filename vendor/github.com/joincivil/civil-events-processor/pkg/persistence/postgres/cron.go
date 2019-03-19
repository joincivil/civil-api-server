package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
)

const (
	// TimestampDataType is the value for a persisted timestamp in the cron table
	TimestampDataType = "timestamp"
	// EventHashesDataType is the value for persisted event hashes for timestamp in the cron table
	EventHashesDataType = "event_hashes"
	// DataPersistedModelName is the string name of DataPersisted field in CronData
	DataPersistedModelName = "DataPersisted"
	// CronTableBaseName is the base name of table this code defines
	CronTableBaseName = "cron"
)

// CreateCronTableQuery returns the query to create this table
func CreateCronTableQuery(tableName string) string {
	queryString := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s(data_persisted TEXT, data_type TEXT);
    `, tableName)
	return queryString
}

// CronData contains all the information related to cronjob that needs to be persisted in cron DB.
type CronData struct {
	DataPersisted string `db:"data_persisted"`
	DataType      string `db:"data_type"`
}

// NewCronData creates a CronData model for DB with data to save
func NewCronData(dataPersisted string, dataType string) *CronData {
	return &CronData{DataPersisted: dataPersisted, DataType: dataType}
}
