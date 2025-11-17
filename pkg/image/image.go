package image

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Image represents a container image
type Image struct {
	Name    string
	Tag     string
	Layers  []Layer
	Config  ImageConfig
	RootFS  string
	BaseDir string
}

// Layer represents a single filesystem layer
type Layer struct {
	ID     string
	Digest string
	Path   string
	Size   int64
}

// ImageConfig holds image configuration
type ImageConfig struct {
	Cmd        []string          `json:"Cmd"`
	Entrypoint []string          `json:"Entrypoint"`
	Env        []string          `json:"Env"`
	WorkingDir string            `json:"WorkingDir"`
	User       string            `json:"User"`
	Labels     map[string]string `json:"Labels"`
}

// Manifest represents an image manifest (simplified OCI format)
type Manifest struct {
	SchemaVersion int            `json:"schemaVersion"`
	MediaType     string         `json:"mediaType"`
	Config        Descriptor     `json:"config"`
	Layers        []Descriptor   `json:"layers"`
	Annotations   map[string]string `json:"annotations"`
}

// Descriptor describes a layer or config
type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// Pull simulates pulling an image (simplified - no registry interaction)
func Pull(imageName, tag string) (*Image, error) {
	// In a real implementation, this would:
	// 1. Connect to a registry (Docker Hub, etc.)
	// 2. Download the manifest
	// 3. Download each layer
	// 4. Verify checksums
	// 5. Store layers locally

	return &Image{
		Name: imageName,
		Tag:  tag,
	}, fmt.Errorf("pull not yet implemented - use Import or Build instead")
}

// Import imports a tarball as an image
func Import(tarPath, name, tag string) (*Image, error) {
	baseDir := filepath.Join("/var/lib/containr/images", fmt.Sprintf("%s:%s", name, tag))
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image directory: %w", err)
	}

	// Open tarball
	file, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tarball: %w", err)
	}
	defer file.Close()

	// Check if gzipped
	var reader io.Reader = file
	if filepath.Ext(tarPath) == ".gz" || filepath.Ext(tarPath) == ".tgz" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Extract tarball
	tarReader := tar.NewReader(reader)
	rootfsPath := filepath.Join(baseDir, "rootfs")

	if err := os.MkdirAll(rootfsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create rootfs directory: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		target := filepath.Join(rootfsPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return nil, err
			}
		case tar.TypeReg:
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, err
			}

			// Create file
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return nil, err
			}

			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return nil, err
			}
			file.Close()

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return nil, err
			}
		}
	}

	img := &Image{
		Name:    name,
		Tag:     tag,
		RootFS:  rootfsPath,
		BaseDir: baseDir,
	}

	return img, nil
}

// Export exports an image to a tarball
func (img *Image) Export(outputPath string) error {
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk the rootfs and add files to tar
	return filepath.Walk(img.RootFS, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(img.RootFS, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetRootFS returns the path to the image's root filesystem
func (img *Image) GetRootFS() string {
	return img.RootFS
}

// List lists all locally stored images
func List() ([]*Image, error) {
	imagesDir := "/var/lib/containr/images"
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return []*Image{}, nil
	}

	var images []*Image
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse image name and tag from directory name
		// Format: name:tag
		name := entry.Name()

		img := &Image{
			Name:    name,
			BaseDir: filepath.Join(imagesDir, name),
			RootFS:  filepath.Join(imagesDir, name, "rootfs"),
		}

		images = append(images, img)
	}

	return images, nil
}

// Remove removes an image
func (img *Image) Remove() error {
	return os.RemoveAll(img.BaseDir)
}

// SaveManifest saves the image manifest
func (img *Image) SaveManifest() error {
	manifest := Manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
	}

	manifestPath := filepath.Join(img.BaseDir, "manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

// LoadManifest loads the image manifest
func (img *Image) LoadManifest() error {
	manifestPath := filepath.Join(img.BaseDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	return nil
}
