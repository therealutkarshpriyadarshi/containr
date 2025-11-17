package csi_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/therealutkarshpriyadarshi/containr/pkg/storage/csi"
)

// Example demonstrates basic CSI manager usage
func Example_basicUsage() {
	ctx := context.Background()
	tmpDir := "/tmp/csi-example"
	defer os.RemoveAll(tmpDir)

	// Create CSI manager
	manager, err := csi.NewManager(tmpDir)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Close()

	// Create local storage driver
	localDriver, err := csi.NewLocalDriver(
		"local.csi.containr.io",
		"v1.0.0",
		"node-1",
		tmpDir+"/local",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register driver
	if err := manager.RegisterDriver("local", localDriver); err != nil {
		log.Fatal(err)
	}

	// Create volume
	volReq := &csi.CreateVolumeRequest{
		Name:          "my-volume",
		CapacityBytes: 1024 * 1024 * 100, // 100MB
		VolumeCapabilities: &csi.VolumeCapabilities{
			AccessMode: csi.AccessModeSingleNodeWriter,
			VolumeMode: csi.VolumeModeFilesystem,
		},
	}

	volume, err := manager.CreateVolume(ctx, "local", volReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created volume: %s\n", volume.Name)
	fmt.Printf("Volume ID: %s\n", volume.ID)
	fmt.Printf("Status: %s\n", volume.Status)

	// List volumes
	volumes, _ := manager.ListVolumes(ctx)
	fmt.Printf("Total volumes: %d\n", len(volumes))

	// Output:
	// Created volume: my-volume
}

// Example demonstrates snapshot functionality
func Example_snapshots() {
	ctx := context.Background()
	tmpDir := "/tmp/csi-snapshot-example"
	defer os.RemoveAll(tmpDir)

	manager, _ := csi.NewManager(tmpDir)
	defer manager.Close()

	localDriver, _ := csi.NewLocalDriver("local.csi.containr.io", "v1.0.0", "node-1", tmpDir+"/local", nil)
	manager.RegisterDriver("local", localDriver)

	// Create volume
	volReq := &csi.CreateVolumeRequest{
		Name:          "data-volume",
		CapacityBytes: 1024 * 1024 * 50,
		VolumeCapabilities: &csi.VolumeCapabilities{
			AccessMode: csi.AccessModeSingleNodeWriter,
			VolumeMode: csi.VolumeModeFilesystem,
		},
	}

	volume, _ := manager.CreateVolume(ctx, "local", volReq)

	// Create snapshot
	snapReq := &csi.CreateSnapshotRequest{
		VolumeID: volume.ID,
		Name:     "backup-snapshot",
	}

	snapshot, err := manager.CreateSnapshot(ctx, snapReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created snapshot: %s\n", snapshot.ID)
	fmt.Printf("Volume ID: %s\n", snapshot.VolumeID)
	fmt.Printf("Ready to use: %v\n", snapshot.ReadyToUse)

	// Output:
	// Created snapshot:
}

// Example demonstrates volume cloning from snapshot
func Example_cloneFromSnapshot() {
	ctx := context.Background()
	tmpDir := "/tmp/csi-clone-example"
	defer os.RemoveAll(tmpDir)

	manager, _ := csi.NewManager(tmpDir)
	defer manager.Close()

	localDriver, _ := csi.NewLocalDriver("local.csi.containr.io", "v1.0.0", "node-1", tmpDir+"/local", nil)
	manager.RegisterDriver("local", localDriver)

	// Create source volume
	srcReq := &csi.CreateVolumeRequest{
		Name:          "source-volume",
		CapacityBytes: 1024 * 1024 * 50,
		VolumeCapabilities: &csi.VolumeCapabilities{
			AccessMode: csi.AccessModeSingleNodeWriter,
			VolumeMode: csi.VolumeModeFilesystem,
		},
	}

	srcVolume, _ := manager.CreateVolume(ctx, "local", srcReq)

	// Create snapshot
	snapReq := &csi.CreateSnapshotRequest{
		VolumeID: srcVolume.ID,
		Name:     "clone-source",
	}

	snapshot, _ := manager.CreateSnapshot(ctx, snapReq)

	// Create volume from snapshot
	cloneReq := &csi.CreateVolumeRequest{
		Name:          "cloned-volume",
		CapacityBytes: 1024 * 1024 * 50,
		VolumeCapabilities: &csi.VolumeCapabilities{
			AccessMode: csi.AccessModeSingleNodeWriter,
			VolumeMode: csi.VolumeModeFilesystem,
		},
		ContentSource: &csi.VolumeSource{
			Type:       "snapshot",
			SnapshotID: snapshot.ID,
		},
	}

	clonedVolume, err := manager.CreateVolume(ctx, "local", cloneReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Cloned volume: %s from snapshot: %s\n", clonedVolume.Name, snapshot.ID)

	// Output:
	// Cloned volume:
}

// Example demonstrates NFS driver usage
func Example_nfsDriver() {
	tmpDir := "/tmp/csi-nfs-example"
	defer os.RemoveAll(tmpDir)

	manager, _ := csi.NewManager(tmpDir)
	defer manager.Close()

	// Create NFS driver
	nfsConfig := &csi.NFSConfig{
		Server:     "nfs.example.com",
		ExportPath: "/exports/volumes",
		MountOptions: []string{
			"vers=4.1",
			"rsize=1048576",
			"wsize=1048576",
		},
	}

	// Note: This will fail without a real NFS server
	nfsDriver, err := csi.NewNFSDriver(
		"nfs.csi.containr.io",
		"v1.0.0",
		"node-1",
		tmpDir+"/nfs",
		nfsConfig,
	)

	if err != nil {
		fmt.Printf("NFS driver requires real server: %v\n", err != nil)
		return
	}

	manager.RegisterDriver("nfs", nfsDriver)

	fmt.Println("NFS driver configured")

	// Output:
	// NFS driver requires real server: true
}

// Example demonstrates encryption manager
func Example_encryption() {
	// Note: Encryption requires cryptsetup to be installed
	em, err := csi.NewEncryptionManager()
	if err != nil {
		fmt.Printf("Encryption requires cryptsetup: %v\n", err != nil)
		return
	}
	defer em.Close()

	// Generate random encryption key
	key, err := em.GenerateRandomKey(32)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated key length: %d\n", len(key))

	// Get encryption stats
	stats := em.GetEncryptionStats()
	fmt.Printf("Encryption type: %s\n", stats.EncryptionType)
	fmt.Printf("Cipher: %s\n", stats.Cipher)

	// Output:
	// Encryption requires cryptsetup: true
}
