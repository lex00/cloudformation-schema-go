package spec

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DefaultSpecURL is the URL for the CloudFormation Resource Specification.
const DefaultSpecURL = "https://d1uauaxba7bl26.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json"

// FetchOptions configures how the spec is fetched.
type FetchOptions struct {
	// URL to fetch the spec from. Defaults to DefaultSpecURL.
	URL string
	// Force re-download even if cached.
	Force bool
	// CacheDir is the directory to cache the spec. Defaults to system temp dir.
	CacheDir string
	// Quiet suppresses progress output.
	Quiet bool
}

// FetchSpec downloads and parses the CloudFormation spec.
// If opts is nil, default options are used.
func FetchSpec(opts *FetchOptions) (*Spec, error) {
	if opts == nil {
		opts = &FetchOptions{}
	}
	if opts.URL == "" {
		opts.URL = DefaultSpecURL
	}
	if opts.CacheDir == "" {
		opts.CacheDir = filepath.Join(os.TempDir(), "cloudformation-schema-go")
	}

	cachePath := filepath.Join(opts.CacheDir, "spec.json")

	// Check for cached spec
	if !opts.Force {
		if data, err := os.ReadFile(cachePath); err == nil {
			var spec Spec
			if err := json.Unmarshal(data, &spec); err == nil {
				if !opts.Quiet {
					fmt.Println("Using cached spec...")
				}
				return &spec, nil
			}
		}
	}

	// Download the spec
	if !opts.Quiet {
		fmt.Printf("Downloading from %s...\n", opts.URL)
	}
	resp, err := http.Get(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("downloading spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Decompress gzip
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" || filepath.Ext(opts.URL) == ".json" {
		// The URL says gzip, try to decompress
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			// Not actually gzipped, use raw body
			resp.Body.Close()
			resp, err = http.Get(opts.URL)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			reader = resp.Body
		} else {
			defer gzReader.Close()
			reader = gzReader
		}
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading spec: %w", err)
	}

	// Parse JSON
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec: %w", err)
	}

	// Cache the spec
	if err := os.MkdirAll(opts.CacheDir, 0755); err == nil {
		_ = os.WriteFile(cachePath, data, 0644)
	}

	return &spec, nil
}

// LoadSpec loads a spec from a JSON file.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec file: %w", err)
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec: %w", err)
	}

	return &spec, nil
}
