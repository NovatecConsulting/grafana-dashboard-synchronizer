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
func (grafanaApi GrafanaApi) UpdateDashboardObjectByID(dashboard sdk.Board, folderId int) (sdk.StatusMessage, error) {
	statusMessage, err := grafanaApi.grafanaClient.SetDashboard(context.Background() ,dashboard, sdk.SetDashboardParams{
			Overwrite: false,
			FolderID: folderId,
		})
	return statusMessage, err
}

// CreateFolder create a folder in Grafana
func (grafanaApi GrafanaApi) CreateFolder(folderName string) int {
	folder := sdk.Folder{Title: folderName}
	folder, err := grafanaApi.grafanaClient.CreateFolder(context.Background(), folder)
	if err != nil && folderName != "General" {
		log.DefaultLogger.Error("get folders error", "error", err.Error())
	}
	return folder.ID
}

// GetOrCreateFolderID returns the ID of a given folder or create it
func (grafanaApi GrafanaApi) GetOrCreateFolderID(folderName string) int {
	folders, err := grafanaApi.grafanaClient.GetAllFolders(context.Background())
	if err != nil {
		log.DefaultLogger.Error("get all folders error", "error", err.Error())
	}
	for _, folder := range folders {
		if folder.Title == folderName {
			return folder.ID
		}
	}
	generatedFolderID := grafanaApi.CreateFolder(folderName)
	return generatedFolderID
}

// CreateDashboardObjects set a Dashboard with the given raw dashboard object
func (grafanaApi GrafanaApi) CreateDashboardObjects(fileMap map[string]map[string][]byte) {
	for dashboardDir, dashboardFile := range fileMap {
		dirID := grafanaApi.GetOrCreateFolderID(dashboardDir)
		for dashboardName, rawDashboard := range dashboardFile {
			_, err := grafanaApi.grafanaClient.SetRawDashboardWithParam(context.Background(), sdk.RawBoardRequest{
				Dashboard: rawDashboard,
				Parameters: sdk.SetDashboardParams{
					Overwrite: false,
					FolderID:  dirID,
				},
			})
			if err != nil {
				log.DefaultLogger.Error("set dashboard error", "error", err.Error())
			}
			log.DefaultLogger.Debug("Dashboard created", "name", dashboardName)
		}
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

