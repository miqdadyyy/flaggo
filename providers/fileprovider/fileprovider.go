package fileprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/miqdadyyy/flaggo"
)

// Options holds configuration for the file provider.
type Options struct {
	// Path is the filesystem path to the JSON file storing flags.
	Path string
}

// FileProvider implements flaggo.Provider using a JSON file.
type FileProvider struct {
	path  string
	flags map[string]flaggo.FlagConfig
	mu    sync.RWMutex
}

// New creates a FileProvider, loading existing flags or creating an empty file.
func New(opts Options) (*FileProvider, error) {
	p := &FileProvider{
		path:  opts.Path,
		flags: make(map[string]flaggo.FlagConfig),
	}

	if dir := filepath.Dir(p.path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("flaggo/fileprovider: create directory: %w", err)
		}
	}

	if err := p.load(); err != nil {
		return nil, fmt.Errorf("flaggo/fileprovider: load: %w", err)
	}

	return p, nil
}

// Set stores a flag configuration and persists to disk.
func (p *FileProvider) Set(ctx context.Context, key string, config flaggo.FlagConfig) error {
	p.mu.Lock()
	p.flags[key] = config
	p.mu.Unlock()

	return p.save()
}

// Get returns the configuration for a flag and whether it exists.
func (p *FileProvider) Get(ctx context.Context, key string) (flaggo.FlagConfig, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cfg, ok := p.flags[key]
	return cfg, ok
}

// All returns a copy of all flags.
func (p *FileProvider) All(ctx context.Context) (map[string]flaggo.FlagConfig, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make(map[string]flaggo.FlagConfig, len(p.flags))
	for k, v := range p.flags {
		out[k] = v
	}
	return out, nil
}

func (p *FileProvider) load() error {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return p.save()
		}
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	return json.Unmarshal(data, &p.flags)
}

func (p *FileProvider) save() error {
	p.mu.RLock()
	data, err := json.MarshalIndent(p.flags, "", "  ")
	p.mu.RUnlock()
	if err != nil {
		return err
	}

	tmp := p.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p.path)
}
