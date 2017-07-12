package main

import (
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"time"
)

type InfluxDB struct {
	host     string
	dbname   string
	username string
	password string
}

func SendToDB(dbName, tableName string, cli client.Client, stat []*DockerStat) {

	batchPoints, batchPointErr := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbName,
		Precision: "s",
	})

	if batchPointErr != nil {
		log.Println(batchPointErr)
		return
	}

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
		batchPoints.AddPoint(pt)
	}
	err := cli.Write(batchPoints)
	if err != nil {
		log.Println(err)
	}
	log.Println(len(batchPoints.Points()))
	cli.Close()
}

func (*InfluxDB) InitDB(host, dbname, username, password string) (client.Client) {
	influxClient, httpConfigErr := client.NewHTTPClient(client.HTTPConfig{
		Addr:     host,
		Username: username,
		Password: password,
	})
	if httpConfigErr != nil {
		log.Fatal(httpConfigErr)
	}
	return influxClient
}
