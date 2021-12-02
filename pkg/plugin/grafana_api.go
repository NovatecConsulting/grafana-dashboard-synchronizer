package plugin

import (
	"context"
	"encoding/json"
	"github.com/NovatecConsulting/grafana-api-go-sdk"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"strconv"
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

// getDashboardObjectFromRawDashboard get the dashboard object from a raw dashboard json
func getDashboardObjectFromRawDashboard(rawDashboard []byte) sdk.Board {
	var dashboard sdk.Board
	err := json.Unmarshal(rawDashboard, &dashboard)
	if err != nil {
		log.DefaultLogger.Error("unmarshal raw dashboard error", "error", err.Error())
	}
	return dashboard
}

// CreateDashboardObjects set a Dashboard with the given raw dashboard object
func (grafanaApi GrafanaApi) CreateDashboardObjects(fileMap map[string]map[string][]byte) {
	for dashboardDir, dashboardFile := range fileMap {
		dirID := grafanaApi.GetOrCreateFolderID(dashboardDir)
		for dashboardName, rawDashboard := range dashboardFile {
			gitDashboardObject := getDashboardObjectFromRawDashboard(rawDashboard)
			//// get grafana dashboard to compare version with git dashboard
			//grafanaDashboardObject, grafanaDashboardProperties, err := grafanaApi.GetDashboardObjectByID(gitDashboardObject.UID)
			//if err != nil {
			//	log.DefaultLogger.Error("get dashboard object error", "error", err.Error())
			//}
			// Todo: Extend Grafana tools fork to get version note
			//if grafanaDashboardObject.Version != gitDashboardObject.Version {
			_, err := grafanaApi.grafanaClient.SetRawDashboardWithParam(context.Background(), sdk.RawBoardRequest{
				Dashboard: rawDashboard,
				Parameters: sdk.SetDashboardParams{
					Overwrite: true,
					FolderID:  dirID,
					Message:   "Synchronized from Version " + strconv.Itoa(int(gitDashboardObject.Version)),
				},
			})
			if err != nil {
				log.DefaultLogger.Error("set dashboard error", "error", err.Error())
			}
			log.DefaultLogger.Debug("Dashboard created", "name", dashboardName)
			//} else {
			//	log.DefaultLogger.Info("Dashboard already up-to-date", "name", dashboardName)
			//}
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

