package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/nightlyone/lockfile"
	"github.com/schollz/progressbar/v2"
)

const (
	APITmpl = "https://api.bitbucket.org/2.0/repositories/%s/"
)

const (
	cloneRepository  = iota
	updateRepository = iota
)

const (
	repositoryFolderName = "repository"
	wikiFolderName       = "wiki"
)

type Options struct {
	Username     string   `short:"u" long:"username" description:"bitbucket username"`
	Password     string   `short:"p" long:"password" description:"bitbucket user's password"`
	Location     string   `short:"l" long:"location" description:"local backup location"`
	Attempts     uint     `short:"a" long:"attempts" description:"number of attempts to make before giving up" default:"1"`
	Ignore       []string `short:"i" long:"ignore" description:"repository to ignore, may be specified several times"`
	Bare         bool     `short:"b" long:"bare" description:"clone bare repository (git only)"`
	WithWiki     bool     `short:"w" long:"with-wiki" description:"also backup wiki"`
	Prune        bool     `short:"P" long:"prune" description:"prune repo on remote update (git only)"`
	HTTP         bool     `short:"h" long:"http" description:"clone/update via https instead of ssh"`
	DryRun       bool     `short:"d" long:"dry-run" description:"do nothing, just print commands"`
	Verbose      bool     `short:"v" long:"verbose" description:"be more verbose"`
	ShowProgress bool     `short:"s" long:"show-progress" description:"show progress bar"`
}

func getRepositories(opts Options) ([]Repository, error) {
	var repositories []Repository

	client := &http.Client{}
	url := fmt.Sprintf(APITmpl, opts.Username)

	authString := fmt.Sprintf("%s:%s", opts.Username, opts.Password)
	header := "Basic " + base64.StdEncoding.EncodeToString([]byte(authString))

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", header)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpecetd HTTP status code: %d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		bbResp := &BitbucketResponse{}
		err = json.Unmarshal(body, bbResp)
		if err != nil {
			return nil, err
		}

		for i := range bbResp.Repositories {
			repositories = append(repositories, bbResp.Repositories[i])
		}

		if len(bbResp.Next) > 0 {
			url = bbResp.Next
		} else {
			break
		}
	}

	// start backup with last recently updated repo
	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].UpdatedOn.UnixNano() > repositories[j].UpdatedOn.UnixNano()
	})

	return repositories, nil
}

