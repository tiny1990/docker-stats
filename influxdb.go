package main

import (
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"time"
	"fmt"
)

type InfluxDB struct {
	host     string
	dbname   string
	username string
	password string
}

func SendToDB(tableName string, cli client.Client, points client.BatchPoints, stat []*DockerStat) {
	for _, dockerStat := range stat {
		tags := map[string]string{"host": dockerStat.serviceName}
		fields := map[string]interface{}{
			"cpuPercent": dockerStat.cpuPercent,
			"memPercent": dockerStat.memPercent,
			"memUsed":    dockerStat.menUsed,
			"memLimit":   dockerStat.memLimit,
			"netRx":      dockerStat.netRx,
			"netTx":      dockerStat.netTx,
		}
		pt, err := client.NewPoint(tableName, tags, fields, time.Now())
		if err != nil {
			log.Println(err)
			continue
		}
		points.AddPoint(pt)
	}
	cli.Write(points)
	size := len(points.Points())
	fmt.Println(size)

	cli.Close()
}

func (*InfluxDB) InitDB(host, dbname, username, password string) (client.Client, client.BatchPoints) {
	influxClient, httpConfigErr := client.NewHTTPClient(client.HTTPConfig{
		Addr:     host,
		Username: username,
		Password: password,
	})
	if httpConfigErr != nil {
		log.Fatal(httpConfigErr)
	}
	batchPointsConfig, batchPointErr := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbname,
		Precision: "s",
	})

	if batchPointErr != nil {
		log.Fatal(batchPointErr)
	}
	return influxClient, batchPointsConfig
}
