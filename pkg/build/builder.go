package build

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// Builder represents the image builder
type Builder struct {
	dockerfile  *Dockerfile
	context     *BuildContext
	cache       *BuildCache
	log         *logger.Logger
	config      *BuildConfig
	layers      []*Layer
}

// BuildConfig contains build configuration
type BuildConfig struct {
	Tags        []string          // Image tags
	BuildArgs   map[string]string // Build arguments
	Target      string            // Target stage for multi-stage builds
	NoCache     bool              // Disable build cache
	Pull        bool              // Always pull base image
	CacheFrom   []string          // Import cache from images
	Labels      map[string]string // Labels to add to image
	Platform    string            // Target platform (e.g., linux/amd64)
}

// BuildContext represents the build context
type BuildContext struct {
	root  string
	files map[string]string // Map of file paths to content hashes
}

// NewBuildContext creates a new build context
func NewBuildContext(root string) (*BuildContext, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	bc := &BuildContext{
		root:  absRoot,
		files: make(map[string]string),
	}

	// Index all files in context
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git and other common directories
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == ".containr" {
				return filepath.SkipDir
			}
			return nil
		}

		// Compute file hash
		hash, err := hashFile(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		bc.files[relPath] = hash
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to index build context: %w", err)
	}

	return bc, nil
}