func dirExists(path string) (bool, error) {
	src, err := os.Stat(path)
	if err == nil && src.IsDir() {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	if os.IsExist(err) && !src.IsDir() {
		return false, nil
	}

	return true, err
}

func getRepositoryURL(opts Options, repository Repository) (*string, error) {
	var https, ssh string

	for j := range repository.Links.Clone {
		switch repository.Links.Clone[j].Name {
		case "https":
			https = repository.Links.Clone[j].Href
		case "ssh":
			ssh = repository.Links.Clone[j].Href
		default:
			return nil, fmt.Errorf("unknown repository url type: '%s'", repository.Links.Clone[j].Name)
		}
	}

	if opts.HTTP {
		if repository.IsPrivate {
			if len(opts.Username) == 0 || len(opts.Password) == 0 {
				return nil, fmt.Errorf("can't backup private repository without username or password")
			}
			info, err := url.ParseRequestURI(https)
			if err != nil {
				return nil, err
			}
			https = fmt.Sprintf(
				"%s://%s:%s@%s%s",
				info.Scheme,
				info.User.Username(),
				opts.Password,
				info.Host,
				info.Path,
			)
		}
		return &https, nil
	} else {
		return &ssh, nil
	}
}

func makeCommand(opts Options, repository Repository, how int, folderName string) ([]string, error) {
	var cmd []string

	url, err := getRepositoryURL(opts, repository)
	if err != nil {
		return nil, err
	}

	switch repository.Scm {
	case "git":
		cmd = append(cmd, "git")
		if how == cloneRepository {
			cmd = append(cmd, "clone")
			if opts.Bare {
				cmd = append(cmd, "--mirror")
			}
			if folderName == repositoryFolderName {
				cmd = append(cmd, *url, folderName)
			} else if folderName == wikiFolderName {
				cmd = append(cmd, *url+"/wiki", folderName)
			} else {
				return nil, fmt.Errorf("unexpected name of a folder: '%s'", folderName)
			}
		} else if how == updateRepository {
			cmd = append(cmd, "remote", "update")
			if opts.Prune {
				cmd = append(cmd, "--prune")
			}
		} else {
			return nil, fmt.Errorf("unknown action id: %d", how)
		}
	case "hg":
		cmd = append(cmd, "hg")
		if how == cloneRepository {
			if folderName == repositoryFolderName {
				cmd = append(cmd, "clone", *url, folderName)
			} else if folderName == wikiFolderName {
				cmd = append(cmd, "clone", *url+"/wiki")
			} else {
				return nil, fmt.Errorf("unexpected name of a folder: '%s'", folderName)
			}
		} else if how == updateRepository {
			cmd = append(cmd, "pull", "-u")
		} else {
			return nil, fmt.Errorf("unknown action id: %d", how)
		}
	default:
		return nil, fmt.Errorf("unexpected repository scheme: %s", repository.Scm)
	}

	return cmd, nil
}

func backup(opts Options, repository Repository, how int, folderName string) error {
	cmd, err := makeCommand(opts, repository, how, folderName)
	if err != nil {
		return err
	}

	err = runCommand(opts, cmd)

	return err
}

func processRepositories(opts Options, repositories []Repository) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	ignore := make(map[string]bool)

	for _, repo := range opts.Ignore {
		ignore[repo] = true
	}

	var bar *progressbar.ProgressBar

	if opts.ShowProgress {
		bar = progressbar.New(len(repositories))
	}

	var chdirPath string
	var how int

	for i := range repositories {
		if bar != nil {
			bar.Add(1)
		}

		if repositories[i].Type != "repository" {
			continue
		}

		_, skip := ignore[repositories[i].Slug]
		if skip {
			continue
		}

		err = os.MkdirAll(filepath.Join(opts.Location, repositories[i].Slug), 0775)
		if err != nil {
			return err
		}

		for _, folder := range []string{repositoryFolderName, wikiFolderName} {
			if folder == wikiFolderName && (!opts.WithWiki || !repositories[i].HasWiki) {
				continue
			}

			exists, err := dirExists(filepath.Join(opts.Location, repositories[i].Slug, folder))
			if err != nil {
				return err
			}

			if !exists {
				chdirPath = filepath.Join(opts.Location, repositories[i].Slug)
				how = cloneRepository
			} else {
				chdirPath = filepath.Join(opts.Location, repositories[i].Slug, folder)
				how = updateRepository
			}

			err = os.Chdir(chdirPath)
			if err != nil {
				return err
			}

			err = backup(opts, repositories[i], how, folder)
			if err != nil {
				return err
			}

			err = os.Chdir(currentDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runCommand(opts Options, c []string) error {
	if opts.DryRun || opts.Verbose {
		currentDir, _ := os.Getwd()
		log.Printf("executing command '%s' in the '%s' directory", strings.Join(c, " "), currentDir)
		if opts.DryRun {
			return nil
		}
	}

	var err error
	var output []byte

	for i := uint(0); i < opts.Attempts; i++ {
		cmd := exec.Command(c[0], c[1:]...)
		output, err = cmd.CombinedOutput()
		if err == nil {
			return nil
		} else {
			if opts.Verbose {
				log.Printf("#%d command '%s' failed with error '%s': '%s'",
					i+1,
					strings.Join(c, " "),
					err,
					string(output),
				)
			}
			time.Sleep(1 * time.Second)
		}
	}

	return err
}

func initLock(file string) (*lockfile.Lockfile, error) {
	lock, err := lockfile.New(filepath.Join(os.TempDir(), file))
	if err != nil {
		return nil, err
	}

	err = lock.TryLock()
	if err != nil {
		return nil, err
	}

	return &lock, nil
}

func main() {
	var opts Options

	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	lock, err := initLock("bitbucket-backup.lock")
	if err != nil {
		log.Fatalf("couldn't init lock file: %s\n", err)
	}
	defer lock.Unlock()

	repos, err := getRepositories(opts)
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = processRepositories(opts, repos)
	if err != nil {
		log.Fatalf("%s", err)
	}
}
