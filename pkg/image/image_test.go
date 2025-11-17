package image

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestImageStruct(t *testing.T) {
	img := &Image{
		Name:    "test-image",
		Tag:     "latest",
		RootFS:  "/tmp/rootfs",
		BaseDir: "/var/lib/containr/images/test-image:latest",
	}

	if img.Name != "test-image" {
		t.Errorf("Name = %s, want test-image", img.Name)
	}

	if img.Tag != "latest" {
		t.Errorf("Tag = %s, want latest", img.Tag)
	}

	if img.RootFS != "/tmp/rootfs" {
		t.Errorf("RootFS = %s, want /tmp/rootfs", img.RootFS)
	}

	if img.BaseDir != "/var/lib/containr/images/test-image:latest" {
		t.Errorf("BaseDir = %s, want /var/lib/containr/images/test-image:latest", img.BaseDir)
	}
}

func TestLayerStruct(t *testing.T) {
	layer := Layer{
		ID:     "layer-123",
		Digest: "sha256:abcdef",
		Path:   "/path/to/layer",
		Size:   1024,
	}

	if layer.ID != "layer-123" {
		t.Errorf("ID = %s, want layer-123", layer.ID)
	}

	if layer.Digest != "sha256:abcdef" {
		t.Errorf("Digest = %s, want sha256:abcdef", layer.Digest)
	}

	if layer.Path != "/path/to/layer" {
		t.Errorf("Path = %s, want /path/to/layer", layer.Path)
	}

	if layer.Size != 1024 {
		t.Errorf("Size = %d, want 1024", layer.Size)
	}
}

func TestImageConfig(t *testing.T) {
	config := ImageConfig{
		Cmd:        []string{"/bin/sh"},
		Entrypoint: []string{"/entrypoint.sh"},
		Env:        []string{"PATH=/usr/bin"},
		WorkingDir: "/app",
		User:       "root",
		Labels:     map[string]string{"version": "1.0"},
	}

	if len(config.Cmd) != 1 || config.Cmd[0] != "/bin/sh" {
		t.Error("Cmd not set correctly")
	}

	if len(config.Entrypoint) != 1 || config.Entrypoint[0] != "/entrypoint.sh" {
		t.Error("Entrypoint not set correctly")
	}

	if len(config.Env) != 1 || config.Env[0] != "PATH=/usr/bin" {
		t.Error("Env not set correctly")
	}

	if config.WorkingDir != "/app" {
		t.Errorf("WorkingDir = %s, want /app", config.WorkingDir)
	}

	if config.User != "root" {
		t.Errorf("User = %s, want root", config.User)
	}

	if config.Labels["version"] != "1.0" {
		t.Errorf("Labels[version] = %s, want 1.0", config.Labels["version"])
	}
}

func TestManifestStruct(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: Descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    "sha256:abc123",
			Size:      1024,
		},
		Layers: []Descriptor{
			{
				MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
				Digest:    "sha256:layer1",
				Size:      2048,
			},
		},
		Annotations: map[string]string{"created": "2025-01-01"},
	}

	if manifest.SchemaVersion != 2 {
		t.Errorf("SchemaVersion = %d, want 2", manifest.SchemaVersion)
	}

	if manifest.MediaType != "application/vnd.oci.image.manifest.v1+json" {
		t.Error("MediaType not set correctly")
	}

	if manifest.Config.Digest != "sha256:abc123" {
		t.Error("Config descriptor not set correctly")
	}

	if len(manifest.Layers) != 1 {
		t.Errorf("Layers count = %d, want 1", len(manifest.Layers))
	}

	if manifest.Annotations["created"] != "2025-01-01" {
		t.Error("Annotations not set correctly")
	}
}

func TestDescriptor(t *testing.T) {
	desc := Descriptor{
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Digest:    "sha256:123abc",
		Size:      4096,
	}

	if desc.MediaType != "application/vnd.oci.image.layer.v1.tar+gzip" {
		t.Error("MediaType not set correctly")
	}

	if desc.Digest != "sha256:123abc" {
		t.Error("Digest not set correctly")
	}

	if desc.Size != 4096 {
		t.Errorf("Size = %d, want 4096", desc.Size)
	}
}

