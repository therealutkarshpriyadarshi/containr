package cri

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// RuntimeService defines the container runtime service interface
type RuntimeService interface {
	// RunPodSandbox creates and starts a pod sandbox
	RunPodSandbox(ctx context.Context, config *PodSandboxConfig) (string, error)

	// StopPodSandbox stops a running pod sandbox
	StopPodSandbox(ctx context.Context, podSandboxID string) error

	// RemovePodSandbox removes a pod sandbox
	RemovePodSandbox(ctx context.Context, podSandboxID string) error

	// PodSandboxStatus returns the status of a pod sandbox
	PodSandboxStatus(ctx context.Context, podSandboxID string) (*PodSandboxStatus, error)

	// ListPodSandbox lists pod sandboxes
	ListPodSandbox(ctx context.Context, filter *PodSandboxFilter) ([]*PodSandbox, error)

	// CreateContainer creates a new container in the sandbox
	CreateContainer(ctx context.Context, podSandboxID string, config *ContainerConfig) (string, error)

	// StartContainer starts a container
	StartContainer(ctx context.Context, containerID string) error

	// StopContainer stops a running container
	StopContainer(ctx context.Context, containerID string, timeout int64) error

	// RemoveContainer removes a container
	RemoveContainer(ctx context.Context, containerID string) error

	// ListContainers lists containers
	ListContainers(ctx context.Context, filter *ContainerFilter) ([]*Container, error)

	// ContainerStatus returns container status
	ContainerStatus(ctx context.Context, containerID string) (*ContainerStatus, error)
}

// ImageService defines the image service interface
type ImageService interface {
	// ListImages lists available images
	ListImages(ctx context.Context, filter *ImageFilter) ([]*Image, error)

	// PullImage pulls an image from a registry
	PullImage(ctx context.Context, image *ImageSpec, auth *AuthConfig) (string, error)

	// RemoveImage removes an image
	RemoveImage(ctx context.Context, image *ImageSpec) error

	// ImageStatus returns image status
	ImageStatus(ctx context.Context, image *ImageSpec) (*Image, error)

	// ImageFsInfo returns filesystem info for images
	ImageFsInfo(ctx context.Context) ([]*FilesystemUsage, error)
}

// PodSandboxConfig contains pod sandbox configuration
type PodSandboxConfig struct {
	Metadata     *PodSandboxMetadata
	Hostname     string
	LogDirectory string
	DNSConfig    *DNSConfig
	PortMappings []*PortMapping
	Labels       map[string]string
	Annotations  map[string]string
}

// PodSandboxMetadata contains pod sandbox metadata
type PodSandboxMetadata struct {
	Name      string
	UID       string
	Namespace string
	Attempt   uint32
}

// DNSConfig specifies DNS settings for the pod sandbox
type DNSConfig struct {
	Servers  []string
	Searches []string
	Options  []string
}

// PortMapping specifies a port mapping
type PortMapping struct {
	Protocol      Protocol
	ContainerPort int32
	HostPort      int32
	HostIP        string
}

// Protocol defines the network protocol
type Protocol int32

const (
	ProtocolTCP Protocol = 0
	ProtocolUDP Protocol = 1
)

// PodSandboxState is the state of the pod sandbox
type PodSandboxState int32

const (
	PodSandboxReady    PodSandboxState = 0
	PodSandboxNotReady PodSandboxState = 1
)

// PodSandbox contains minimal information about a pod sandbox
type PodSandbox struct {
	ID          string
	Metadata    *PodSandboxMetadata
	State       PodSandboxState
	CreatedAt   int64
	Labels      map[string]string
	Annotations map[string]string
}

// PodSandboxStatus contains the status of a pod sandbox
type PodSandboxStatus struct {
	ID          string
	Metadata    *PodSandboxMetadata
	State       PodSandboxState
	CreatedAt   int64
	Network     *PodSandboxNetworkStatus
	Linux       *LinuxPodSandboxStatus
	Labels      map[string]string
	Annotations map[string]string
}

// PodSandboxNetworkStatus contains network status of a pod sandbox
type PodSandboxNetworkStatus struct {
	IP                   string
	AdditionalIPs        []string
}

// LinuxPodSandboxStatus contains Linux-specific pod sandbox status
type LinuxPodSandboxStatus struct {
	Namespaces *Namespace
}

// Namespace contains paths to the namespaces
type Namespace struct {
	Options *NamespaceOption
}

// NamespaceOption provides namespace configuration
type NamespaceOption struct {
	Network    NamespaceMode
	Pid        NamespaceMode
	Ipc        NamespaceMode
	TargetID   string
}

// NamespaceMode is a namespace mode type
type NamespaceMode int32

