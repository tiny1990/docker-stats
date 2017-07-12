package main

import (
	"time"
	"os"
)

var INFLUX_HOST string
var INFLUX_DBNAME string
var INFLUX_USERNAME string
var INFLUX_PASSWORD string
var INFLUX_TABLE_SUFFIX string
var INFLUX_TABLE_NAME = "dp-docker-stats"

func init() {
	INFLUX_HOST = os.Getenv("INFLUX_HOST")
	INFLUX_DBNAME = os.Getenv("INFLUX_DBNAME")
	INFLUX_USERNAME = os.Getenv("INFLUX_USERNAME")
	INFLUX_PASSWORD = os.Getenv("INFLUX_PASSWORD")
	INFLUX_TABLE_SUFFIX = os.Getenv("INFLUX_TABLE_SUFFIX")
	if INFLUX_TABLE_SUFFIX != "" {
		INFLUX_TABLE_NAME = INFLUX_TABLE_NAME + "-" + INFLUX_TABLE_SUFFIX
	}

}

func main() {

	influxDB := new(InfluxDB)
	cli, batchPoints := influxDB.InitDB(INFLUX_HOST, INFLUX_DBNAME, INFLUX_USERNAME, INFLUX_PASSWORD)

	for range time.Tick(5000 * time.Millisecond) {
		dockerstat := GetDockerStat()
		SendToDB(INFLUX_TABLE_NAME, cli, batchPoints, dockerstat)
	}

}
