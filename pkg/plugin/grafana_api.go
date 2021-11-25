package plugin

import (
	"context"
	"github.com/grafana-tools/sdk"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// GrafanaApi access to grafana api
type GrafanaApi struct {
	grafanaClient *sdk.Client
}

// NewGrafanaApi creates a new GrafanaApi instance
func NewGrafanaApi(grafanaURL string, apiToken string) GrafanaApi {
	client, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)
    grafanaApi := GrafanaApi {client}
    log.DefaultLogger.Info("Grafana API Client created")
	return grafanaApi
}

// SearchDashboardsWithTag returns all dashboards with the given tag
func (grafanaApi GrafanaApi) SearchDashboardsWithTag(tag string) ([]sdk.FoundBoard, error) {
	searchParam := sdk.SearchTag(tag)
	foundDashboards, err := grafanaApi.grafanaClient.Search(context.Background(), searchParam)
	return foundDashboards, err
}

// GetRawDashboardByID return Dashboard by the given UID as raw byte object
func (grafanaApi GrafanaApi) GetRawDashboardByID(uid string) ([]byte, sdk.BoardProperties,  error) {
	rawDashboard, props, err := grafanaApi.grafanaClient.GetRawDashboardByUID(context.Background(), uid)
	return rawDashboard, props, err
}

// GetDashboardObjectByID return Dashboard by the given UID as object
func (grafanaApi GrafanaApi) GetDashboardObjectByID(uid string) (sdk.Board, sdk.BoardProperties,  error) {
	dashboardObject, props, err := grafanaApi.grafanaClient.GetDashboardByUID(context.Background(), uid)
	return dashboardObject, props, err
}

// UpdateDashboardObjectByID update the Dashboard with the given dashboard object
func (grafanaApi GrafanaApi) UpdateDashboardObjectByID(dashboard sdk.Board) (sdk.StatusMessage, error) {
	statusMessage, err := grafanaApi.grafanaClient.SetDashboard(context.Background() ,dashboard, sdk.SetDashboardParams{
			Overwrite: false,
		})
	return statusMessage, err
}

// CreateDashboardObjects set a Dashboard with the given raw dashboard object
func (grafanaApi GrafanaApi) CreateDashboardObjects(fileMap map[string][]byte) {
	for dashboardName, rawDashboard := range fileMap {
		_, err := grafanaApi.grafanaClient.SetRawDashboard(context.Background(), rawDashboard)
		if err != nil {
			log.DefaultLogger.Error("set dashboard error", "error", err.Error())
		}
		log.DefaultLogger.Info("Dashboard created", "name", dashboardName)
	}
}

// DeleteTagFromDashboardObjectByID delete the given tag from the Dashboard object
func (grafanaApi GrafanaApi) DeleteTagFromDashboardObjectByID(dashboard sdk.Board, tag string) sdk.Board {
	for i, iTag := range dashboard.Tags {
		if iTag == tag {
			dashboard.Tags = append(dashboard.Tags[:i], dashboard.Tags[i+1:]...)
			break
		}
	}
	return dashboard
}

