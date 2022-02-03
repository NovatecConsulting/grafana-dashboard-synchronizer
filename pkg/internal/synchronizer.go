package internal

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func Test() {
	log.Info("Hello world!")
}

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

func NewSynchronizer(options SynchronizeOptions) *Synchronization {
	synchronization := Synchronization{
		options: options,
	}

	synchronization.grafanaApi = NewGrafanaApi(options.GrafanaUrl, options.GrafanaToken)
	synchronization.gitApi = NewGitApi(options.GitRepositoryUrl, options.PrivateKeyFile)

	return &synchronization
}

type Synchronization struct {
	options    SynchronizeOptions
	grafanaApi *GrafanaApi
	gitApi     GitApi
}

func (s *Synchronization) Synchronize() error {
	log.Info("Backend called with following request", s.options)

	// var properties SynchronizeOptions
	// _ = json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &properties)
	// secureProperties := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData

	// grafanaToken := secureProperties["grafanaApiToken"]
	// privateKey := []byte(secureProperties["privateSshKey"])

	// dashboardTag := properties.PushConfiguration.TagPattern

	//grafanaApi := NewGrafanaApi(options.GrafanaUrl,s.options.GrafanaToken)
	//gitApi := NewGitApi(options.GitRepositoryUrl,s.options.PrivateKeyFile)

	// Push Dashboard into Git
	if s.options.PushConfiguration.Enable {
		log.Info("Push to git repo", "url", s.options.GitRepositoryUrl)

		dashboardTag := s.options.PushConfiguration.TagPattern

		resultBoards, err := s.grafanaApi.SearchDashboardsWithTag(dashboardTag)

		if err != nil {
			log.Fatal("search dashboard", "error", err.Error())
		}

		if len(resultBoards) > 0 {
			// clone repo from specific branch
			repository, err := s.gitApi.CloneRepo(s.options.PushConfiguration.GitBranch)
			if err != nil {
				return err
			}

			for _, board := range resultBoards {
				// get dashboard Object and Properties
				dashboard, boardProperties := s.grafanaApi.GetDashboardObjectByUID(board.UID)
				log.Info(dashboard)

				// delete Tag from dashboard Object
				dashboardWithDeletedTag := s.grafanaApi.DeleteTagFromDashboardObjectByID(dashboard, dashboardTag)

				// get folder name and id, required for update processes and git folder structure
				folderId := boardProperties.FolderID

				// get raw Json Dashboard, required for import and export
				dashboardJson, err := json.Marshal(DashboardWithCustomFields{dashboardWithDeletedTag, s.options.JobName})
				if err != nil {
					log.Fatal("get raw dashboard", "error", err.Error())
				}

				// update dashboard with deleted Tag in Grafana
				s.grafanaApi.CreateOrUpdateDashboardObjectByID(dashboardJson, folderId, fmt.Sprintf("Deleted '%s' tag", dashboardTag))
				log.Debug("Dashboard preparation successfully")

				// Add Dashboard to in memory file system

				s.gitApi.AddFileWithContent(boardProperties.FolderTitle+"/"+dashboard.Title+".json", string(dashboardJson))
				log.Debug("Dashboard added to in memory file system")
			}

			s.gitApi.CommitWorktree(*repository, dashboardTag)
			s.gitApi.PushRepo(*repository)
			log.Info("Dashboards pushed successfully")
		}
	}
	// Pull Dashboards from Git
	if s.options.PullConfiguration.Enable {
		errorResult := s.pullDashboards()
		if errorResult != nil {
			return errorResult
		}
	}

	return nil
}

func (s *Synchronization) pullDashboards() error {
	log.Info("Pulling and importing dashboards")

	// clone and fetch repo from specific branch
	log.Debug("Cloning repository")
	repository, err := s.gitApi.CloneRepo(s.options.PullConfiguration.GitBranch)
	if err != nil {
		log.Fatal("Cloning repository failed", "error", err)
		return err
	}

	commitId, err, _ := s.gitApi.GetLatestCommitId(*repository)
	if err != nil {
		return err
	}

	fileMap := s.gitApi.GetFileContent()
	s.grafanaApi.CreateOrUpdateDashboard(fileMap, commitId)

	log.Info("Successfully synchronized dashboards from Git repositroy")

	return nil
}
