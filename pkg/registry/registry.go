package registry

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var log = logger.New("registry")

// Client is a registry client
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// ImageReference represents a container image reference
type ImageReference struct {
	Registry   string // e.g., "docker.io"
	Repository string // e.g., "library/alpine"
	Tag        string // e.g., "latest"
	Digest     string // e.g., "sha256:..."
}

// ManifestV2 represents a Docker v2 manifest
type ManifestV2 struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

// PullOptions contains options for pulling an image
type PullOptions struct {
	DestDir string
	Verbose bool
}

// DefaultClient returns a new registry client with default settings
func DefaultClient() *Client {
	return &Client{
		baseURL: "https://registry-1.docker.io",
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ParseImageReference parses an image reference string
func ParseImageReference(ref string) (*ImageReference, error) {
	imageRef := &ImageReference{
		Registry: "docker.io",
		Tag:      "latest",
	}

	// Remove schema if present
	ref = strings.TrimPrefix(ref, "https://")
	ref = strings.TrimPrefix(ref, "http://")

	// Check for digest
	if strings.Contains(ref, "@") {
		parts := strings.Split(ref, "@")
		ref = parts[0]
		imageRef.Digest = parts[1]
	}

	// Check for tag
	if strings.Contains(ref, ":") {
		parts := strings.Split(ref, ":")
		if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
			imageRef.Tag = parts[len(parts)-1]
			ref = strings.Join(parts[:len(parts)-1], ":")
		}
	}

	// Check for custom registry
	if strings.Contains(ref, "/") {
		parts := strings.Split(ref, "/")
		if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") {
			imageRef.Registry = parts[0]
			imageRef.Repository = strings.Join(parts[1:], "/")
		} else {
			imageRef.Repository = ref
		}
	} else {
		imageRef.Repository = ref
	}

	// Add library prefix for Docker Hub official images
	if imageRef.Registry == "docker.io" && !strings.Contains(imageRef.Repository, "/") {
		imageRef.Repository = "library/" + imageRef.Repository
	}

	return imageRef, nil
}

// String returns the string representation of an image reference
func (i *ImageReference) String() string {
	ref := i.Repository
	if i.Registry != "" && i.Registry != "docker.io" {
		ref = i.Registry + "/" + ref
	}
	if i.Digest != "" {
		return ref + "@" + i.Digest
	}
	return ref + ":" + i.Tag
}

// Pull pulls an image from the registry
func (c *Client) Pull(imageRef *ImageReference, opts *PullOptions) error {
	log.WithField("image", imageRef.String()).Info("Pulling image")

	// Get authentication token
	if err := c.authenticate(imageRef); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to authenticate", err).
			WithField("image", imageRef.String())
	}

	// Fetch manifest
	manifest, err := c.fetchManifest(imageRef)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to fetch manifest", err).
			WithField("image", imageRef.String())
	}

	log.WithField("layers", len(manifest.Layers)).Info("Manifest fetched")

	// Create destination directory
	destDir := opts.DestDir
	if destDir == "" {
		destDir = filepath.Join("/var/lib/containr/images", imageRef.Repository)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create destination directory", err).
			WithField("path", destDir)
	}

	// Download config blob
	configPath := filepath.Join(destDir, "config.json")
	if err := c.downloadBlob(imageRef, manifest.Config.Digest, configPath); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to download config", err).
			WithField("digest", manifest.Config.Digest)
	}

	log.Info("Config downloaded")

	// Download layers
	layersDir := filepath.Join(destDir, "layers")
	if err := os.MkdirAll(layersDir, 0755); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create layers directory", err).
			WithField("path", layersDir)
	}

	for i, layer := range manifest.Layers {
		layerPath := filepath.Join(layersDir, fmt.Sprintf("layer-%d.tar.gz", i))

		if opts.Verbose {
			log.Infof("Downloading layer %d/%d: %s", i+1, len(manifest.Layers), layer.Digest)
		}

		if err := c.downloadBlob(imageRef, layer.Digest, layerPath); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to download layer", err).
				WithField("digest", layer.Digest).
				WithField("layer", i)
		}
	}

	// Save manifest
	manifestPath := filepath.Join(destDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to marshal manifest", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to write manifest", err).
			WithField("path", manifestPath)
	}

	log.WithField("image", imageRef.String()).Info("Image pulled successfully")
	return nil
}

