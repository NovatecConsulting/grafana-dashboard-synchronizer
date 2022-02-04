package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/NovatecConsulting/grafana-dashboard-sync/pkg/internal"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "configuration.yml",
				Usage:   "the configuration file to use",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "performs a dry run without actually importing or exporting dashboards",
			},
			&cli.BoolFlag{
				Name:  "log-as-json",
				Usage: "printing logs as structured json objects",
			},
		},
		Action: func(c *cli.Context) error {
			return synchronizeDashboards(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Starts the synchronization of the Grafana dashboards.
func synchronizeDashboards(c *cli.Context) error {
	// setup logger
	if c.Bool("log-as-json") {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
	}

	log.Info("Synchronizing Grafana dashboards...")

	if c.Bool("dry-run") {
		log.Info("DRY-RUN : The application will NOT perform any changes to Git or Grafana due to the fry-run flag!")
	}

	// read configuration
	input, err := readConf(c.String("config"))
	if err != nil {
		log.WithField("error", err).Fatal("Error while reading configuration file.")
		return err
	}

	// do the synchronization
	for _, element := range *input {
		synchronizer := internal.NewSynchronizer(element)
		synchronizer.Synchronize(c.Bool("dry-run"))
	}

	log.Info("Synchronization completed.")
	return nil
}

// Reads the given file and parses it into a struct representing the configuration to use.
func readConf(filename string) (*[]internal.SynchronizeOptions, error) {
	log.WithField("file", filename).Info("Reading configuration file...")

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
