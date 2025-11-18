package csi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionManager manages LUKS volume encryption
type EncryptionManager struct {
	deviceMapper   string
	cryptsetupPath string
	openDevices    map[string]*EncryptedDevice
	mu             sync.RWMutex
}

// EncryptedDevice represents an open encrypted device
type EncryptedDevice struct {
	VolumeID     string
	DevicePath   string
	MapperName   string
	MountPath    string
	KeyHash      string
	CreatedAt    string
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Cipher     string
	KeySize    int
	HashAlgo   string
	Iterations int
}

// NewEncryptionManager creates a new encryption manager
func NewEncryptionManager() (*EncryptionManager, error) {
	// Verify cryptsetup is available
	cryptsetupPath, err := exec.LookPath("cryptsetup")
	if err != nil {
		return nil, fmt.Errorf("cryptsetup not found: %w (install with: apt-get install cryptsetup)", err)
	}

	// Verify device mapper is available
	if _, err := os.Stat("/dev/mapper"); err != nil {
		return nil, fmt.Errorf("device mapper not available: %w", err)
	}

	em := &EncryptionManager{
		deviceMapper:   "/dev/mapper",
		cryptsetupPath: cryptsetupPath,
		openDevices:    make(map[string]*EncryptedDevice),
	}

	return em, nil
}

// CreateEncryptedVolume creates a new LUKS encrypted volume
func (em *EncryptionManager) CreateEncryptedVolume(ctx context.Context, volumePath, passphrase string, sizeBytes int64) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Create backing file for the encrypted volume
	if err := em.createBackingFile(volumePath, sizeBytes); err != nil {
		return fmt.Errorf("failed to create backing file: %w", err)
	}

	// Derive encryption key from passphrase
	key := em.deriveKey(passphrase)

	// Format as LUKS volume
	if err := em.luksFormat(volumePath, key); err != nil {
		os.Remove(volumePath)
		return fmt.Errorf("failed to format LUKS volume: %w", err)
	}

	return nil
}

// OpenEncryptedVolume opens a LUKS encrypted volume
func (em *EncryptionManager) OpenEncryptedVolume(ctx context.Context, volumeID, volumePath, passphrase string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Check if already open
	if _, exists := em.openDevices[volumeID]; exists {
		return nil // Already open
	}

	// Derive encryption key
	key := em.deriveKey(passphrase)

	// Generate mapper name
	mapperName := fmt.Sprintf("crypt-%s", volumeID[:16])

	// Open LUKS device
	if err := em.luksOpen(volumePath, mapperName, key); err != nil {
		return fmt.Errorf("failed to open LUKS volume: %w", err)
	}

	// Get device path
	devicePath := filepath.Join(em.deviceMapper, mapperName)

	// Wait for device to be available
	if err := em.waitForDevice(devicePath); err != nil {
		em.luksClose(mapperName)
		return fmt.Errorf("encrypted device not available: %w", err)
	}

	// Create mount point
	mountPath := filepath.Join("/mnt", "crypt", volumeID)
	if err := os.MkdirAll(mountPath, 0755); err != nil {
		em.luksClose(mapperName)
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Format the device with ext4 if not already formatted
	if !em.isFormatted(devicePath) {
		if err := em.formatDevice(devicePath); err != nil {
			em.luksClose(mapperName)
			return fmt.Errorf("failed to format device: %w", err)
		}
	}

	// Mount the encrypted device
	if err := em.mountDevice(devicePath, mountPath); err != nil {
		em.luksClose(mapperName)
		return fmt.Errorf("failed to mount device: %w", err)
	}

	// Store device info
	em.openDevices[volumeID] = &EncryptedDevice{
		VolumeID:   volumeID,
		DevicePath: devicePath,
		MapperName: mapperName,
		MountPath:  mountPath,
		KeyHash:    em.hashKey(key),
	}

	return nil
}

// CloseEncryptedVolume closes a LUKS encrypted volume
func (em *EncryptionManager) CloseEncryptedVolume(ctx context.Context, volumeID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	device, exists := em.openDevices[volumeID]
	if !exists {
		return nil // Already closed or never opened
	}

	// Unmount the device
	if device.MountPath != "" {
		if err := syscall.Unmount(device.MountPath, 0); err != nil && !os.IsNotExist(err) {
			// Try lazy unmount
			syscall.Unmount(device.MountPath, syscall.MNT_DETACH)
		}
		os.RemoveAll(device.MountPath)
	}

	// Close LUKS device
	if err := em.luksClose(device.MapperName); err != nil {
		return fmt.Errorf("failed to close LUKS device: %w", err)
	}

	delete(em.openDevices, volumeID)

	return nil
}

// GetDevicePath returns the device path for an encrypted volume
func (em *EncryptionManager) GetDevicePath(volumeID string) string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if device, exists := em.openDevices[volumeID]; exists {
		return device.MountPath
	}

	return ""
}