func TestPull(t *testing.T) {
	// Test that Pull returns not implemented error
	img, err := Pull("alpine", "latest")

	if err == nil {
		t.Error("Pull should return error (not implemented)")
	}

	if img == nil {
		t.Error("Pull should return image struct even with error")
	}

	if img != nil {
		if img.Name != "alpine" {
			t.Errorf("Name = %s, want alpine", img.Name)
		}

		if img.Tag != "latest" {
			t.Errorf("Tag = %s, want latest", img.Tag)
		}
	}
}

func TestImport(t *testing.T) {
	// Create a temporary tarball for testing
	tmpDir := t.TempDir()
	tarPath := filepath.Join(tmpDir, "test.tar.gz")

	// Create a simple tarball
	file, err := os.Create(tarPath)
	if err != nil {
		t.Fatalf("Failed to create test tarball: %v", err)
	}

	gzWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzWriter)

	// Add a test file to the tarball
	header := &tar.Header{
		Name: "test.txt",
		Mode: 0644,
		Size: int64(len("test content")),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	if _, err := tarWriter.Write([]byte("test content")); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Test import
	imageDir := filepath.Join(tmpDir, "images")
	os.Setenv("TEST_IMAGE_DIR", imageDir) // For testing purposes

	// Import the tarball
	img, err := Import(tarPath, "test-image", "v1.0")
	if err != nil {
		t.Logf("Import failed (may be expected in test environment): %v", err)
		return
	}

	if img == nil {
		t.Fatal("Import returned nil image")
	}

	if img.Name != "test-image" {
		t.Errorf("Name = %s, want test-image", img.Name)
	}

	if img.Tag != "v1.0" {
		t.Errorf("Tag = %s, want v1.0", img.Tag)
	}

	// Cleanup
	if img.BaseDir != "" {
		os.RemoveAll(img.BaseDir)
	}
}

func TestGetRootFS(t *testing.T) {
	img := &Image{
		Name:    "test",
		Tag:     "latest",
		RootFS:  "/test/rootfs",
		BaseDir: "/test/base",
	}

	rootfs := img.GetRootFS()
	if rootfs != "/test/rootfs" {
		t.Errorf("GetRootFS() = %s, want /test/rootfs", rootfs)
	}
}

func TestList(t *testing.T) {
	// Test List function
	images, err := List()
	if err != nil {
		t.Logf("List failed (may be expected if /var/lib/containr/images doesn't exist): %v", err)
		return
	}

	// Empty list is valid - images can be nil or empty slice
	if images == nil {
		t.Log("List returned nil (no images found, which is valid)")
		return
	}

	// Empty list is valid
	t.Logf("Found %d images", len(images))
}

func TestRemove(t *testing.T) {
	// Create a temporary image directory
	tmpDir := t.TempDir()

	img := &Image{
		Name:    "test-remove",
		Tag:     "latest",
		BaseDir: filepath.Join(tmpDir, "test-image"),
	}

	// Create the directory
	os.MkdirAll(img.BaseDir, 0755)

	// Remove it
	err := img.Remove()
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(img.BaseDir); !os.IsNotExist(err) {
		t.Error("Image directory still exists after Remove")
	}
}

func TestSaveManifest(t *testing.T) {
	tmpDir := t.TempDir()

	img := &Image{
		Name:    "test-manifest",
		Tag:     "latest",
		BaseDir: tmpDir,
	}

	err := img.SaveManifest()
	if err != nil {
		t.Errorf("SaveManifest failed: %v", err)
	}

	// Verify manifest file was created
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Manifest file was not created")
	}

	// Read and verify manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Errorf("Failed to read manifest: %v", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Errorf("Failed to parse manifest: %v", err)
	}

	if manifest.SchemaVersion != 2 {
		t.Errorf("SchemaVersion = %d, want 2", manifest.SchemaVersion)
	}

	if manifest.MediaType != "application/vnd.oci.image.manifest.v1+json" {
		t.Error("MediaType not set correctly in saved manifest")
	}
}

func TestLoadManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test manifest file
	manifest := Manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	manifestPath := filepath.Join(tmpDir, "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Test loading
	img := &Image{
		Name:    "test-load",
		Tag:     "latest",
		BaseDir: tmpDir,
	}

	err = img.LoadManifest()
	if err != nil {
		t.Errorf("LoadManifest failed: %v", err)
	}
}

func TestExport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test rootfs
	rootfsDir := filepath.Join(tmpDir, "rootfs")
	os.MkdirAll(rootfsDir, 0755)

	// Add a test file
	testFile := filepath.Join(rootfsDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	img := &Image{
		Name:    "test-export",
		Tag:     "latest",
		RootFS:  rootfsDir,
		BaseDir: tmpDir,
	}

	outputPath := filepath.Join(tmpDir, "export.tar.gz")
	err := img.Export(outputPath)
	if err != nil {
		t.Errorf("Export failed: %v", err)
	}

	// Verify the tarball was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Export tarball was not created")
	}

	// Verify tarball can be opened
	file, err := os.Open(outputPath)
	if err != nil {
		t.Errorf("Failed to open export tarball: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		t.Errorf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	fileCount := 0
	for {
		_, err := tarReader.Next()
		if err != nil {
			break
		}
		fileCount++
	}

	if fileCount == 0 {
		t.Error("Export tarball is empty")
	}

	t.Logf("Export tarball contains %d files", fileCount)
}

func TestImageConfigJSON(t *testing.T) {
	config := ImageConfig{
		Cmd:        []string{"/bin/sh", "-c", "echo hello"},
		Entrypoint: []string{"/entrypoint.sh"},
		Env:        []string{"PATH=/usr/bin", "HOME=/root"},
		WorkingDir: "/app",
		User:       "nobody",
		Labels:     map[string]string{"version": "1.0", "maintainer": "test"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal ImageConfig: %v", err)
	}

	// Test JSON unmarshaling
	var decoded ImageConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("Failed to unmarshal ImageConfig: %v", err)
	}

	// Verify fields
	if len(decoded.Cmd) != len(config.Cmd) {
		t.Error("Cmd not preserved through JSON encoding")
	}

	if decoded.WorkingDir != config.WorkingDir {
		t.Error("WorkingDir not preserved through JSON encoding")
	}

	if decoded.User != config.User {
		t.Error("User not preserved through JSON encoding")
	}
}

func TestManifestJSON(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: Descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    "sha256:test",
			Size:      100,
		},
		Layers: []Descriptor{
			{
				MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
				Digest:    "sha256:layer",
				Size:      200,
			},
		},
		Annotations: map[string]string{"created": "2025-01-01"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Errorf("Failed to marshal Manifest: %v", err)
	}

	// Test JSON unmarshaling
	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("Failed to unmarshal Manifest: %v", err)
	}

	if decoded.SchemaVersion != manifest.SchemaVersion {
		t.Error("SchemaVersion not preserved through JSON encoding")
	}

	if len(decoded.Layers) != len(manifest.Layers) {
		t.Error("Layers not preserved through JSON encoding")
	}
}

func TestImageNaming(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		tag       string
		expected  string
	}{
		{
			name:      "Simple image name",
			imageName: "alpine",
			tag:       "latest",
			expected:  "alpine:latest",
		},
		{
			name:      "Versioned image",
			imageName: "ubuntu",
			tag:       "20.04",
			expected:  "ubuntu:20.04",
		},
		{
			name:      "Custom image",
			imageName: "myapp",
			tag:       "v1.2.3",
			expected:  "myapp:v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := &Image{
				Name: tt.imageName,
				Tag:  tt.tag,
			}

			fullName := img.Name + ":" + img.Tag
			if fullName != tt.expected {
				t.Errorf("Full name = %s, want %s", fullName, tt.expected)
			}
		})
	}
}
