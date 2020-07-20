package definition

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

type Index struct {
	ContentType string     `json:"content_type"`
	ValidUntil  time.Time  `json:"valid_until"`
	Packages    []*Package `json:"packages"`
}

type Package struct {
	// to be generated and served
	ValidUntil  time.Time `json:"valid_until"`
	URL         string    `json:"url"`
	LatestURL   string    `json:"latest_url"`
	DownloadURL string    `json:"download_url"`
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

func FromDir(dir string) ([]*Package, error) {
	matches, err := filepath.Glob(path.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("globbing dir: %w", err)
	}
	var definitions []*Package
	for _, match := range matches {
		file, err := os.Open(match)
		if err != nil {
			return nil, fmt.Errorf("open file: %w", err)
		}
		var def Package
		if err := yaml.NewDecoder(file).Decode(&def); err != nil {
			return nil, fmt.Errorf("decode file: %w", err)
		}
		definitions = append(definitions, &def)
	}
	return definitions, nil
}
