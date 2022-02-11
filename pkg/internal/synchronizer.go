package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	sdk "github.com/NovatecConsulting/grafana-api-go-sdk"
	log "github.com/sirupsen/logrus"
)

type SynchronizeOptions struct {
	JobName      string `yaml:"job-name"`
	GrafanaToken string `yaml:"grafana-token"`
	GrafanaUrl   string `yaml:"grafana-url"`

	GitRepositoryUrl string `yaml:"git-repository-url"`
	PrivateKeyFile   string `yaml:"private-key-file"`

	PushConfiguration PushConfiguration `yaml:"push-configuration"`
	PullConfiguration PullConfiguration `yaml:"pull-configuration"`
}

type PullConfiguration struct {
	Enable    bool   `yaml:"enable"`
	GitBranch string `yaml:"git-branch"`
	Filter    string `yaml:"filter"`
}

type PushConfiguration struct {
	PullConfiguration `yaml:",inline"`
	TagPattern        string `yaml:"tag-pattern"`
	PushTags          bool   `yaml:"push-tags"`
}

// Creates a new Synchronizer instance.
func NewSynchronizer(options SynchronizeOptions) *Synchronization {
	synchronization := Synchronization{
		options: options,
	}

	log.WithFields(log.Fields{
		"job":              options.JobName,
		"repository-url":   options.GitRepositoryUrl,
		"private-key-file": options.PrivateKeyFile,
		"grafana-url":      options.GrafanaUrl,
	}).Info("Initialize synchronizer job.")

	synchronization.grafanaApi = NewGrafanaApi(options.GrafanaUrl, options.GrafanaToken)
	synchronization.gitApi = NewGitApi(options.GitRepositoryUrl, options.PrivateKeyFile)

	return &synchronization
}

type Synchronization struct {
	options    SynchronizeOptions
	grafanaApi *GrafanaApi
	gitApi     *GitApi
}

// Executes the synchronization using the configuration stored in this struct.
func (s *Synchronization) Synchronize(dryRun bool) error {
	log.WithFields(log.Fields{
		"job":     s.options.JobName,
		"dry-run": strconv.FormatBool(dryRun),
	}).Info("Starting synchronization.")

	// push dashboard into Git
	if s.options.PushConfiguration.Enable {
		err := s.pushDashboards(dryRun)
		if err != nil {
			return err
		}
	}

	// Pull Dashboards from Git
	if s.options.PullConfiguration.Enable {
		err := s.pullDashboards(dryRun)
		if err != nil {
			return err
		}
	}

	log.WithFields(log.Fields{
		"job": s.options.JobName,
	}).Info("Job was successfully completed.")

	return nil
}

// Pushs dashboards from the configured Grafana into Git.
func (s *Synchronization) pushDashboards(dryRun bool) error {
	configuration := s.options.PushConfiguration

	log.WithFields(log.Fields{
		"job":           s.options.JobName,
		"target-branch": configuration.GitBranch,
		"filter":        configuration.Filter,
		"tag-pattern":   configuration.TagPattern,
		"push-tags":     configuration.PushTags,
	}).Info("Starting dashboard synchroization (export) into the Git repository.")

	// initializing the dashboard filter
	var regexFilter *regexp.Regexp
	var err error
	if configuration.Filter != "" {
		regexFilter, err = regexp.Compile(configuration.Filter)
		if err != nil {
			log.WithFields(log.Fields{
				"error":  err,
				"job":    s.options.JobName,
				"filter": configuration.Filter,
			}).Fatal("Invalid filter pattern for the push configuration. Skipping exportation of dashboard.")
			return err
		}
	}

	dashboardTag := configuration.TagPattern

	resultBoards, err := s.grafanaApi.SearchDashboardsWithTag(dashboardTag)

	if err != nil {
		log.WithField("error", err).Fatal("Failed fetching dashboards from Grafana.")
	}

	if len(resultBoards) > 0 {
		log.WithField("amount", len(resultBoards)).Info("Successfully fetched dashboards.")

		// clone repo from specific branch
		repository, err := s.gitApi.CloneRepo(configuration.GitBranch)
		if err != nil {
			log.WithField("error", err).Fatal("Error while cloning repository.")
			return err
		}

		for _, board := range resultBoards {
			// get dashboard Object and Properties
			dashboard, boardProperties := s.grafanaApi.GetDashboardObjectByUID(board.UID)

			// synchronize only dashboards matching the filter
			if regexFilter != nil {
				folderAndTitle := boardProperties.FolderTitle + "/" + dashboard.Title
				if regexFilter.FindStringIndex(folderAndTitle) == nil {
					log.WithFields(log.Fields{
						"dashboard-path": folderAndTitle,
						"filter":         configuration.Filter,
					}).Info("Skipping export because dashboard does not match the specified filter pattern.")
					continue
				}
			}

			// delete Tag from dashboard Object
			var dashboardWithDeletedTag sdk.Board
			if configuration.PushTags {
				dashboardWithDeletedTag = dashboard
			} else {
				dashboardWithDeletedTag = s.grafanaApi.DeleteTagFromDashboardObjectByID(dashboard, dashboardTag)
			}

			// get folder name and id, required for update processes and git folder structure
			folderId := boardProperties.FolderID

			// get raw Json Dashboard, required for import and export
			dashboardJson, err := json.Marshal(DashboardWithCustomFields{dashboardWithDeletedTag, s.options.JobName})
			if err != nil {
				log.WithField("error", err).Fatal("Error while parsing dashboard JSON.")
			}

			// update dashboard with deleted Tag in Grafana
			log.WithField("dashboard", dashboard.Title).Info("Removing sync tag from dashboard.")
			if !dryRun {
				s.grafanaApi.CreateOrUpdateDashboardObjectByID(dashboardJson, folderId, fmt.Sprintf("Deleted '%s' tag", dashboardTag))
			}
			log.Debug("Dashboard preparation successfully")

			// Add Dashboard to in memory file system
			log.WithField("dashboard", dashboard.Title).Info("Adding dashboard for synchronization.")
			s.gitApi.AddFileWithContent(boardProperties.FolderTitle+"/"+dashboard.Title+".json", string(dashboardJson))
		}

		log.Info("Pushing dashboards to the remote Git repository.")
		if !dryRun {
			s.gitApi.CommitWorktree(*repository, dashboardTag)
			s.gitApi.PushRepo(*repository)
		}

		log.Info("Successfully pushed dashboards to the remote Git repository.")
	} else {
		log.WithField("tag-pattern", configuration.TagPattern).Info("No dashboards found using the configured tag pattern.")
	}

	return nil
}

