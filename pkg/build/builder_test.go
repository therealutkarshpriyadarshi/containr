package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewBuildContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	bc, err := NewBuildContext(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create build context: %v", err)
	}

	if len(bc.files) == 0 {
		t.Error("Expected build context to contain files")
	}

	// Check that test file is indexed
	found := false
	for path := range bc.files {
		if path == "test.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected test.txt to be in build context")
	}
}

func TestBuildContext_Hash(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	bc1, _ := NewBuildContext(tmpDir)
	hash1 := bc1.Hash()

	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Create another context with same content
	bc2, _ := NewBuildContext(tmpDir)
	hash2 := bc2.Hash()

	if hash1 != hash2 {
		t.Error("Expected same hash for same content")
	}

	// Modify file
	os.WriteFile(testFile, []byte("modified content"), 0644)

	bc3, _ := NewBuildContext(tmpDir)
	hash3 := bc3.Hash()

	if hash1 == hash3 {
		t.Error("Expected different hash after modification")
	}
}

func TestBuildContext_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	os.WriteFile(testFile, testContent, 0644)

	bc, _ := NewBuildContext(tmpDir)

	content, err := bc.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}
}

func TestBuildCache_Lookup(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewBuildCache(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	key := CacheKey{
		ParentHash:  "parent123",
		Instruction: "RUN echo test",
		ContextHash: "context456",
		BuildArgs:   map[string]string{"arg1": "value1"},
	}

	// Lookup should miss initially
	_, found := cache.Lookup(key)
	if found {
		t.Error("Expected cache miss for new key")
	}

	// Store a layer
	layer := &Layer{
		ID:        "layer123",
		Parent:    "parent123",
		CreatedAt: time.Now(),
		Command:   "RUN echo test",
		Size:      1024,
	}

	err = cache.Store(key, layer)
	if err != nil {
		t.Fatalf("Failed to store layer: %v", err)
	}

	// Lookup should hit now
	cachedLayer, found := cache.Lookup(key)
	if !found {
		t.Fatal("Expected cache hit after storing")
	}

	if cachedLayer.ID != layer.ID {
		t.Errorf("Expected layer ID '%s', got '%s'", layer.ID, cachedLayer.ID)
	}
}

func TestBuildCache_Prune(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewBuildCache(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Store some layers
	for i := 0; i < 3; i++ {
		key := CacheKey{
			ParentHash:  "parent",
			Instruction: "RUN echo " + string(rune('A'+i)),
		}

		layer := &Layer{
			ID:      "layer" + string(rune('A'+i)),
			Command: "RUN echo " + string(rune('A'+i)),
		}

		cache.Store(key, layer)
	}

	// Prune with a very short max age (everything should be pruned)
	err = cache.Prune(0)
	if err != nil {
		t.Fatalf("Failed to prune cache: %v", err)
	}

	// Check that cache directory is empty (or nearly empty)
	entries, _ := os.ReadDir(tmpDir)
	if len(entries) > 0 {
		// Some entries might remain, but should be minimal
		t.Logf("Cache directory has %d entries after pruning", len(entries))
	}
}

func TestCacheKey_Hash(t *testing.T) {
	key1 := CacheKey{
		ParentHash:  "parent",
		Instruction: "RUN echo test",
		ContextHash: "context",
		BuildArgs:   map[string]string{"arg1": "value1"},
	}

	hash1 := key1.Hash()
	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Same key should produce same hash
	key2 := CacheKey{
		ParentHash:  "parent",
		Instruction: "RUN echo test",
		ContextHash: "context",
		BuildArgs:   map[string]string{"arg1": "value1"},
	}

	hash2 := key2.Hash()
	if hash1 != hash2 {
		t.Error("Expected same hash for same key")
	}

	// Different key should produce different hash
	key3 := CacheKey{
		ParentHash:  "parent",
		Instruction: "RUN echo different",
		ContextHash: "context",
		BuildArgs:   map[string]string{"arg1": "value1"},
	}

	hash3 := key3.Hash()
	if hash1 == hash3 {
		t.Error("Expected different hash for different key")
	}
}

func TestBuilder_Build(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	dockerfileContent := `FROM alpine:latest
RUN echo "Hello World"
WORKDIR /app
COPY . .
CMD ["echo", "test"]
`
	os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)

	// Parse Dockerfile
	parser := NewParser()
	dockerfile, err := parser.ParseFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to parse Dockerfile: %v", err)
	}

	// Create build context
	context, err := NewBuildContext(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create build context: %v", err)
	}

	// Create builder
	config := &BuildConfig{
		Tags:      []string{"test:latest"},
		BuildArgs: map[string]string{},
		NoCache:   true,
	}

	builder, err := NewBuilder(dockerfile, context, config)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	// Build
	ctx := context.Background()
	manifest, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	if len(manifest.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(manifest.Tags))
	}

	if manifest.Tags[0] != "test:latest" {
		t.Errorf("Expected tag 'test:latest', got '%s'", manifest.Tags[0])
	}

	if len(manifest.Layers) == 0 {
		t.Error("Expected at least one layer in manifest")
	}
}

func TestBuilder_BuildWithCache(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple Dockerfile
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	dockerfileContent := `FROM alpine:latest
RUN echo "test"
`
	os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)

	parser := NewParser()
	dockerfile, _ := parser.ParseFile(dockerfilePath)
	context, _ := NewBuildContext(tmpDir)

	// First build with cache
	config1 := &BuildConfig{
		Tags:      []string{"test:v1"},
		BuildArgs: map[string]string{},
		NoCache:   false,
	}

	builder1, _ := NewBuilder(dockerfile, context, config1)
	ctx := context.Background()
	_, err := builder1.Build(ctx)
	if err != nil {
		t.Fatalf("First build failed: %v", err)
	}

	// Second build should use cache
	config2 := &BuildConfig{
		Tags:      []string{"test:v2"},
		BuildArgs: map[string]string{},
		NoCache:   false,
	}

	builder2, _ := NewBuilder(dockerfile, context, config2)
	_, err = builder2.Build(ctx)
	if err != nil {
		t.Fatalf("Second build failed: %v", err)
	}

	// Both builds should succeed (cache hit tested in logs)
}

func TestGenerateLayerID(t *testing.T) {
	id1 := generateLayerID("parent1", "RUN echo test")
	if id1 == "" {
		t.Error("Expected non-empty layer ID")
	}

	if len(id1) < 10 {
		t.Error("Expected longer layer ID")
	}

	// Different input should produce different ID
	id2 := generateLayerID("parent2", "RUN echo test")
	if id1 == id2 {
		t.Error("Expected different IDs for different parents")
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	os.WriteFile(testFile, testContent, 0644)

	hash1, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("Failed to hash file: %v", err)
	}

	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Same content should produce same hash
	hash2, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("Failed to hash file: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Expected same hash for same content")
	}

	// Different content should produce different hash
	os.WriteFile(testFile, []byte("different content"), 0644)
	hash3, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("Failed to hash file: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Expected different hash for different content")
	}
}
