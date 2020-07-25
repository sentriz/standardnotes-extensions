package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"go.senan.xyz/standardnotes-extensions/definition"
)

var (
	validUntil       = time.Date(2030, 0, 0, 0, 0, 0, 0, time.Local)
	snExtRepo        = "https://github.com/sn-extensions"
	regexpPackgeHash = regexp.MustCompile(`\/[0-9a-f]+\b`)
)

type Controller struct {
	BaseURL        string
	ReposDir       string
	DefinitionsDir string
	Packages       map[string]*definition.Package
}

func (c *Controller) UpdatePackage(definition *definition.Package) error {
	repoPath := path.Join(c.ReposDir, definition.ID)
	repo, err := RepoUpdate(repoPath, definition.RepoURL)
	if err != nil {
		return fmt.Errorf("update repo: %w", err)
	}
	version, err := RepoGetHEAD(repo)
	if err != nil {
		return fmt.Errorf("get repo version: %w", err)
	}
	definition.Version = version
	definition.ValidUntil = validUntil
	pkgURL, _ := url.Parse(c.BaseURL)
	pkgURL.Path = path.Join(pkgURL.Path, definition.ID, version, definition.Index)
	definition.URL = pkgURL.String()
	definition.DownloadURL = snExtRepo
	lastestURL, _ := url.Parse(c.BaseURL)
	lastestURL.Path = path.Join(lastestURL.Path, definition.ID, "index.json")
	definition.LatestURL = lastestURL.String()
	c.Packages[definition.ID] = definition
	return nil
}

func (c *Controller) UpdatePackages() error {
	definitions, err := definition.FromDir(c.DefinitionsDir)
	if err != nil {
		return fmt.Errorf("build definitions: %w", err)
	}
	log.Printf("loaded %d definitions from disk", len(definitions))
	for _, definition := range definitions {
		if err := c.UpdatePackage(definition); err != nil {
			return fmt.Errorf("updating package %q: %w", definition.ID, err)
		}
	}
	return nil
}

func (c *Controller) ServeIndex(w http.ResponseWriter, r *http.Request) {
	var index definition.Index
	index.ContentType = "SN|Repo"
	index.ValidUntil = validUntil
	index.Packages = make([]*definition.Package, 0, len(c.Packages))
	for _, pkg := range c.Packages {
		index.Packages = append(index.Packages, pkg)
	}
	sort.Slice(index.Packages, func(i, j int) bool {
		return index.Packages[i].ID > index.Packages[j].ID
	})
	data, err := json.MarshalIndent(index, "", "    ")
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal packages: %v", err), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *Controller) ServePackageIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pkg, ok := c.Packages[vars["id"]]
	if !ok {
		http.Error(w, "can't find that package", 404)
		return
	}
	data, err := json.MarshalIndent(pkg, "", "    ")
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal package: %v", err), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *Controller) ServePackage(w http.ResponseWriter, r *http.Request) {
	filePath := path.Join(
		c.ReposDir,
		regexpPackgeHash.ReplaceAllString(r.URL.Path, ""),
	)
	// would have preferred to use http.ServeFile or http.Dir etc.
	// but they seem to 301 requests for /index.html to /, which the
	// android app doesn't seem to like.
	// so using ServeContent here instead
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't open file: %v", err), 500)
		return
	}
	defer file.Close()
	http.ServeContent(w, r, path.Join(c.ReposDir, r.URL.Path), time.Now(), file)
}