// Pulling dashboards from the configured Git and importing them into Grafana.
func (s *Synchronization) pullDashboards(dryRun bool) error {
	configuration := s.options.PullConfiguration

	log.WithFields(log.Fields{
		"job":           s.options.JobName,
		"target-branch": configuration.GitBranch,
		"filter":        configuration.Filter,
	}).Info("Starting dashboard synchroization (import) from the Git repository.")

	// initializing the dashboard filter
	var regexFilter *regexp.Regexp
	var err error
	if configuration.Filter != "" {
		regexFilter, err = regexp.Compile(configuration.Filter)
		if err != nil {
			log.WithFields(log.Fields{
				"error":  err,
				"job":    s.options.JobName,
				"filter": configuration.Filter,
			}).Fatal("Invalid filter pattern for the pull configuration. Skipping importation of dashboard.")
			return err
		}
	}

	// clone and fetch the configured repository
	repository, err := s.gitApi.CloneRepo(configuration.GitBranch)
	if err != nil {
		log.WithField("error", err).Fatal("Error while cloning repository.")
		return err
	}

	commitId, err, _ := s.gitApi.GetLatestCommitId(*repository)
	if err != nil {
		return err
	}

	// stats counter
	countImport := 0
	countUpToDate := 0

	// get files from Git repository
	fileMap := s.gitApi.GetFileContent()

	// for each folder
	for folderName, dashboardFiles := range fileMap {
		// get Grafana folder or create it if it doesn't exist
		folder, err := s.grafanaApi.GetFolder(folderName)
		if err != nil {
			log.WithFields(log.Fields{
				"folder": folderName,
				"error":  err,
			}).Fatal("Could not fetch Grafana folder.")
			continue
		} else if folder == nil {
			log.WithField("folder", folderName).Info("Creating Grafana folder.")
			if !dryRun {
				folder, err = s.grafanaApi.CreateFolder(folderName)
				if err != nil {
					log.WithFields(log.Fields{
						"folder": folderName,
						"error":  err,
					}).Fatal("Could not create Grafana folder.")
					continue
				}
			}
		}

		// for each dashboard within folder
		for _, dashboardJson := range dashboardFiles {
			// get dashboards from Git and Grafana for comparison
			dashboard := DashboardWithCustomFields{}
			err := json.Unmarshal(dashboardJson, &dashboard)
			if err != nil {
				log.WithFields(log.Fields{
					"dashboard": dashboard.Title,
					"error":     err,
				}).Fatal("Failed to unmarshal dashboard.")
			}

			// synchronize only dashboards matching the filter
			if regexFilter != nil {
				folderAndTitle := folderName + "/" + dashboard.Title
				if regexFilter.FindStringIndex(folderAndTitle) == nil {
					log.WithFields(log.Fields{
						"dashboard-path": folderAndTitle,
						"filter":         configuration.Filter,
					}).Info("Skipping import because dashboard does not match the specified filter pattern.")
					continue
				}
			}

			//gitDashboardExtended := getDashboardObjectFromRawDashboard(gitRawDashboard)
			grafanaDashboard, _ := s.grafanaApi.GetDashboardObjectByUID(dashboard.UID)

			// extract the custom tags from the dashboard model
			syncOrigin := dashboard.SyncOrigin

			// we need to explicitly set certain attributes for comparision
			// ---
			// 'Version' and 'Dashboard ID' need to be set equal, as they are different because of import mechanisms
			grafanaDashboard.Version = dashboard.Version
			grafanaDashboard.ID = dashboard.ID
			// 'SyncOrigin' need to be set, because custom fields are lost through the import
			grafanaDashboardExtended := DashboardWithCustomFields{grafanaDashboard, dashboard.SyncOrigin}

			// import dashboard if it differs from the current one
			if !reflect.DeepEqual(grafanaDashboardExtended, dashboard) {
				versionMessage := fmt.Sprintf("[SYNC] Synchronized dashboard. Version '%s' from origin '%s' (commit %s).", strconv.Itoa(int(grafanaDashboardExtended.Version)), syncOrigin, commitId)

				log.WithFields(log.Fields{
					"dashboard": dashboard.Title,
					"folder":    folderName,
				}).Info("Importing dashboard into Grafana.")
				if !dryRun {
					s.grafanaApi.CreateOrUpdateDashboardObjectByID(dashboardJson, folder.ID, versionMessage)
				}

				countImport++
			} else {
				log.WithField("dashboard", dashboard.Title).Info("Dashboard ignored because it is already up-to-date.")
				countUpToDate++
			}
		}
	}

	log.WithFields(log.Fields{
		"imported":   countImport,
		"up-to-date": countUpToDate,
	}).Info("Successfully synchronized dashboards from Git repositroy")

	return nil
}