const (
	NamespaceModeNode      NamespaceMode = 0
	NamespaceModePod       NamespaceMode = 1
	NamespaceModeContainer NamespaceMode = 2
	NamespaceModeTarget    NamespaceMode = 3
)

// PodSandboxFilter is used to filter pod sandboxes
type PodSandboxFilter struct {
	ID            string
	State         *PodSandboxState
	LabelSelector map[string]string
}

// ContainerConfig holds container configuration
type ContainerConfig struct {
	Metadata    *ContainerMetadata
	Image       *ImageSpec
	Command     []string
	Args        []string
	WorkingDir  string
	Envs        []*KeyValue
	Mounts      []*Mount
	Devices     []*Device
	Labels      map[string]string
	Annotations map[string]string
	LogPath     string
	Stdin       bool
	StdinOnce   bool
	Tty         bool
	Linux       *LinuxContainerConfig
}

// ContainerMetadata holds container metadata
type ContainerMetadata struct {
	Name    string
	Attempt uint32
}

// ImageSpec specifies an image
type ImageSpec struct {
	Image string
}

// KeyValue is a key-value pair
type KeyValue struct {
	Key   string
	Value string
}

// Mount specifies a mount for a container
type Mount struct {
	ContainerPath string
	HostPath      string
	Readonly      bool
	SelinuxRelabel bool
	Propagation    MountPropagation
}

// MountPropagation defines mount propagation mode
type MountPropagation int32

const (
	MountPropagationPrivate       MountPropagation = 0
	MountPropagationHostToContainer MountPropagation = 1
	MountPropagationBidirectional MountPropagation = 2
)

// Device specifies a device
type Device struct {
	ContainerPath string
	HostPath      string
	Permissions   string
}

// LinuxContainerConfig contains Linux-specific container configuration
type LinuxContainerConfig struct {
	Resources       *LinuxContainerResources
	SecurityContext *LinuxContainerSecurityContext
}

// LinuxContainerResources specifies Linux container resources
type LinuxContainerResources struct {
	CPUPeriod          int64
	CPUQuota           int64
	CPUShares          int64
	MemoryLimitInBytes int64
	OomScoreAdj        int64
	CPUsetCPUs         string
	CPUsetMems         string
}

// LinuxContainerSecurityContext holds security configuration
type LinuxContainerSecurityContext struct {
	Capabilities       *Capability
	Privileged         bool
	NamespaceOptions   *NamespaceOption
	SelinuxOptions     *SELinuxOption
	RunAsUser          *Int64Value
	RunAsGroup         *Int64Value
	RunAsUsername      string
	ReadonlyRootfs     bool
	SupplementalGroups []int64
	NoNewPrivs         bool
	MaskedPaths        []string
	ReadonlyPaths      []string
	Seccomp            *SecurityProfile
	Apparmor           *SecurityProfile
}

// Capability specifies Linux capabilities
type Capability struct {
	AddCapabilities  []string
	DropCapabilities []string
}

// SELinuxOption is SELinux configuration
type SELinuxOption struct {
	User  string
	Role  string
	Type  string
	Level string
}

// Int64Value wraps an int64 value
type Int64Value struct {
	Value int64
}

// SecurityProfile is a security profile
type SecurityProfile struct {
	ProfileType  SecurityProfileType
	LocalhostRef string
}

// SecurityProfileType defines profile type
type SecurityProfileType int32

const (
	SecurityProfileTypeUnconfined SecurityProfileType = 0
	SecurityProfileTypeRuntimeDefault SecurityProfileType = 1
	SecurityProfileTypeLocalhost SecurityProfileType = 2
)

// ContainerState is container state
type ContainerState int32

const (
	ContainerCreated ContainerState = 0
	ContainerRunning ContainerState = 1
	ContainerExited  ContainerState = 2
	ContainerUnknown ContainerState = 3
)

// Container contains information about a container
type Container struct {
	ID           string
	PodSandboxID string
	Metadata     *ContainerMetadata
	Image        *ImageSpec
	ImageRef     string
	State        ContainerState
	CreatedAt    int64
	Labels       map[string]string
	Annotations  map[string]string
}

// ContainerStatus contains container status
type ContainerStatus struct {
	ID          string
	Metadata    *ContainerMetadata
	State       ContainerState
	CreatedAt   int64
	StartedAt   int64
	FinishedAt  int64
	ExitCode    int32
	Image       *ImageSpec
	ImageRef    string
	Reason      string
	Message     string
	Labels      map[string]string
	Annotations map[string]string
	Mounts      []*Mount
	LogPath     string
}

// ContainerFilter filters containers
type ContainerFilter struct {
	ID            string
	State         *ContainerState
	PodSandboxID  string
	LabelSelector map[string]string
}

