package extensions

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

//go:embed *
var definitionsFS embed.FS
var Extensions map[string]*Extension

func init() {
	var err error
	Extensions, err = readExtensions()
	if err != nil {
		log.Fatalf("failed to read extensions: %v", err)
	}
}

type Extension struct {
	// to be generated and served
	ValidUntil  time.Time `json:"valid_until"`
	URL         string    `json:"url"`
	LatestURL   string    `json:"latest_url"`
	DownloadURL string    `json:"download_url"`
	LastStamp   time.Time `json:"last_stamp"`

	// to be read from file and served
	ID           string `json:"identifier"              yaml:"id"`
	RepoURL      string `json:"repo_url"                yaml:"repo_url"`
	Index        string `json:"index"                   yaml:"index"`
	Name         string `json:"name"                    yaml:"name"`
	ContentType  string `json:"content_type"            yaml:"content_type"`
	Area         string `json:"area"                    yaml:"area"`
	Version      string `json:"version"                 yaml:"version"`
	MarketingURL string `json:"marketing_url,omitempty" yaml:"marketing_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty" yaml:"thumbnail_url"`
	Description  string `json:"description"             yaml:"description"`
	DockIcon     *struct {
		Type             string `json:"type"                       yaml:"type"`
		BackgroundColour string `json:"background_color,omitempty" yaml:"background_color,omitempty"`
		ForegroundColour string `json:"foreground_color,omitempty" yaml:"foreground_color,omitempty"`
		BorderColour     string `json:"border_color,omitempty"     yaml:"border_color,omitempty"`
		Source           string `json:"source,omitempty"           yaml:"source,omitempty"`
	} `json:"dock_icon,omitempty" yaml:"dock_icon,omitempty"`
	Flags     []string `json:"flags,omitempty"     yaml:"flags,omitempty"`
	Layerable bool     `json:"layerable,omitempty" yaml:"layerable,omitempty"`
}

func readExtensions() (map[string]*Extension, error) {
	definitions, err := fs.Glob(definitionsFS, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("glob defintions fs: %w", err)
	}

	extensions := map[string]*Extension{}
	for _, name := range definitions {
		file, err := definitionsFS.Open(name)
		if err != nil {
			return nil, fmt.Errorf("open file: %w", err)
		}

		var ext Extension
		if err := yaml.NewDecoder(file).Decode(&ext); err != nil {
			return nil, fmt.Errorf("decode file: %w", err)
		}

		extensions[ext.ID] = &ext
	}

	return extensions, nil
}
