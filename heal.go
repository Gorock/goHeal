package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	dockerSock   = "DOCKER_SOCK"
	healStart    = "HEAL_START_PERIOD"
	healInterval = "HEAL_INTERVAL"
)

type container struct {
	Id string
}

func restartContainers(url string) {
	containerUrl := fmt.Sprintf("%s/v1.44/containers", url)
	resp, err := http.Get(containerUrl + "/json?health=unhealthy")
	if err != nil {
		log.Fatalf("http:get %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	_ = resp.Body.Close()
	var containers []container
	err = json.Unmarshal(body, &containers)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", containers)

	for _, container := range containers {
		var s = fmt.Sprintf("%s/%s/restart", containerUrl, container.Id)
		req, err := http.Post(s, "text/plain", nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("\n%s,%d", container.Id, req.StatusCode)
	}

}
func getEnvDuration(def time.Duration, env string) time.Duration {
	val, ok := os.LookupEnv(env)
	if ok {
		dval, err := time.ParseDuration(val)
		if err != nil {
			panic(err)
		}
		return dval
	}
	return def
}

func main() {
	dockerSocket := "http://localhost:3001"
	interval := getEnvDuration(5.0*(time.Second), healInterval)
	startPeriod := getEnvDuration(time.Duration(0), healStart)
	val, ok := os.LookupEnv(dockerSock)
	if ok {
		dockerSocket = val
	}
	time.Sleep(startPeriod)
	for {
		restartContainers(dockerSocket)
		time.Sleep(interval)
	}
}
