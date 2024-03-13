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
	healLabel    = "HEAL_LABEL"
)

type container struct {
	Id string
}

type containerFilter struct {
	Health []string `json:"health"`
	Label  []string `json:"label"`
	Status []string `json:"status"`
}

func restartContainers(url string, containerFilter string) {
	containerUrl := fmt.Sprintf("%s/v1.44/containers", url)

	resp, err := http.Get(fmt.Sprintf("%s/json?filters=%s", containerUrl, containerFilter))
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
	label := "health"
	x := containerFilter{
		Health: []string{"unhealthy"},
		Label:  []string{label},
		Status: []string{"running"},
	}
	j, _ := json.Marshal(x)
	print(string(j))
	val, ok := os.LookupEnv(dockerSock)
	if ok {
		dockerSocket = val
	}

	val, ok = os.LookupEnv(healLabel)
	if ok {
		label = val
	}

	time.Sleep(startPeriod)
	for {
		restartContainers(dockerSocket, label)
		time.Sleep(interval)
	}
}
