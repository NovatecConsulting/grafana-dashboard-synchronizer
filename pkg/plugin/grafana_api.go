package plugin

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	sdk "github.com/NovatecConsulting/grafana-api-go-sdk"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// GrafanaApi access to grafana api
type GrafanaApi struct {
	grafanaClient *sdk.Client
}

// NewGrafanaApi creates a new GrafanaApi instance
func NewGrafanaApi(grafanaURL string, apiToken string) GrafanaApi {
	client, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)
	grafanaApi := GrafanaApi{client}
	log.DefaultLogger.Info("Grafana API Client created")
	return grafanaApi
}

// SearchDashboardsWithTag returns all dashboards with the given tag
func (grafanaApi GrafanaApi) SearchDashboardsWithTag(tag string) ([]sdk.FoundBoard, error) {
	searchParam := sdk.SearchTag(tag)
	foundDashboards, err := grafanaApi.grafanaClient.Search(context.Background(), searchParam)
	return foundDashboards, err
}

// GetDashboardObjectByUID return Dashboard by the given UID as object
func (grafanaApi GrafanaApi) GetDashboardObjectByUID(uid string) (sdk.Board, sdk.BoardProperties) {
	dashboardObject, dashboardProperties, err := grafanaApi.grafanaClient.GetDashboardByUID(context.Background(), uid)
	if err != nil {
		dashboardNotFound := strings.Contains(err.Error(), "Dashboard not found")
		if !dashboardNotFound {
			log.DefaultLogger.Error("get dashboard object error", "error", err.Error())
		}
		return sdk.Board{}, sdk.BoardProperties{}
	}
	return dashboardObject, dashboardProperties
}

// CreateOrUpdateDashboardObjectByID create or update the Dashboard with the given dashboard object
func (grafanaApi GrafanaApi) CreateOrUpdateDashboardObjectByID(rawDashboard []byte, folderId int, message string) sdk.StatusMessage {
	statusMessage, err := grafanaApi.grafanaClient.SetRawDashboardWithParam(context.Background(), sdk.RawBoardRequest{
		Dashboard: rawDashboard,
		Parameters: sdk.SetDashboardParams{
			Overwrite: true,
			FolderID:  folderId,
			Message:   message,
		},
	})
	if err != nil {
		log.DefaultLogger.Error("set dashboard error", "error", err.Error())
	}
	return statusMessage
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

//// getLatestVersionsFromDashboardById get the latest version information of a dashboard by id
//func (grafanaApi GrafanaApi) getLatestVersionsFromDashboardById(dashboardId int) int {
//	versionResponse, err := grafanaApi.grafanaClient.GetAllDashboardVersions(context.Background(), dashboardId, 1)
//	if err != nil {
//		log.DefaultLogger.Error("get dashboard versions error", "error", err.Error())
//	}
//
//	// get version message
//	latestVersion := versionResponse.Versions[0]
//	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d]*`)
//	idFromVersionMessage := re.FindString(latestVersion.Message)
//	log.DefaultLogger.Debug("latest ID in version message", "version", idFromVersionMessage)
//
//	id, err := strconv.ParseUint(idFromVersionMessage, 10, 32)
//	if err != nil {
//		log.DefaultLogger.Error("parsing integer error", "error", err.Error())
//	}
//	return int(id)
//}

// CreateDashboardObjects set a Dashboard with the given raw dashboard object
func (grafanaApi GrafanaApi) CreateDashboardObjects(fileMap map[string]map[string][]byte) {
	// for each folder
	for dashboardDir, dashboardFile := range fileMap {
		// get Grafana folder ID or create if not exists
		folderID := grafanaApi.GetOrCreateFolderID(dashboardDir)
		// for each dashboard within folder
		for gitDashboardName, gitRawDashboard := range dashboardFile {
			// get git and grafana dashboard object for comparison
			gitDashboardObject := getDashboardObjectFromRawDashboard(gitRawDashboard)
			grafanaDashboardObject, _ := grafanaApi.GetDashboardObjectByUID(gitDashboardObject.UID)

			// First, 'Version' and 'Dashboard ID' need to be set equal, as they are fundamentally different because of import mechanisms
			grafanaDashboardObject.Version = gitDashboardObject.Version
			grafanaDashboardObject.ID = gitDashboardObject.ID

			if !reflect.DeepEqual(grafanaDashboardObject, gitDashboardObject) {
				grafanaApi.CreateOrUpdateDashboardObjectByID(gitRawDashboard, folderID, "Synchronized from Version " + strconv.Itoa(int(gitDashboardObject.Version)))
				log.DefaultLogger.Debug("Dashboard created", "name", gitDashboardName)
			} else {
				log.DefaultLogger.Debug("Dashboard already up-to-date", "name", gitDashboardName)
			}
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
