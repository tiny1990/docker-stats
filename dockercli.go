package main

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"context"
	"strings"
	"encoding/json"
)

type DockerStat struct {
	serviceName string
	netRx       float64
	netTx       float64
	menUsed     float64
	memLimit    float64
	memPercent  float64
	cpuPercent  float64
}

func GetDockerStat() []*DockerStat {
	stats := []*DockerStat{}
	cli, err := client.NewEnvClient()
	defer cli.Close()

	if err != nil {
		log.Println(err)
		return stats
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Println(err)
		return  stats
	}
	for _, container := range containers {
		inspectJson, err := cli.ContainerInspect(context.Background(), container.ID[:10])
		if err != nil {
			log.Println(err)
			continue
		}
		dockerstat := new(DockerStat)
		dockerstat.serviceName = container.ID[:10]

		for _, envStr := range inspectJson.Config.Env {
			env := strings.Split(envStr, "=")
			if "SERVICE_NAME" == env[0] {
				dockerstat.serviceName = env[1]
				break
			}
		}
		statJson, _ := cli.ContainerStats(context.Background(), container.ID[:10], false)
		dec := json.NewDecoder(statJson.Body)
		var v *types.StatsJSON

		if err := dec.Decode(&v); err != nil {
			//do what???
			log.Println(err)
			continue
		}
		daemonOsType := statJson.OSType
		// do not support windows
		if daemonOsType != "windows" {
			if v.MemoryStats.Limit != 0 {
				dockerstat.memPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
			}
			dockerstat.cpuPercent = calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v)
			dockerstat.menUsed = float64(v.MemoryStats.Usage)
			dockerstat.memLimit = float64(v.MemoryStats.Limit)
			dockerstat.netRx, dockerstat.netTx = calculateNetwork(v.Networks)
		}
		stats = append(stats, dockerstat)

	}
	return stats

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
