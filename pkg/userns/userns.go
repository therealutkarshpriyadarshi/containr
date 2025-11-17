package userns

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

var log = logger.New("userns")

// IDMap represents a user/group ID mapping
type IDMap struct {
	ContainerID int // ID inside the container
	HostID      int // ID on the host
	Size        int // Number of IDs to map
}

// Config holds user namespace configuration
type Config struct {
	UIDMappings  []IDMap // UID mappings
	GIDMappings  []IDMap // GID mappings
	RootlessMode bool    // Enable rootless mode
}

const (
	subUIDFile = "/etc/subuid"
	subGIDFile = "/etc/subgid"
)

// DefaultConfig returns a default user namespace configuration
func DefaultConfig() (*Config, error) {
	// Check if running as root
	uid := os.Getuid()
	if uid == 0 {
		// Running as root - use simple 1:1 mapping
		return &Config{
			UIDMappings: []IDMap{
				{ContainerID: 0, HostID: 0, Size: 1},
			},
			GIDMappings: []IDMap{
				{ContainerID: 0, HostID: 0, Size: 1},
			},
			RootlessMode: false,
		}, nil
	}

	// Running as non-root - enable rootless mode
	return RootlessConfig()
}

// RootlessConfig returns a configuration for rootless containers
func RootlessConfig() (*Config, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to get current user", err)
	}

	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to parse UID", err)
	}

	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to parse GID", err)
	}

	// Parse subuid and subgid files
	uidMappings, err := parseSubIDFile(subUIDFile, currentUser.Username)
	if err != nil {
		log.WithError(err).Warn("Failed to parse subuid file, using default mapping")
		uidMappings = []IDMap{
			{ContainerID: 0, HostID: uid, Size: 1},
		}
	} else {
		// Add mapping for current user
		uidMappings = append([]IDMap{
			{ContainerID: 0, HostID: uid, Size: 1},
		}, uidMappings...)
	}

	gidMappings, err := parseSubIDFile(subGIDFile, currentUser.Username)
	if err != nil {
		log.WithError(err).Warn("Failed to parse subgid file, using default mapping")
		gidMappings = []IDMap{
			{ContainerID: 0, HostID: gid, Size: 1},
		}
	} else {
		// Add mapping for current group
		gidMappings = append([]IDMap{
			{ContainerID: 0, HostID: gid, Size: 1},
		}, gidMappings...)
	}

	return &Config{
		UIDMappings:  uidMappings,
		GIDMappings:  gidMappings,
		RootlessMode: true,
	}, nil
}

// parseSubIDFile parses /etc/subuid or /etc/subgid files
func parseSubIDFile(path, username string) ([]IDMap, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mappings []IDMap
	scanner := bufio.NewScanner(file)
	containerIDStart := 1 // Start from 1 since 0 is reserved for the user

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 3 {
			continue
		}

		if parts[0] != username {
			continue
		}

		hostID, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		size, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		mappings = append(mappings, IDMap{
			ContainerID: containerIDStart,
			HostID:      hostID,
			Size:        size,
		})

		containerIDStart += size
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("no mappings found for user %s", username)
	}

	return mappings, nil
}

// WriteIDMap writes ID mappings to a file
func WriteIDMap(pid int, mapType string, mappings []IDMap) error {
	mapFile := filepath.Join("/proc", strconv.Itoa(pid), mapType)

	log.WithFields(map[string]interface{}{
		"pid":      pid,
		"map_type": mapType,
		"mappings": len(mappings),
	}).Debug("Writing ID mappings")

	var content string
	for _, mapping := range mappings {
		content += fmt.Sprintf("%d %d %d\n", mapping.ContainerID, mapping.HostID, mapping.Size)
	}

	if err := os.WriteFile(mapFile, []byte(content), 0644); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to write ID mapping", err).
			WithField("file", mapFile).
			WithField("map_type", mapType).
			WithHint("Ensure you have permissions to write to /proc or disable user namespaces")
	}

	log.WithField("map_type", mapType).Info("ID mappings written successfully")
	return nil
}