// Image contains basic image information
type Image struct {
	ID          string
	RepoTags    []string
	RepoDigests []string
	Size        uint64
	UID         *Int64Value
	Username    string
	Spec        *ImageSpec
}

// ImageFilter filters images
type ImageFilter struct {
	Image *ImageSpec
}

// AuthConfig contains registry authentication config
type AuthConfig struct {
	Username      string
	Password      string
	Auth          string
	ServerAddress string
	IdentityToken string
	RegistryToken string
}

// FilesystemUsage contains filesystem usage information
type FilesystemUsage struct {
	Timestamp  int64
	FsID       *FilesystemIdentifier
	UsedBytes  *UInt64Value
	InodesUsed *UInt64Value
}

// FilesystemIdentifier uniquely identifies a filesystem
type FilesystemIdentifier struct {
	Mountpoint string
}

// UInt64Value wraps a uint64 value
type UInt64Value struct {
	Value uint64
}

// CRIService implements both RuntimeService and ImageService
type CRIService struct {
	mu            sync.RWMutex
	podSandboxes  map[string]*PodSandboxStatus
	containers    map[string]*ContainerStatus
	images        map[string]*Image
	log           *logger.Logger
}

// NewCRIService creates a new CRI service
func NewCRIService() *CRIService {
	return &CRIService{
		podSandboxes: make(map[string]*PodSandboxStatus),
		containers:   make(map[string]*ContainerStatus),
		images:       make(map[string]*Image),
		log:          logger.New("cri-service"),
	}
}

// RunPodSandbox creates and starts a pod sandbox
func (c *CRIService) RunPodSandbox(ctx context.Context, config *PodSandboxConfig) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	podID := generateID("pod")
	c.log.Infof("Creating pod sandbox: %s (name: %s)", podID, config.Metadata.Name)

	status := &PodSandboxStatus{
		ID:       podID,
		Metadata: config.Metadata,
		State:    PodSandboxReady,
		CreatedAt: time.Now().UnixNano(),
		Network: &PodSandboxNetworkStatus{
			IP: "10.244.0.1", // Simulated IP
		},
		Labels:      config.Labels,
		Annotations: config.Annotations,
	}

	c.podSandboxes[podID] = status
	return podID, nil
}

// StopPodSandbox stops a pod sandbox
func (c *CRIService) StopPodSandbox(ctx context.Context, podSandboxID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	pod, exists := c.podSandboxes[podSandboxID]
	if !exists {
		return fmt.Errorf("pod sandbox %s not found", podSandboxID)
	}

	c.log.Infof("Stopping pod sandbox: %s", podSandboxID)
	pod.State = PodSandboxNotReady
	return nil
}

// RemovePodSandbox removes a pod sandbox
func (c *CRIService) RemovePodSandbox(ctx context.Context, podSandboxID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.podSandboxes[podSandboxID]; !exists {
		return fmt.Errorf("pod sandbox %s not found", podSandboxID)
	}

	c.log.Infof("Removing pod sandbox: %s", podSandboxID)
	delete(c.podSandboxes, podSandboxID)
	return nil
}

// PodSandboxStatus returns pod sandbox status
func (c *CRIService) PodSandboxStatus(ctx context.Context, podSandboxID string) (*PodSandboxStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	pod, exists := c.podSandboxes[podSandboxID]
	if !exists {
		return nil, fmt.Errorf("pod sandbox %s not found", podSandboxID)
	}

	return pod, nil
}

