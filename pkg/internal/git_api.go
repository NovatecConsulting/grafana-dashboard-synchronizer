package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	object2 "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	log "github.com/sirupsen/logrus"
	ssh2 "golang.org/x/crypto/ssh"
)

// GitApi access to git api
type GitApi struct {
	gitUrl             string
	authenticator      transport.AuthMethod
	inMemoryStore      memory.Storage
	inMemoryFileSystem billy.Filesystem
	repository         *git.Repository
}

// NewGitApi creates a new NewGitApi instance
func NewGitApi(gitUrl string, privateKeyFile *string, gitUserName *string, gitPassword *string) *GitApi {
	var authenticator transport.AuthMethod
	var err error
	if privateKeyFile != nil {
		authenticator, err = createPublicKeys(*privateKeyFile)
	} else if gitUserName != nil && gitPassword != nil {
		authenticator, err = createBasicAuth(*gitUserName, *gitPassword)
	} else {
		err = errors.New("No authentication method provided, either set `private-key-file` or (`git-user-name` and `git-password`)")
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error":            err,
			"private-key-file": privateKeyFile,
		}).Fatal("Failed to load publiy key from the private key.")
	}
	inMemoryStore, inMemoryFileSystem := createInMemory()
	gitApi := GitApi{gitUrl, authenticator, *inMemoryStore, inMemoryFileSystem, nil}

	return &gitApi
}

// helper function to create the git authenticator
func createPublicKeys(privateKeyFile string) (transport.AuthMethod, error) {
	if privateKeyFile == "" {
		return nil, errors.New("Private key must not be empty.")
	}
	// git authentication with ssh
	authenticator, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		return nil, err
	}

	// TODO delete and set known hosts?
	authenticator.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

	// After this point use the AuthMethod interface
	var output transport.AuthMethod = authenticator
	return output, err
}

func createBasicAuth(userName string, password string) (transport.AuthMethod, error) {
	if userName == "" {
		return nil, errors.New("Username not be empty.")
	}
	if password == "" {
		return nil, errors.New("Password not be empty.")
	}
	authenticator := &http.BasicAuth{
		Username: userName,
		Password: password,
	}

	// After this point use the AuthMethod interface
	var output transport.AuthMethod = authenticator
	return output, nil
}

// helper function to create the in memory storage and filesystem
func createInMemory() (*memory.Storage, billy.Filesystem) {
	// prepare in memory
	store := memory.NewStorage()
	var fs billy.Filesystem
	fs = memfs.New()

	return store, fs
}

// CloneRepo clones the gitApi.gitUrls repository
func (gitApi *GitApi) CloneRepo(branchName string) (*git.Repository, error) {
	//todo: refactor this function

	if gitApi.repository != nil {
		// only checkout branch if repository has already be cloned
		log.WithFields(log.Fields{
			"repository-url": gitApi.gitUrl,
			"branch":         branchName,
		}).Debug("Checkout branch because repository already exists..")

		repo := gitApi.repository

		worktree, _ := repo.Worktree()

		err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", branchName)),
			Force:  true,
		})
		if err != nil {
			return nil, err
		}

		return gitApi.repository, nil
	} else {
		// clone repository into memory
		log.WithFields(log.Fields{
			"repository-url": gitApi.gitUrl,
			"branch":         branchName,
		}).Info("Cloning repository..")

		r, err := git.Clone(&gitApi.inMemoryStore, gitApi.inMemoryFileSystem, &git.CloneOptions{
			URL:           gitApi.gitUrl,
			Auth:          gitApi.authenticator,
			ReferenceName: plumbing.NewBranchReferenceName(branchName),
			SingleBranch:  false,
		})

		if err != nil {
			return nil, err
		}

		gitApi.repository = r
		return r, nil
	}
}

// AddFileWithContent add the given filename and content to the in memory filesystem
func (gitApi GitApi) AddFileWithContent(fileName string, fileContent string) {
	// add file with content to in memory filesystem
	tempFile, err := gitApi.inMemoryFileSystem.Create(fileName)
	if err != nil {
		log.Fatal("create file error", "error", err)
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
		log.Fatal("worktree error", "error", err)
		return
	} else {
		w.Add("./")
		wStatus, _ := w.Status()
		log.Debug("worktree status", "status", wStatus)

		_, err := w.Commit("Synchronized Dashboards with tag <"+tag+">", &git.CommitOptions{
			Author: (*object2.Signature)(&object.Signature{
				Name: "grafana-dashboard-sync-plugin",
				When: time.Now(),
			}),
		})
		if err != nil {
			log.Fatal("worktree commit error", "error", err.Error())
			return
		}
	}
}

// PushRepo pushes the given repository
func (gitApi GitApi) PushRepo(repository git.Repository) {
	// push repo
	err := repository.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       gitApi.authenticator,
	})
	if err != nil {
		log.Fatal("push error", "error", err.Error())
	}
}

func (gitApi GitApi) GetLatestCommitId(repository git.Repository) (string, error, string) {
	// retrieves the branch pointed by HEAD
	ref, err := repository.Head()
	if err != nil {
		return "", err, "Cannot resolve head of repository"
	}

	// get the commit object, pointed by ref
	commit, err := repository.CommitObject(ref.Hash())
	if err != nil {
		return "", err, "Cannot access commit by hash"
	}

	return commit.ID().String(), nil, ""
}

// GetFileContent get the given content of a file from the in memory filesystem
func (gitApi GitApi) GetFileContent() map[string]map[string][]byte {
	// read current in memory filesystem to get dirs
	filesOrDirs, err := gitApi.inMemoryFileSystem.ReadDir("./")
	if err != nil {
		log.Fatal("inMemoryFileSystem error", "error", err)
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
			log.Fatal("inMemoryFileSystem ReadDir error", "error", err)
			return nil
		}

		for _, file := range files {

			log.Debug("file", "name", file.Name())

			if file.IsDir() {
				continue
			}

			src, err := gitApi.inMemoryFileSystem.Open("./" + dir + "/" + file.Name())

			if err != nil {
				log.Fatal("inMemoryFileSystem Open error", "error", err)
				return nil
			}
			byteFile, err := ioutil.ReadAll(src)
			if err != nil {
				log.Fatal("read error", "error", err)
			} else {
				fileMap[dir][file.Name()] = byteFile
				src.Close()
			}
		}
	}
	return fileMap
}
