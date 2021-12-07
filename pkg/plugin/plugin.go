package plugin

import (
	"context"
	"encoding/json"

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

	var uiProperties SynchronizeOptions
	_ = json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &uiProperties)
	uiSecureProperties := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData

	var status = backend.HealthStatusOk
	var message = "Data source is working yeah"

	// TODO Git health check
	// random error disabled:
	//if rand.Int()%2 == 0 {
	//	status = backend.HealthStatusError
	//	message = "randomized error"
	//}

	// TODO: Set workflow cron job?

	grafanaUrl := uiProperties.GrafanaUrl
	token := uiSecureProperties["grafanaApiToken"]

	gitUrl := uiProperties.GitUrl
	dashboardTag := uiProperties.PushConfiguration.TagPattern
	privateKey := []byte(uiSecureProperties["privateSshKey"])

	// privateKeyFilePath := uiSecureProperties["privateKeyFilePath"]

	grafanaApi := NewGrafanaApi(grafanaUrl, token)

	gitApi := NewGitApi(uiProperties.GitUrl, privateKey)
	log.DefaultLogger.Info("Using Git repository from: %s", uiProperties.GitUrl)

	repository, err := gitApi.CloneRepo()
	if err != nil {
		return nil, err
	}
	gitApi.FetchRepo(*repository)

	if uiProperties.PullConfiguration.Enable {
		log.DefaultLogger.Info("Pull from git repo", "url", gitUrl)

		gitApi.PullRepo(*repository)
		fileMap := gitApi.GetFileContent()
		grafanaApi.CreateDashboardObjects(fileMap)
		log.DefaultLogger.Info("Dashboards created")
	}

	if uiProperties.PushConfiguration.Enable {
		log.DefaultLogger.Info("Push to git repo", "url", gitUrl)

		dashboards, err := grafanaApi.SearchDashboardsWithTag(dashboardTag)
		if err != nil {
			log.DefaultLogger.Error("search dashboard", "error", err.Error())
		}
		for _, dashboard := range dashboards {
			// get dashboard Object and Properties
			dashboardObject, boardProperties, err := grafanaApi.GetDashboardObjectByID(dashboard.UID)
			if err != nil {
				log.DefaultLogger.Error("get dashboard", "error", err.Error())
			}

			// delete Tag from dashboard Object
			dashboardWithDeletedTag := grafanaApi.DeleteTagFromDashboardObjectByID(dashboardObject, dashboardTag)

			// get folder name and id, required for update processes and git folder structure
			folderName := boardProperties.FolderTitle
			folderId := boardProperties.FolderID

			// update dashboard with deleted Tag in Grafana
			_, err = grafanaApi.UpdateDashboardObjectByID(dashboardWithDeletedTag, folderId)
			if err != nil {
				log.DefaultLogger.Error("update dashboard", "error", err.Error())
			}

			// get raw Json Dashboard, required for import and export
			dashboardJson, _, err := grafanaApi.GetRawDashboardByID(dashboard.UID)
			if err != nil {
				log.DefaultLogger.Error("get raw dashboard", "error", err.Error())
			}

			log.DefaultLogger.Debug("Dashboard preparation successfully ")
			gitApi.AddFileWithContent(folderName+"/"+dashboardObject.Title+".json", string(dashboardJson))
			log.DefaultLogger.Debug("Dashboard added to in memory file system")
		}

		if len(dashboards) > 0 {
			gitApi.CommitWorktree(*repository, dashboardTag)
			gitApi.PushRepo(*repository)
		}

		log.DefaultLogger.Info("Dashboard pushed successfully")
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
