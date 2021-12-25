package controller

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"go.senan.xyz/standardnotes-extensions/pkg/extensions"
)

var (
	validUntil       = time.Date(2030, 0, 0, 0, 0, 0, 0, time.Local)
	snExtRepo        = "https://github.com/sn-extensions"
	regexpPackgeHash = regexp.MustCompile(`\/[0-9a-f]+\/`)

	//go:embed web/*
	webFS embed.FS
)

type Controller struct {
	BaseURL  string
	ReposDir string
}

func (c *Controller) UpdateExtension(ext *extensions.Extension) error {
	repoPath := path.Join(c.ReposDir, ext.ID)
	repo, err := RepoUpdate(repoPath, ext.RepoURL)
	if err != nil {
		return fmt.Errorf("update repo: %w", err)
	}

	version, err := RepoGetHEAD(repo)
	if err != nil {
		return fmt.Errorf("get repo version: %w", err)
	}
	lastStamp, err := RepoGetLatestStamp(repo)
	if err != nil {
		return fmt.Errorf("get repo last stamp: %w", err)
	}

	ext.Version = version
	ext.LastStamp = lastStamp
	ext.ValidUntil = validUntil

	extURL, _ := url.Parse(c.BaseURL)
	extURL.Path = path.Join(extURL.Path, ext.ID, version, ext.Index)
	ext.URL = extURL.String()
	ext.DownloadURL = snExtRepo
	lastestURL, _ := url.Parse(c.BaseURL)
	lastestURL.Path = path.Join(lastestURL.Path, ext.ID, "index.json")
	ext.LatestURL = lastestURL.String()
	return nil
}

func (c *Controller) UpdateExtensions() error {
	log.Printf("loaded %d definitions", len(extensions.Extensions))
	for _, ext := range extensions.Extensions {
		if err := c.UpdateExtension(ext); err != nil {
			return fmt.Errorf("updating extension %q: %w", ext.ID, err)
		}
	}
	return nil
}

func (c *Controller) ServeIndex() (http.Handler, error) {
	tmpl, err := template.
		New("index.tmpl").
		Funcs(template.FuncMap{
			"date": func(t time.Time) string {
				return t.Format("02 Jan 2006 15:04")
			},
		}).
		ParseFS(webFS, "**/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var exts []*extensions.Extension
		for _, v := range extensions.Extensions {
			exts = append(exts, v)
		}
		sort.Slice(exts, func(i, j int) bool {
			return exts[i].LastStamp.After(exts[j].LastStamp)
		})

		tmpl.Execute(w, struct {
			Extensions []*extensions.Extension
		}{
			Extensions: exts,
		})
	}), nil
}

func (c *Controller) ServeWeb() (http.Handler, error) {
	fs, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, fmt.Errorf("create sub fs: %w", err)
	}
	return http.StripPrefix("/web/", http.FileServer(http.FS(fs))), nil
}

func (c *Controller) ServeExtensionIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ext, ok := extensions.Extensions[vars["id"]]
	if !ok {
		http.Error(w, "can't find that extension", 404)
		return
	}
	data, err := json.MarshalIndent(ext, "", "    ")
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal extension: %v", err), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *Controller) ServeExtension(w http.ResponseWriter, r *http.Request) {
	filePath := path.Join(
		c.ReposDir,
		regexpPackgeHash.ReplaceAllString(r.URL.Path, "/"),
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
