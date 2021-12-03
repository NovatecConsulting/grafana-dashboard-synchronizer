package plugin

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	ssh2 "golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	object2 "gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// GitApi access to git api
type GitApi struct {
	gitUrl string
	authenticator *ssh.PublicKeys
	inMemoryStore memory.Storage
	inMemoryFileSystem billy.Filesystem
}

// NewGitApi creates a new NewGitApi instance
func NewGitApi(gitUrl string, privateKeyFilePath string) GitApi {
	authenticator, err := createAuthenticator(privateKeyFilePath)
	if err != nil {
		log.DefaultLogger.Error("authentication failed", "error", err.Error())
	}
	inMemoryStore, inMemoryFileSystem := createInMemory()
	gitApi := GitApi {gitUrl,authenticator, *inMemoryStore, inMemoryFileSystem}

	return gitApi
}

// helper function to create the git authenticator
func createAuthenticator(privateKeyFilePath string) (*ssh.PublicKeys, error) {
	// git authentication with ssh
	_, err := os.Stat(privateKeyFilePath)
	if err != nil {
		log.DefaultLogger.Error("read file %s failed", privateKeyFilePath, err.Error())
		return nil, err
	}
	authenticator, err:= ssh.NewPublicKeysFromFile("git", privateKeyFilePath, "")
	if err != nil {
		log.DefaultLogger.Error("generate public keys failed", "error", err.Error())
		return nil, err
	}

	// TODO delete and set known hosts?
	authenticator.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

	return authenticator, err
}

// helper function to create the in memory storage and filesystem
func createInMemory() (*memory.Storage, billy.Filesystem){
	// prepare in memory
	store := memory.NewStorage()
	var fs billy.Filesystem
	fs = memfs.New()

	return store, fs
}

// CloneRepo clones the gitApi.gitUrls repository
func (gitApi GitApi) CloneRepo() (*git.Repository, error){
	// git clone
	r, err := git.Clone(&gitApi.inMemoryStore, gitApi.inMemoryFileSystem, &git.CloneOptions{
		URL: gitApi.gitUrl,
		Auth: gitApi.authenticator,
	})
	if err != nil {
		log.DefaultLogger.Error("clone error" , "error", err)
		return nil, err
	} else {
		log.DefaultLogger.Info("repo cloned")
	}

	return r, err
}

// FetchRepo fetches the given repository
func (gitApi GitApi) FetchRepo(repository git.Repository) {
	// fetch repo
	log.DefaultLogger.Info("fetching repo")
	err := repository.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth: gitApi.authenticator,
	})
	if strings.Contains(err.Error(), "already up-to-date") {
		log.DefaultLogger.Info("fetching completed", "message", err.Error())
	} else {
		log.DefaultLogger.Error("fetch error", "fetchMessage", err)
	}
}

// AddFileWithContent add the given filename and content to the in memory filesystem
func (gitApi GitApi) AddFileWithContent(fileName string, fileContent string) {
	// add file with content to in memory filesystem
	tempFile, err := gitApi.inMemoryFileSystem.Create(fileName)
	if err != nil {
		log.DefaultLogger.Error("create file error", "error", err)
		return
	} else {
		tempFile.Write([]byte(fileContent))
		tempFile.Close()
	}
}

// CommitWorktree commits all changes in the filesystem
func (gitApi GitApi) CommitWorktree(repository git.Repository, tag string) {
	// get worktree and commit
	w, err := repository.Worktree()
	if err != nil {
		log.DefaultLogger.Error("worktree error" , "error", err)
		return
	} else {
		w.Add("./")
		wStatus, _ := w.Status()
		log.DefaultLogger.Debug("worktree status" , "status", wStatus)

		_, err := w.Commit("Synchronized Dashboards with tag <" + tag + ">", &git.CommitOptions{
			Author: (*object2.Signature)(&object.Signature{
				Name:  "grafana-dashboard-sync-plugin",
				When:  time.Now(),
			}),
		})
		if err != nil {
			log.DefaultLogger.Error("worktree commit error" , "error", err.Error())
			return
		}
	}
}

// PushRepo pushes the given repository
func (gitApi GitApi) PushRepo(repository git.Repository) {
	// push repo
	err := repository.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: gitApi.authenticator,
	})
	if err != nil {
		log.DefaultLogger.Error("push error", "error", err.Error())
	}
}

// PullRepo pull the given repository
func (gitApi GitApi) PullRepo(repository git.Repository) {
	// pull repo
	w, err := repository.Worktree()
	if err != nil {
		log.DefaultLogger.Error("worktree error" , "error", err)
	} else {
		log.DefaultLogger.Debug("Pulling from Repo")
		err := w.Pull(&git.PullOptions{
			RemoteName: "origin",
			Auth: gitApi.authenticator,
		})
		if err != nil {
			if strings.Contains(err.Error(), "already up-to-date") {
				log.DefaultLogger.Info("pulling completed", "message", err.Error())
			} else {
				log.DefaultLogger.Error("pull error", "error", err.Error())
			}
		}
	}
}

// GetFileContent get the given content of a file from the in memory filesystem
func (gitApi GitApi) GetFileContent() map[string]map[string][]byte {
	// read current in memory filesystem to get dirs
	filesOrDirs, err := gitApi.inMemoryFileSystem.ReadDir("./")
	if err != nil {
		log.DefaultLogger.Error("inMemoryFileSystem error" , "error", err)
		return nil
	}

	var dirMap []string

	for _, fileOrDir := range filesOrDirs {
		if fileOrDir.IsDir() {
			dirName := fileOrDir.Name()
			dirMap = append(dirMap, dirName)
		}
	}

	fileMap := make(map[string]map[string][]byte)

	for _, dir := range dirMap {
		// prepare fileMap for dir
		fileMap[dir] = make(map[string][]byte)

		// read current in memory filesystem to get files
		files, err := gitApi.inMemoryFileSystem.ReadDir("./" + dir + "/")
		if err != nil {
			log.DefaultLogger.Error("inMemoryFileSystem ReadDir error", "error", err)
			return nil
		}

		for _, file := range files {

			log.DefaultLogger.Debug("file", "name", file.Name())

			if file.IsDir() {
				continue
			}

			src, err := gitApi.inMemoryFileSystem.Open("./" + dir + "/" + file.Name())

			if err != nil {
				log.DefaultLogger.Error("inMemoryFileSystem Open error", "error", err)
				return nil
			}
			byteFile, err := ioutil.ReadAll(src)
			if err != nil {
				log.DefaultLogger.Error("read error", "error", err)
			} else {
				fileMap[dir][file.Name()] = byteFile
				src.Close()
			}
		}
	}
	return fileMap
}