package main

import (
	"fmt"
	"io/ioutil"

	"github.com/NovatecConsulting/grafana-dashboard-sync/pkg/internal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func main() {
	// setup logger
	log.SetFormatter(&log.JSONFormatter{})

	log.Info("Synchronizing Grafana dashboards...")

	input, err := readConf("../test.yml")
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, element := range *input {
		synchronizer := internal.NewSynchronizer(element)

		synchronizer.Synchronize()
	}
}

func readConf(filename string) (*[]internal.SynchronizeOptions, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &[]internal.SynchronizeOptions{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return c, nil
}
