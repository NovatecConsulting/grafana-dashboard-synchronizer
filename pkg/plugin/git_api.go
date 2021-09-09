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
	"os"
	"time"
)

// Workflow to fetch remote branch, add Json Content as File to memory, commit and push to remote repo
func callGit(gitUrl string, privateKeyFilePath string, fileName string, fileContent string)  {
	// git authentication with ssh
	_, err := os.Stat(privateKeyFilePath)
	if err != nil {
		log.DefaultLogger.Warn("read file %s failed", privateKeyFilePath, err.Error())
	}
	authenticator, err:= ssh.NewPublicKeysFromFile("git", privateKeyFilePath, "")
	if err != nil {
		log.DefaultLogger.Warn("generate public keys failed", "error", err.Error())
	}
	// TODO delete and set known hosts?
	authenticator.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

	// TODO develop small functions:
	// prepare in memory
	var store *memory.Storage
	store = memory.NewStorage()
	var fs billy.Filesystem
	fs = memfs.New()

	// git clone
	r, err := git.Clone(store, fs, &git.CloneOptions{
		URL: gitUrl,
		Auth: authenticator,
	})
	if err != nil {
		log.DefaultLogger.Error("clone error" , "error", err)
	} else {
		log.DefaultLogger.Info("repo cloned")
	}

	// fetch repo
	log.DefaultLogger.Info("fetching repo")
	err = r.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Auth: authenticator,
	})
	if err != nil {
		log.DefaultLogger.Error("fetch error", "fetchMessage", err)
	}

	// add file with content
	tempFile, err := fs.Create(fileName)
	if err != nil {
		log.DefaultLogger.Error("create file error" , "error", err)
	} else {
		tempFile.Write([]byte(fileContent))
		tempFile.Close()
	}

	// get worktree and commit
	w, err := r.Worktree()
	if err != nil {
		log.DefaultLogger.Error("worktree error" , "error", err)
	} else {
		w.Add("./")
		wStatus, _ := w.Status()
		log.DefaultLogger.Info("worktree status" , "status", wStatus)
		// TODO get tag from frontend
		_, err := w.Commit("Dashboards synced with tag " + "<tag>", &git.CommitOptions{
			Author: (*object2.Signature)(&object.Signature{
				Name:  "grafana_backend",
				Email: "",
				When:  time.Now(),
			}),
		})
		if err != nil {
			log.DefaultLogger.Error("worktree commit error" , "error", err.Error())
		}
	}

	// push repo
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: authenticator,
	})
	if err != nil {
		log.DefaultLogger.Error("push error" , "error", err.Error())
	}
}