// authenticate gets an authentication token from the registry
func (c *Client) authenticate(imageRef *ImageReference) error {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", imageRef.Repository)

	resp, err := c.httpClient.Get(authURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	var authResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	c.token = authResp.Token
	return nil
}

// fetchManifest fetches the image manifest
func (c *Client) fetchManifest(imageRef *ImageReference) (*ManifestV2, error) {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", c.baseURL, imageRef.Repository, imageRef.Tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch manifest: status %d, body: %s", resp.StatusCode, string(body))
	}

	var manifest ManifestV2
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// downloadBlob downloads a blob from the registry
func (c *Client) downloadBlob(imageRef *ImageReference, blobDigest, destPath string) error {
	url := fmt.Sprintf("%s/v2/%s/blobs/%s", c.baseURL, imageRef.Repository, blobDigest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download blob: status %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy blob data
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

// ExtractImageToRootFS extracts an image to a rootfs directory
func ExtractImageToRootFS(imageDir, rootfsDir string) error {
	log.WithFields(map[string]interface{}{
		"image_dir":  imageDir,
		"rootfs_dir": rootfsDir,
	}).Info("Extracting image to rootfs")

	// Create rootfs directory
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to create rootfs directory", err).
			WithField("path", rootfsDir)
	}

	// Read manifest to get layer order
	manifestPath := filepath.Join(imageDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to read manifest", err).
			WithField("path", manifestPath)
	}

	var manifest ManifestV2
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to unmarshal manifest", err)
	}

	// Extract layers in order
	layersDir := filepath.Join(imageDir, "layers")
	for i := range manifest.Layers {
		layerPath := filepath.Join(layersDir, fmt.Sprintf("layer-%d.tar.gz", i))

		log.WithField("layer", i).Debug("Extracting layer")

		if err := extractLayer(layerPath, rootfsDir); err != nil {
			return errors.Wrap(errors.ErrInternal, "failed to extract layer", err).
				WithField("layer", i).
				WithField("path", layerPath)
		}
	}

	log.Info("Image extracted successfully")
	return nil
}

// extractLayer extracts a single layer tar.gz file to the destination
func extractLayer(layerPath, destDir string) error {
	file, err := os.Open(layerPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Open gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Open tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		// Validate path to prevent directory traversal
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			log.Warnf("Skipping file with suspicious path: %s", header.Name)
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Create file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// Copy file contents
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()

		case tar.TypeSymlink:
			// Create parent directory
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Create symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Ignore if symlink already exists
				if !os.IsExist(err) {
					return err
				}
			}
		}
	}

	return nil
}

// VerifyDigest verifies a file's digest
func VerifyDigest(path string, expectedDigest string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Parse expected digest
	expected, err := digest.Parse(expectedDigest)
	if err != nil {
		return err
	}

	// Calculate actual digest
	digester := expected.Algorithm().Digester()
	if _, err := io.Copy(digester.Hash(), file); err != nil {
		return err
	}

	actual := digester.Digest()

	// Compare
	if actual != expected {
		return fmt.Errorf("digest mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

// ImageConfig represents the OCI image config
type ImageConfig v1.Image

// LoadImageConfig loads the image configuration
func LoadImageConfig(imageDir string) (*ImageConfig, error) {
	configPath := filepath.Join(imageDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to read config", err).
			WithField("path", configPath)
	}

	var config ImageConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to unmarshal config", err)
	}

	return &config, nil
}
