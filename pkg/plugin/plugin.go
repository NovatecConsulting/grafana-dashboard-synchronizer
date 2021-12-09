package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.CheckHealthHandler    = (*SampleDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SampleDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// TODO; initial aufgerufen??
	return &SampleDatasource{}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SampleDatasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SampleDatasource) Dispose() {
	// Clean up datasource instance resources.
}

type PullConfiguration struct {
	Enable       bool   `json:"enable"`
	GitBranch    string `json:"gitBranch"`
	SyncInterval int64  `json:"syncInterval"`
	Filter       string `json:"filter"`
}

type PushConfiguration struct {
	PullConfiguration
	TagPattern string `json:"tagPattern"`
	PushTags   bool   `json:"pushTags"`
}

type SynchronizeOptions struct {
	GrafanaUrl        string            `json:"grafanaUrl"`
	GitUrl            string            `json:"gitUrl"`
	PushConfiguration PushConfiguration `json:"pushConfiguration"`
	PullConfiguration PullConfiguration `json:"pullConfiguration"`
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Debug("Backend called with following request", "request", req)

	var properties SynchronizeOptions
	_ = json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &properties)
	secureProperties := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData

	// TODO: Set workflow cron job?

	grafanaToken := secureProperties["grafanaApiToken"]
	privateKey := []byte(secureProperties["privateSshKey"])

	dashboardTag := properties.PushConfiguration.TagPattern

	grafanaApi := NewGrafanaApi(properties.GrafanaUrl, grafanaToken)
	gitApi := NewGitApi(properties.GitUrl, privateKey)

	// Push
	if properties.PushConfiguration.Enable {
		log.DefaultLogger.Info("Push to git repo", "url", properties.GitUrl)

		dashboards, err := grafanaApi.SearchDashboardsWithTag(dashboardTag)
		if err != nil {
			log.DefaultLogger.Error("search dashboard", "error", err.Error())
		}
		for _, dashboard := range dashboards {
			// get dashboard Object and Properties
			dashboardObject, boardProperties := grafanaApi.GetDashboardObjectByUID(dashboard.UID)

			// delete Tag from dashboard Object
			dashboardWithDeletedTag := grafanaApi.DeleteTagFromDashboardObjectByID(dashboardObject, dashboardTag)

			// get folder name and id, required for update processes and git folder structure
			folderName := boardProperties.FolderTitle
			folderId := boardProperties.FolderID

			// get raw Json Dashboard, required for import and export
			dashboardJson, err := json.Marshal(DashboardWithCustomFields{dashboardWithDeletedTag, req.PluginContext.DataSourceInstanceSettings.Name})
			if err != nil {
				log.DefaultLogger.Error("get raw dashboard", "error", err.Error())
			}

			// update dashboard with deleted Tag in Grafana
			grafanaApi.CreateOrUpdateDashboardObjectByID(dashboardJson, folderId, fmt.Sprintf("Deleted '%s' tag", dashboardTag))
			log.DefaultLogger.Debug("Dashboard preparation successfully")

			// Add Dashboard to in memory file system
			gitApi.AddFileWithContent(folderName+"/"+dashboardObject.Title+".json", string(dashboardJson))
			log.DefaultLogger.Debug("Dashboard added to in memory file system")
		}

		if len(dashboards) > 0 {
			// clone repo from specific branch
			repository, err := gitApi.CloneRepo(properties.PushConfiguration.GitBranch)
			if err != nil {
				return nil, err
			}

			gitApi.CommitWorktree(*repository, dashboardTag)
			gitApi.PushRepo(*repository)
			log.DefaultLogger.Info("Dashboards pushed successfully")
		}
	}

	// Pull
	if properties.PullConfiguration.Enable {
		errorResult := d.PullDashboards(grafanaApi, gitApi, properties.GitUrl, properties.PullConfiguration.GitBranch)
		if errorResult != nil {
			return errorResult, nil
		}
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Dashboards have been synchronized",
	}, nil
}

func (d *SampleDatasource) PullDashboards(grafanaApi GrafanaApi, gitApi GitApi, repositoryUrl string, branch string) *backend.CheckHealthResult {
	log.DefaultLogger.Info("Pulling and importing dashboards", "repositoryUrl", repositoryUrl, "branch", branch)

	// clone and fetch repo from specific branch
	log.DefaultLogger.Debug("Cloning repository")
	repository, err := gitApi.CloneRepo(branch)
	if err != nil {
		log.DefaultLogger.Error("Cloning repository failed", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Could not clone specified Git repository",
		}
	}

	commitId, err, errMessage := gitApi.GetLatestCommitId(*repository)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Error getting latest commit: " + errMessage,
		}
	}

	fileMap := gitApi.GetFileContent()
	grafanaApi.CreateOrUpdateDashboard(fileMap, commitId)

	log.DefaultLogger.Info("Successfully synchronized dashboards from Git repositroy")

	return nil
}