// SetupUserNamespace sets up user namespace mappings for a process
func SetupUserNamespace(pid int, config *Config) error {
	log.WithField("pid", pid).Info("Setting up user namespace")

	// Write "deny" to setgroups to allow GID mapping as non-root
	if config.RootlessMode {
		setgroupsFile := filepath.Join("/proc", strconv.Itoa(pid), "setgroups")
		if err := os.WriteFile(setgroupsFile, []byte("deny\n"), 0644); err != nil {
			// This might fail on older kernels, log but continue
			log.WithError(err).Warn("Failed to write setgroups file")
		}
	}

	// Write UID mappings
	if len(config.UIDMappings) > 0 {
		if err := WriteIDMap(pid, "uid_map", config.UIDMappings); err != nil {
			return err
		}
	}

	// Write GID mappings
	if len(config.GIDMappings) > 0 {
		if err := WriteIDMap(pid, "gid_map", config.GIDMappings); err != nil {
			return err
		}
	}

	log.Info("User namespace setup complete")
	return nil
}

// ValidateConfig validates the user namespace configuration
func ValidateConfig(config *Config) error {
	if len(config.UIDMappings) == 0 {
		return errors.New(errors.ErrInvalidArgument, "at least one UID mapping is required")
	}

	if len(config.GIDMappings) == 0 {
		return errors.New(errors.ErrInvalidArgument, "at least one GID mapping is required")
	}

	// Validate UID mappings
	for i, mapping := range config.UIDMappings {
		if mapping.Size <= 0 {
			return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("UID mapping %d has invalid size %d", i, mapping.Size)).
				WithField("mapping_index", i).
				WithField("size", mapping.Size)
		}
		if mapping.ContainerID < 0 || mapping.HostID < 0 {
			return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("UID mapping %d has negative ID", i)).
				WithField("mapping_index", i)
		}
	}

	// Validate GID mappings
	for i, mapping := range config.GIDMappings {
		if mapping.Size <= 0 {
			return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("GID mapping %d has invalid size %d", i, mapping.Size)).
				WithField("mapping_index", i).
				WithField("size", mapping.Size)
		}
		if mapping.ContainerID < 0 || mapping.HostID < 0 {
			return errors.New(errors.ErrInvalidArgument, fmt.Sprintf("GID mapping %d has negative ID", i)).
				WithField("mapping_index", i)
		}
	}

	return nil
}

// IsRootless checks if the current process is running in rootless mode
func IsRootless() bool {
	return os.Getuid() != 0
}

// SupportsUserNamespaces checks if user namespaces are supported
func SupportsUserNamespaces() bool {
	// Check if /proc/self/ns/user exists
	_, err := os.Stat("/proc/self/ns/user")
	return err == nil
}

// GetMaxUIDMapping returns the maximum UID that can be mapped
func GetMaxUIDMapping() (int, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, err
	}

	mappings, err := parseSubIDFile(subUIDFile, currentUser.Username)
	if err != nil {
		// Return a default if subuid file doesn't exist or can't be parsed
		return 65536, nil
	}

	maxUID := 0
	for _, mapping := range mappings {
		end := mapping.ContainerID + mapping.Size
		if end > maxUID {
			maxUID = end
		}
	}

	return maxUID, nil
}

// GetMaxGIDMapping returns the maximum GID that can be mapped
func GetMaxGIDMapping() (int, error) {
	currentUser, err := user.Current()
	if err != nil {
		return 0, err
	}

	mappings, err := parseSubIDFile(subGIDFile, currentUser.Username)
	if err != nil {
		// Return a default if subgid file doesn't exist or can't be parsed
		return 65536, nil
	}

	maxGID := 0
	for _, mapping := range mappings {
		end := mapping.ContainerID + mapping.Size
		if end > maxGID {
			maxGID = end
		}
	}

	return maxGID, nil
}
