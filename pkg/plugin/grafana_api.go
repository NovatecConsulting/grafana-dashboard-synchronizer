package plugin

import (
	"context"
	"github.com/grafana-tools/sdk"
)

// SearchDashboardsWithTag returns all dashboards with the given tag
// TODO global grafana url and token settings
func SearchDashboardsWithTag(grafanaURL string, apiToken string, tag string) ([]sdk.FoundBoard, error) {
	c, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)

	searchParam := sdk.SearchTag(tag)
	foundDashboards, err := c.Search(context.Background(), searchParam)

	return foundDashboards, err
}

// GetRawDashboardByID return Dashboard by the given UID as raw byte object
func GetRawDashboardByID(grafanaURL string, apiToken string, uid string) ([]byte, sdk.BoardProperties,  error) {
	c, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)

	rawDashboard, props, err := c.GetRawDashboardByUID(context.Background(), uid)

	return rawDashboard, props, err
}

// GetDashboardObjectByID return Dashboard by the given UID as object
func GetDashboardObjectByID(grafanaURL string, apiToken string, uid string) (sdk.Board, sdk.BoardProperties,  error) {
	c, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)

	dashboardObject, props, err := c.GetDashboardByUID(context.Background(), uid)

	return dashboardObject, props, err
}

// UpdateDashboardObjectByID update the Dashboard with the given dashboard object
func UpdateDashboardObjectByID(grafanaURL string, apiToken string, dashboard sdk.Board) (sdk.StatusMessage, error) {
	c, _ := sdk.NewClient(grafanaURL, apiToken, sdk.DefaultHTTPClient)

	statusMessage, err := c.SetDashboard(context.Background() ,dashboard, sdk.SetDashboardParams{
			Overwrite: false,
		})
	return statusMessage, err
}

// DeleteTagFromDashboardObjectByID delete the given tag from the Dashboard object
func DeleteTagFromDashboardObjectByID(dashboard sdk.Board, tag string) sdk.Board {
	for i, iTag := range dashboard.Tags {
		if iTag == tag {
			dashboard.Tags = append(dashboard.Tags[:i], dashboard.Tags[i+1:]...)
			break
		}
	}
	return dashboard
}

