package internal

import (
	"context"
	"strings"

	sdk "github.com/NovatecConsulting/grafana-api-go-sdk"
	log "github.com/sirupsen/logrus"
)

// GrafanaApi access to grafana api
type GrafanaApi struct {
	grafanaClient *sdk.Client
}

type DashboardWithCustomFields struct {
	sdk.Board
	SyncOrigin string `json:"syncOrigin"`
}

// NewGrafanaApi creates a new GrafanaApi instance
func NewGrafanaApi(grafanaURL string, apiToken string) *GrafanaApi {
	client, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)
	grafanaApi := GrafanaApi{client}
	return &grafanaApi
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
			log.Fatal("get dashboard object error", "error", err.Error())
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
		log.Fatal("set dashboard error", "error", err.Error())
	}
	return statusMessage
}

// CreateFolder create a folder in Grafana
func (grafanaApi GrafanaApi) CreateFolder(folderName string) int {
	folder := sdk.Folder{Title: folderName}
	folder, err := grafanaApi.grafanaClient.CreateFolder(context.Background(), folder)
	if err != nil && folderName != "General" {
		log.Fatal("get folders error", "error", err.Error())
	}
	return folder.ID
}

// GetOrCreateFolderID returns the ID of a given folder or create it
func (grafanaApi GrafanaApi) GetOrCreateFolderID(folderName string) int {
	folders, err := grafanaApi.grafanaClient.GetAllFolders(context.Background())
	if err != nil {
		log.Fatal("get all folders error", "error", err.Error())
	}
	for _, folder := range folders {
		if folder.Title == folderName {
			return folder.ID
		}
	}
	generatedFolderID := grafanaApi.CreateFolder(folderName)
	return generatedFolderID
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