// ResizeEncryptedVolume resizes a LUKS encrypted volume
func (em *EncryptionManager) ResizeEncryptedVolume(ctx context.Context, volumeID, volumePath string, newSizeBytes int64) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	device, exists := em.openDevices[volumeID]
	if !exists {
		return fmt.Errorf("volume %s is not open", volumeID)
	}

	// Resize backing file
	f, err := os.OpenFile(volumePath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open backing file: %w", err)
	}
	defer f.Close()

	if err := f.Truncate(newSizeBytes); err != nil {
		return fmt.Errorf("failed to resize backing file: %w", err)
	}

	// Resize LUKS container
	cmd := exec.Command(em.cryptsetupPath, "resize", device.MapperName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to resize LUKS container: %w, output: %s", err, string(output))
	}

	// Resize filesystem
	cmd = exec.Command("resize2fs", device.DevicePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to resize filesystem: %w, output: %s", err, string(output))
	}

	return nil
}

// ChangePassphrase changes the passphrase for an encrypted volume
func (em *EncryptionManager) ChangePassphrase(ctx context.Context, volumePath, oldPassphrase, newPassphrase string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	oldKey := em.deriveKey(oldPassphrase)
	newKey := em.deriveKey(newPassphrase)

	// Use luksChangeKey to change the passphrase
	cmd := exec.Command(em.cryptsetupPath, "luksChangeKey", volumePath, "--key-file", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start luksChangeKey: %w", err)
	}

	// Write old key then new key
	if _, err := stdin.Write([]byte(oldKey + "\n" + newKey + "\n")); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to write keys: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("luksChangeKey failed: %w", err)
	}

	return nil
}

// GetVolumeInfo returns information about an encrypted volume
func (em *EncryptionManager) GetVolumeInfo(volumePath string) (*EncryptionInfo, error) {
	cmd := exec.Command(em.cryptsetupPath, "luksDump", volumePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to dump LUKS info: %w", err)
	}

	info := &EncryptionInfo{
		Type:    "LUKS",
		Cipher:  em.extractField(string(output), "Cipher:"),
		KeySize: em.extractField(string(output), "MK bits:"),
		Hash:    em.extractField(string(output), "Hash spec:"),
	}

	return info, nil
}

// EncryptionInfo holds encryption information
type EncryptionInfo struct {
	Type    string
	Cipher  string
	KeySize string
	Hash    string
}

// IsEncrypted checks if a volume is LUKS encrypted
func (em *EncryptionManager) IsEncrypted(volumePath string) bool {
	cmd := exec.Command(em.cryptsetupPath, "isLuks", volumePath)
	err := cmd.Run()
	return err == nil
}

// createBackingFile creates a backing file for encrypted volume
func (em *EncryptionManager) createBackingFile(path string, sizeBytes int64) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create sparse file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := f.Truncate(sizeBytes); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	return nil
}

// luksFormat formats a file as LUKS encrypted volume
func (em *EncryptionManager) luksFormat(volumePath, key string) error {
	// Use cryptsetup luksFormat
	cmd := exec.Command(em.cryptsetupPath,
		"luksFormat",
		"--type", "luks2",
		"--cipher", "aes-xts-plain64",
		"--key-size", "512",
		"--hash", "sha256",
		"--iter-time", "2000",
		"--key-file", "-",
		"--batch-mode", // Don't ask for confirmation
		volumePath,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start luksFormat: %w", err)
	}

	// Write key to stdin
	if _, err := stdin.Write([]byte(key)); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to write key: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("luksFormat failed: %w", err)
	}

	return nil
}

