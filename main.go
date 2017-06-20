package main

import (
	"context"
	"encoding/json"
	"time"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	influxcli "github.com/influxdata/influxdb/client/v2"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	svctx := context.Background()
	sv, _ := cli.ServerVersion(svctx)
	daemonOSType := sv.Os

	MyDB := os.Getenv("INFLUX_DBNAME")
	username := os.Getenv("INFLUX_USERNAME")
	password := os.Getenv("INFLUX_PASSWORD")
	influxhost := os.Getenv("INFLUX_HOST")

	c, err := influxcli.NewHTTPClient(influxcli.HTTPConfig{
		Addr:     influxhost,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}

	bp, err := influxcli.NewBatchPoints(influxcli.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})

	for range time.Tick(5000 * time.Millisecond) {

		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}

		for _, container := range containers {
			container_json, _ := cli.ContainerInspect(context.Background(), container.ID[:10])

			servicename := container.ID

			for _, str := range container_json.Config.Env {
				envarr := strings.Split(str, "=")
				if strings.EqualFold("SERVICE_NAME", envarr[0]) {
					servicename = envarr[1]
					break
				}

			}

			resp, _ := cli.ContainerStats(context.Background(), container.ID[:10], false)
			dec := json.NewDecoder(resp.Body)
			var (
				v *types.StatsJSON
				//memPercent       = 0.0
				cpuPercent = 0.0
				mem        = 0.0
				memLimit   = 0.0
				//memPerc          = 0.0
				//pidsStatsCurrent uint64
			)
			if err := dec.Decode(&v); err != nil {
				dec = json.NewDecoder(io.MultiReader(dec.Buffered(), resp.Body))
				if err == io.EOF {
					break
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}
			daemonOSType = resp.OSType

			if daemonOSType != "windows" {
				// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
				// got any data from cgroup
				//if v.MemoryStats.Limit != 0 {
				//	memPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
				//}
				previousCPU := v.PreCPUStats.CPUUsage.TotalUsage
				previousSystem := v.PreCPUStats.SystemUsage
				cpuPercent = calculateCPUPercentUnix(previousCPU, previousSystem, v)
				mem = float64(v.MemoryStats.Usage)
				memLimit = float64(v.MemoryStats.Limit)
				//memPerc = memPercent
				//pidsStatsCurrent = v.PidsStats.Current
			}
			netRx, netTx := calculateNetwork(v.Networks)

			// Create a point and add to batch
			tags := map[string]string{"host": servicename}
			fields := map[string]interface{}{
				"hostname":   container_json.Config.Hostname,
				"cpuPercent": cpuPercent,
				"mem":        mem,
				"memLimit":   memLimit,
				"netRx":      netRx,
				"netTx":      netTx,
			}

			pt, err := influxcli.NewPoint("dp-docker-stats", tags, fields, time.Now())
			if err != nil {
				log.Fatal(err)
			}
			bp.AddPoint(pt)

			// Write the batch
			if err := c.Write(bp); err != nil {
				log.Fatal(err)
			}

		}
	}
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
}