// ListPodSandbox lists pod sandboxes
func (c *CRIService) ListPodSandbox(ctx context.Context, filter *PodSandboxFilter) ([]*PodSandbox, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var pods []*PodSandbox
	for _, status := range c.podSandboxes {
		// Apply filters
		if filter != nil {
			if filter.ID != "" && status.ID != filter.ID {
				continue
			}
			if filter.State != nil && status.State != *filter.State {
				continue
			}
		}

		pod := &PodSandbox{
			ID:          status.ID,
			Metadata:    status.Metadata,
			State:       status.State,
			CreatedAt:   status.CreatedAt,
			Labels:      status.Labels,
			Annotations: status.Annotations,
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

// CreateContainer creates a container in a pod sandbox
func (c *CRIService) CreateContainer(ctx context.Context, podSandboxID string, config *ContainerConfig) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.podSandboxes[podSandboxID]; !exists {
		return "", fmt.Errorf("pod sandbox %s not found", podSandboxID)
	}

	containerID := generateID("ctr")
	c.log.Infof("Creating container: %s in pod %s", containerID, podSandboxID)

	status := &ContainerStatus{
		ID:           containerID,
		Metadata:     config.Metadata,
		State:        ContainerCreated,
		CreatedAt:    time.Now().UnixNano(),
		Image:        config.Image,
		ImageRef:     config.Image.Image,
		Labels:       config.Labels,
		Annotations:  config.Annotations,
		Mounts:       config.Mounts,
		LogPath:      config.LogPath,
	}

	c.containers[containerID] = status
	return containerID, nil
}

// StartContainer starts a container
func (c *CRIService) StartContainer(ctx context.Context, containerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	container, exists := c.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	c.log.Infof("Starting container: %s", containerID)
	container.State = ContainerRunning
	container.StartedAt = time.Now().UnixNano()
	return nil
}

// StopContainer stops a container
func (c *CRIService) StopContainer(ctx context.Context, containerID string, timeout int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	container, exists := c.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	c.log.Infof("Stopping container: %s (timeout: %d)", containerID, timeout)
	container.State = ContainerExited
	container.FinishedAt = time.Now().UnixNano()
	container.ExitCode = 0
	return nil
}

// RemoveContainer removes a container
func (c *CRIService) RemoveContainer(ctx context.Context, containerID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.containers[containerID]; !exists {
		return fmt.Errorf("container %s not found", containerID)
	}

	c.log.Infof("Removing container: %s", containerID)
	delete(c.containers, containerID)
	return nil
}

// ListContainers lists containers
func (c *CRIService) ListContainers(ctx context.Context, filter *ContainerFilter) ([]*Container, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var containers []*Container
	for _, status := range c.containers {
		// Apply filters
		if filter != nil {
			if filter.ID != "" && status.ID != filter.ID {
				continue
			}
			if filter.State != nil && status.State != *filter.State {
				continue
			}
			if filter.PodSandboxID != "" {
				// Note: We'd need to track pod associations for this
				// Simplified for now
			}
		}

		container := &Container{
			ID:           status.ID,
			Metadata:     status.Metadata,
			Image:        status.Image,
			ImageRef:     status.ImageRef,
			State:        status.State,
			CreatedAt:    status.CreatedAt,
			Labels:       status.Labels,
			Annotations:  status.Annotations,
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// ContainerStatus returns container status
func (c *CRIService) ContainerStatus(ctx context.Context, containerID string) (*ContainerStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	container, exists := c.containers[containerID]
	if !exists {
		return nil, fmt.Errorf("container %s not found", containerID)
	}

	return container, nil
}

// ListImages lists images
func (c *CRIService) ListImages(ctx context.Context, filter *ImageFilter) ([]*Image, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var images []*Image
	for _, img := range c.images {
		// Apply filter
		if filter != nil && filter.Image != nil {
			if img.Spec.Image != filter.Image.Image {
				continue
			}
		}

		images = append(images, img)
	}

	return images, nil
}

// PullImage pulls an image from a registry
func (c *CRIService) PullImage(ctx context.Context, imageSpec *ImageSpec, auth *AuthConfig) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.log.Infof("Pulling image: %s", imageSpec.Image)

	imageID := generateID("img")
	image := &Image{
		ID:          imageID,
		RepoTags:    []string{imageSpec.Image},
		RepoDigests: []string{imageSpec.Image + "@sha256:abcd1234"},
		Size:        10485760, // 10MB simulated
		Spec:        imageSpec,
	}

	c.images[imageID] = image
	return imageID, nil
}

// RemoveImage removes an image
func (c *CRIService) RemoveImage(ctx context.Context, imageSpec *ImageSpec) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.log.Infof("Removing image: %s", imageSpec.Image)

	for id, img := range c.images {
		if img.Spec.Image == imageSpec.Image {
			delete(c.images, id)
			return nil
		}
	}

	return fmt.Errorf("image %s not found", imageSpec.Image)
}

// ImageStatus returns image status
func (c *CRIService) ImageStatus(ctx context.Context, imageSpec *ImageSpec) (*Image, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, img := range c.images {
		if img.Spec.Image == imageSpec.Image {
			return img, nil
		}
	}

	return nil, fmt.Errorf("image %s not found", imageSpec.Image)
}

// ImageFsInfo returns filesystem info
func (c *CRIService) ImageFsInfo(ctx context.Context) ([]*FilesystemUsage, error) {
	return []*FilesystemUsage{
		{
			Timestamp: time.Now().UnixNano(),
			FsID: &FilesystemIdentifier{
				Mountpoint: "/var/lib/containr",
			},
			UsedBytes: &UInt64Value{
				Value: 104857600, // 100MB simulated
			},
			InodesUsed: &UInt64Value{
				Value: 1000,
			},
		},
	}, nil
}

// Helper function to generate IDs
func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