// luksOpen opens a LUKS encrypted volume
func (em *EncryptionManager) luksOpen(volumePath, mapperName, key string) error {
	cmd := exec.Command(em.cryptsetupPath,
		"luksOpen",
		"--key-file", "-",
		volumePath,
		mapperName,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start luksOpen: %w", err)
	}

	// Write key to stdin
	if _, err := stdin.Write([]byte(key)); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to write key: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("luksOpen failed: %w", err)
	}

	return nil
}

// luksClose closes a LUKS encrypted volume
func (em *EncryptionManager) luksClose(mapperName string) error {
	cmd := exec.Command(em.cryptsetupPath, "luksClose", mapperName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("luksClose failed: %w, output: %s", err, string(output))
	}
	return nil
}

// deriveKey derives an encryption key from a passphrase
func (em *EncryptionManager) deriveKey(passphrase string) string {
	// Use PBKDF2 to derive a strong key from passphrase
	salt := []byte("containr-csi-encryption-salt-v1")
	iterations := 100000
	keyLen := 64

	key := pbkdf2.Key([]byte(passphrase), salt, iterations, keyLen, sha256.New)
	return hex.EncodeToString(key)
}

// hashKey creates a hash of the key for comparison
func (em *EncryptionManager) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// waitForDevice waits for a device to become available
func (em *EncryptionManager) waitForDevice(devicePath string) error {
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(devicePath); err == nil {
			return nil
		}
		syscall.Sync() // Sync to ensure device is ready
	}
	return fmt.Errorf("device %s did not become available", devicePath)
}

// isFormatted checks if a device has a filesystem
func (em *EncryptionManager) isFormatted(devicePath string) bool {
	cmd := exec.Command("blkid", devicePath)
	return cmd.Run() == nil
}

// formatDevice formats a device with ext4
func (em *EncryptionManager) formatDevice(devicePath string) error {
	cmd := exec.Command("mkfs.ext4", "-F", devicePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mkfs.ext4 failed: %w, output: %s", err, string(output))
	}
	return nil
}

// mountDevice mounts a device to a mount point
func (em *EncryptionManager) mountDevice(devicePath, mountPath string) error {
	if err := syscall.Mount(devicePath, mountPath, "ext4", 0, ""); err != nil {
		return fmt.Errorf("mount failed: %w", err)
	}
	return nil
}

// extractField extracts a field value from cryptsetup output
func (em *EncryptionManager) extractField(output, field string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, field) {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// GenerateRandomKey generates a random encryption key
func (em *EncryptionManager) GenerateRandomKey(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	return hex.EncodeToString(key), nil
}

// ListOpenDevices returns a list of open encrypted devices
func (em *EncryptionManager) ListOpenDevices() []*EncryptedDevice {
	em.mu.RLock()
	defer em.mu.RUnlock()

	devices := make([]*EncryptedDevice, 0, len(em.openDevices))
	for _, device := range em.openDevices {
		devices = append(devices, device)
	}

	return devices
}

// BackupHeader backs up the LUKS header
func (em *EncryptionManager) BackupHeader(volumePath, backupPath string) error {
	cmd := exec.Command(em.cryptsetupPath, "luksHeaderBackup", volumePath, "--header-backup-file", backupPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("header backup failed: %w, output: %s", err, string(output))
	}
	return nil
}

// RestoreHeader restores the LUKS header
func (em *EncryptionManager) RestoreHeader(volumePath, backupPath string) error {
	cmd := exec.Command(em.cryptsetupPath, "luksHeaderRestore", volumePath, "--header-backup-file", backupPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("header restore failed: %w, output: %s", err, string(output))
	}
	return nil
}

// GetEncryptionStats returns encryption statistics
func (em *EncryptionManager) GetEncryptionStats() *EncryptionStats {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return &EncryptionStats{
		TotalEncryptedVolumes: len(em.openDevices),
		OpenDevices:           len(em.openDevices),
		EncryptionType:        "LUKS2",
		Cipher:                "aes-xts-plain64",
		KeySize:               512,
	}
}

// EncryptionStats holds encryption statistics
type EncryptionStats struct {
	TotalEncryptedVolumes int
	OpenDevices           int
	EncryptionType        string
	Cipher                string
	KeySize               int
}

// Close closes the encryption manager
func (em *EncryptionManager) Close() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	var lastErr error

	// Close all open devices
	for volumeID := range em.openDevices {
		if err := em.CloseEncryptedVolume(context.Background(), volumeID); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
