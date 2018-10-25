package postgres // import "github.com/joincivil/civil-events-processor/pkg/persistence/postgres"

import (
	"fmt"
	"strconv"
)

const (
	// TimestampDataType is the value for data type of a persisted timestamp for the cron table
	TimestampDataType = "timestamp"
	// DataPersistedModelName is the string name of DataPersisted field in CronData
	DataPersistedModelName = "DataPersisted"

	defaultCronTableName = "cron"
)

// CreateCronTableQuery returns the query to create the cron table
func CreateCronTableQuery() string {
	return CreateCronTableQueryString(defaultCronTableName)
}

// CreateCronTableQueryString returns the query to create this table
func CreateCronTableQueryString(tableName string) string {
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

// TimestampToString converts an int64 timestamp to string
func TimestampToString(timestamp int64) string {
	return strconv.FormatInt(timestamp, 10)
}

// StringToTimestamp converts a string timestamp to int64
func StringToTimestamp(timestamp string) (int64, error) {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return i, fmt.Errorf("Could not convert timestamp from string to int64: %v", err)
	}
	return i, nil
}
