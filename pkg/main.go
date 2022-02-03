package main

import (
	"github.com/NovatecConsulting/grafana-dashboard-sync/pkg/internal"
	log "github.com/sirupsen/logrus"
)

func main() {
	// setup logger
	log.SetFormatter(&log.JSONFormatter{})

	log.Info("Synchronizing Grafana dashboards...")

	pushConfiguration := internal.PushConfiguration{
		PushTags:   true,
		TagPattern: "agent",
		PullConfiguration: internal.PullConfiguration{
			Enable:    false,
			GitBranch: "standalone",
			Filter:    "",
		},
	}

	pullConfiguration := internal.PullConfiguration{
		Enable:    true,
		GitBranch: "standalone",
		Filter:    "",
	}

	options := internal.SynchronizeOptions{
		JobName:      "test-job",
		GrafanaToken: "eyJrIjoiSEp4dzhGdVBxMUhBdm5Dbkxhdnd0b2Rzbm1wS3laTjMiLCJuIjoidGVzdCIsImlkIjoxfQ==",
		GrafanaUrl:   "http://localhost:3000",

		GitRepositoryUrl: "git@github.com:mariusoe/config-push.git",
		PrivateKeyFile:   "/Users/mo/.ssh/id_ed25519",

		PushConfiguration: pushConfiguration,
		PullConfiguration: pullConfiguration,
	}

	synchronizer := internal.NewSynchronizer(options)

	synchronizer.Synchronize()

	// log.DefaultLogger.Debug("Synchronizing Grafana dashboards..")
	// Start listening to requests sent from Grafana. This call is blocking so
	// it won't finish until Grafana shuts down the process or the plugin choose
	// to exit by itself using os.Exit. Manage automatically manages life cycle
	// of datasource instances. It accepts datasource instance factory as first
	// argument. This factory will be automatically called on incoming request
	// from Grafana to create different instances of SampleDatasource (per datasource
	// ID). When datasource configuration changed Dispose method will be called and
	// new datasource instance created using NewSynchronizeDatasource factory.

	//if err := datasource.Manage("novatec-dashboardsync-datasource", plugin.NewSynchronizeDatasource, datasource.ManageOpts{}); err != nil {
	//	log.DefaultLogger.Error(err.Error())
	//	os.Exit(1)
	//}
}