// Hash returns a hash of the entire build context
func (bc *BuildContext) Hash() string {
	h := sha256.New()

	// Sort file paths for consistent hashing
	var paths []string
	for path := range bc.files {
		paths = append(paths, path)
	}

	// Simple sort
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[i] > paths[j] {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}

	for _, path := range paths {
		h.Write([]byte(path))
		h.Write([]byte(bc.files[path]))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// ReadFile reads a file from the build context
func (bc *BuildContext) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(bc.root, path)
	return os.ReadFile(fullPath)
}

// Layer represents an image layer
type Layer struct {
	ID        string            `json:"id"`
	Parent    string            `json:"parent,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Command   string            `json:"command"`
	Size      int64             `json:"size"`
	Config    map[string]string `json:"config,omitempty"`
}

// BuildCache manages build layer caching
type BuildCache struct {
	root    string
	enabled bool
	log     *logger.Logger
}

// NewBuildCache creates a new build cache
func NewBuildCache(root string, enabled bool) (*BuildCache, error) {
	if enabled {
		if err := os.MkdirAll(root, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	return &BuildCache{
		root:    root,
		enabled: enabled,
		log:     logger.New("build-cache"),
	}, nil
}

// CacheKey represents a cache lookup key
type CacheKey struct {
	ParentHash  string
	Instruction string
	ContextHash string
	BuildArgs   map[string]string
}

// Hash returns a hash of the cache key
func (ck *CacheKey) Hash() string {
	h := sha256.New()
	h.Write([]byte(ck.ParentHash))
	h.Write([]byte(ck.Instruction))
	h.Write([]byte(ck.ContextHash))

	// Sort build args for consistent hashing
	var keys []string
	for k := range ck.BuildArgs {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(ck.BuildArgs[k]))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// Lookup checks if a cached layer exists for the given key
func (bc *BuildCache) Lookup(key CacheKey) (*Layer, bool) {
	if !bc.enabled {
		return nil, false
	}

	keyHash := key.Hash()
	cachePath := filepath.Join(bc.root, keyHash+".json")

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var layer Layer
	if err := json.Unmarshal(data, &layer); err != nil {
		bc.log.Warnf("Failed to unmarshal cached layer: %v", err)
		return nil, false
	}

	bc.log.Infof("Cache hit for key %s", keyHash[:12])
	return &layer, true
}

// Store saves a layer to the cache
func (bc *BuildCache) Store(key CacheKey, layer *Layer) error {
	if !bc.enabled {
		return nil
	}

	keyHash := key.Hash()
	cachePath := filepath.Join(bc.root, keyHash+".json")

	data, err := json.MarshalIndent(layer, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal layer: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	bc.log.Infof("Stored layer in cache: %s", keyHash[:12])
	return nil
}

// Prune removes old cache entries
func (bc *BuildCache) Prune(maxAge time.Duration) error {
	if !bc.enabled {
		return nil
	}

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	entries, err := os.ReadDir(bc.root)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(bc.root, entry.Name())
			if err := os.Remove(path); err != nil {
				bc.log.Warnf("Failed to remove cache file %s: %v", path, err)
			} else {
				removed++
			}
		}
	}

	bc.log.Infof("Pruned %d cache entries older than %v", removed, maxAge)
	return nil
}

// NewBuilder creates a new builder
func NewBuilder(dockerfile *Dockerfile, context *BuildContext, config *BuildConfig) (*Builder, error) {
	cache, err := NewBuildCache(filepath.Join(context.root, ".containr", "cache"), !config.NoCache)
	if err != nil {
		return nil, fmt.Errorf("failed to create build cache: %w", err)
	}

	return &Builder{
		dockerfile: dockerfile,
		context:    context,
		cache:      cache,
		config:     config,
		log:        logger.New("builder"),
		layers:     make([]*Layer, 0),
	}, nil
}

// Build executes the build
func (b *Builder) Build(ctx context.Context) (*ImageManifest, error) {
	b.log.Info("Starting image build")

	// Determine which stage to build
	var targetStage *BuildStage
	if b.config.Target != "" {
		targetStage = b.dockerfile.GetStageByName(b.config.Target)
		if targetStage == nil {
			return nil, fmt.Errorf("target stage %s not found", b.config.Target)
		}
	} else {
		targetStage = b.dockerfile.GetFinalStage()
		if targetStage == nil {
			return nil, fmt.Errorf("no stages found in Dockerfile")
		}
	}

	// Build the target stage
	if err := b.buildStage(ctx, targetStage); err != nil {
		return nil, fmt.Errorf("failed to build stage: %w", err)
	}

	// Create image manifest
	manifest := &ImageManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: ManifestConfig{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Size:      0, // TODO: Calculate
			Digest:    "", // TODO: Calculate
		},
		Layers:  make([]ManifestLayer, len(b.layers)),
		Created: time.Now(),
		Tags:    b.config.Tags,
	}

	for i, layer := range b.layers {
		manifest.Layers[i] = ManifestLayer{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Size:      layer.Size,
			Digest:    layer.ID,
		}
	}

	b.log.Infof("Build completed successfully: %d layers", len(b.layers))
	return manifest, nil
}

// buildStage builds a single stage
func (b *Builder) buildStage(ctx context.Context, stage *BuildStage) error {
	b.log.Infof("Building stage: %s (base: %s)", stage.Name, stage.BaseImage)

	var parentHash string

	// Execute instructions
	for _, instr := range stage.Instructions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check cache
		cacheKey := CacheKey{
			ParentHash:  parentHash,
			Instruction: instr.String(),
			ContextHash: b.context.Hash(),
			BuildArgs:   b.config.BuildArgs,
		}

		if cachedLayer, found := b.cache.Lookup(cacheKey); found {
			b.log.Infof("Using cached layer for: %s", instr.Command)
			b.layers = append(b.layers, cachedLayer)
			parentHash = cachedLayer.ID
			continue
		}

		// Execute instruction
		layer, err := b.executeInstruction(ctx, instr, parentHash)
		if err != nil {
			return fmt.Errorf("failed to execute %s: %w", instr.Command, err)
		}

		b.layers = append(b.layers, layer)
		parentHash = layer.ID

		// Store in cache
		b.cache.Store(cacheKey, layer)
	}

	return nil
}

// executeInstruction executes a single instruction
func (b *Builder) executeInstruction(ctx context.Context, instr *Instruction, parentHash string) (*Layer, error) {
	b.log.Infof("Executing: %s", instr.String())

	layer := &Layer{
		ID:        generateLayerID(parentHash, instr.String()),
		Parent:    parentHash,
		CreatedAt: time.Now(),
		Command:   instr.String(),
		Config:    make(map[string]string),
	}

	// Execute based on instruction type
	switch instr.Command {
	case "FROM":
		// Base image is handled separately
		b.log.Infof("Base image: %s", strings.Join(instr.Args, " "))

	case "RUN":
		// Execute command in container
		cmd := strings.Join(instr.Args, " ")
		b.log.Infof("Running: %s", cmd)
		// TODO: Actually execute the command in a container

	case "COPY", "ADD":
		// Copy files from context
		if len(instr.Args) >= 2 {
			src := instr.Args[:len(instr.Args)-1]
			dst := instr.Args[len(instr.Args)-1]
			b.log.Infof("Copying %v to %s", src, dst)
			// TODO: Actually copy files
		}

	case "ENV":
		// Set environment variable
		if len(instr.Args) >= 2 {
			key := instr.Args[0]
			value := strings.Join(instr.Args[1:], " ")
			layer.Config["env_"+key] = value
			b.log.Infof("Setting ENV %s=%s", key, value)
		}

	case "WORKDIR":
		// Set working directory
		if len(instr.Args) > 0 {
			layer.Config["workdir"] = instr.Args[0]
			b.log.Infof("Setting WORKDIR %s", instr.Args[0])
		}

	case "EXPOSE":
		// Expose port
		if len(instr.Args) > 0 {
			layer.Config["expose"] = strings.Join(instr.Args, ",")
			b.log.Infof("Exposing ports: %s", instr.Args[0])
		}

	case "CMD":
		// Set default command
		layer.Config["cmd"] = strings.Join(instr.Args, " ")
		b.log.Infof("Setting CMD: %s", strings.Join(instr.Args, " "))

	case "ENTRYPOINT":
		// Set entrypoint
		layer.Config["entrypoint"] = strings.Join(instr.Args, " ")
		b.log.Infof("Setting ENTRYPOINT: %s", strings.Join(instr.Args, " "))

	default:
		b.log.Warnf("Instruction %s not fully implemented", instr.Command)
	}

	return layer, nil
}

// ImageManifest represents an OCI image manifest
type ImageManifest struct {
	SchemaVersion int              `json:"schemaVersion"`
	MediaType     string           `json:"mediaType"`
	Config        ManifestConfig   `json:"config"`
	Layers        []ManifestLayer  `json:"layers"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Created       time.Time        `json:"created"`
	Tags          []string         `json:"tags"`
}

// ManifestConfig represents the config section of a manifest
type ManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// ManifestLayer represents a layer in a manifest
type ManifestLayer struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// generateLayerID generates a unique layer ID
func generateLayerID(parent, instruction string) string {
	h := sha256.New()
	h.Write([]byte(parent))
	h.Write([]byte(instruction))
	h.Write([]byte(time.Now().String()))
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

// hashFile computes the SHA256 hash of a file
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}
