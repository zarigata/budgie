package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/leases/leaseutil"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/snapshotter"
	"github.com/opencontainers/image-spec/identity"
	"github.com/opencontainers/image-spec/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"

	"github.com/zarigata/budgie/pkg/types"
)

type Runtime interface {
	Create(ctx context.Context, ctr *types.Container) error
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, timeout time.Duration) error
	Delete(ctx context.Context, id string) error
	Exists(id string) bool
	Status(ctx context.Context, id string) (string, error)
}

type containerdRuntime struct {
	client *containerd.Client
}

func NewContainerdRuntime(address string) (Runtime, error) {
	client, err := containerd.New(address, containerd.WithDefaultNamespace("budgie"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	return &containerdRuntime{client: client}, nil
}

func (r *containerdRuntime) Create(ctx context.Context, ctr *types.Container) error {
	imageName := ctr.Image.DockerImage
	image, err := r.client.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}

	logrus.Infof("Pulled image %s", imageName)

	// Build OCI spec options
	opts := []oci.SpecOpts{
		oci.WithImageConfig(image),
	}

	// Set working directory
	if ctr.Image.WorkDir != "" {
		opts = append(opts, oci.WithProcessCwd(ctr.Image.WorkDir))
	}

	// Set command
	if len(ctr.Image.Command) > 0 {
		opts = append(opts, oci.WithProcessArgs(ctr.Image.Command...))
	}

	// Set environment variables
	if len(ctr.Env) > 0 {
		opts = append(opts, oci.WithEnv(ctr.Env))
	}

	// Add port environment variables
	for _, port := range ctr.Ports {
		opts = append(opts, oci.WithEnv([]string{
			fmt.Sprintf("PORT_%d=%d", port.ContainerPort, port.ContainerPort),
		}))
	}

	// Apply resource limits
	if ctr.Resources != nil {
		opts = append(opts, withResourceLimits(ctr.Resources))
	}

	// Apply volume mounts
	if len(ctr.Volumes) > 0 {
		opts = append(opts, withVolumeMounts(ctr.Volumes))
	}

	container, err := r.client.NewContainer(
		ctx,
		ctr.ID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(ctr.ID+"-snapshot", image),
		containerd.WithNewSpec(opts...),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	logrus.Infof("Created container %s", ctr.ShortID())
	_ = container // Avoid unused variable warning
	return nil
}

// withResourceLimits creates OCI spec options for resource limits
func withResourceLimits(res *types.ResourceLimits) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Linux == nil {
			s.Linux = &specs.Linux{}
		}
		if s.Linux.Resources == nil {
			s.Linux.Resources = &specs.LinuxResources{}
		}

		// CPU limits
		if res.CPUShares > 0 || res.CPUQuota > 0 {
			if s.Linux.Resources.CPU == nil {
				s.Linux.Resources.CPU = &specs.LinuxCPU{}
			}
			if res.CPUShares > 0 {
				shares := uint64(res.CPUShares)
				s.Linux.Resources.CPU.Shares = &shares
			}
			if res.CPUQuota > 0 {
				s.Linux.Resources.CPU.Quota = &res.CPUQuota
			}
		}

		// Memory limits
		if res.MemoryLimit > 0 || res.MemorySwap > 0 {
			if s.Linux.Resources.Memory == nil {
				s.Linux.Resources.Memory = &specs.LinuxMemory{}
			}
			if res.MemoryLimit > 0 {
				s.Linux.Resources.Memory.Limit = &res.MemoryLimit
			}
			if res.MemorySwap > 0 {
				s.Linux.Resources.Memory.Swap = &res.MemorySwap
			}
		}

		// Block I/O weight
		if res.BlkioWeight > 0 {
			if s.Linux.Resources.BlockIO == nil {
				s.Linux.Resources.BlockIO = &specs.LinuxBlockIO{}
			}
			s.Linux.Resources.BlockIO.Weight = &res.BlkioWeight
		}

		// PIDs limit
		if res.PidsLimit > 0 {
			if s.Linux.Resources.Pids == nil {
				s.Linux.Resources.Pids = &specs.LinuxPids{}
			}
			s.Linux.Resources.Pids.Limit = res.PidsLimit
		}

		return nil
	}
}

// withVolumeMounts creates OCI spec options for volume mounts
func withVolumeMounts(volumes []types.VolumeMapping) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		for _, vol := range volumes {
			options := []string{"rbind"}
			if vol.Mode == "ro" {
				options = append(options, "ro")
			} else {
				options = append(options, "rw")
			}

			mount := specs.Mount{
				Destination: vol.Target,
				Source:      vol.Source,
				Type:        "bind",
				Options:     options,
			}
			s.Mounts = append(s.Mounts, mount)
		}
		return nil
	}
}

func (r *containerdRuntime) Start(ctx context.Context, id string) error {
	container, err := r.client.LoadContainer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Start the task first
	if err := task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	// Setup wait channel for exit monitoring (non-blocking)
	go func() {
		exitCh, err := task.Wait(ctx)
		if err != nil {
			logrus.Errorf("Failed to setup wait channel for container %s: %v", id[:12], err)
			return
		}
		exitStatus := <-exitCh
		logrus.Infof("Container %s exited with status %d", id[:12], exitStatus.ExitCode())
	}()

	logrus.Infof("Started container %s", id[:12])
	return nil
}

func (r *containerdRuntime) Stop(ctx context.Context, id string, timeout time.Duration) error {
	container, err := r.client.LoadContainer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		// No task means container is already stopped
		logrus.Debugf("Container %s has no running task", id[:12])
		return nil
	}

	// Graceful shutdown: SIGTERM first
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		logrus.Warnf("Failed to send SIGTERM to container %s: %v", id[:12], err)
	}

	// Wait for graceful shutdown with timeout
	waitCh, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup wait channel: %w", err)
	}

	select {
	case <-waitCh:
		logrus.Infof("Container %s stopped gracefully", id[:12])
	case <-time.After(timeout):
		// Force kill with SIGKILL
		logrus.Warnf("Container %s did not stop gracefully, sending SIGKILL", id[:12])
		if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
			logrus.Errorf("Failed to send SIGKILL to container %s: %v", id[:12], err)
		}
		<-waitCh
	}

	// Delete the task
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	logrus.Infof("Stopped container %s", id[:12])
	return nil
}

func (r *containerdRuntime) Delete(ctx context.Context, id string) error {
	container, err := r.client.LoadContainer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	// Try to stop any running task first
	task, err := container.Task(ctx, nil)
	if err == nil {
		// Task exists, try to delete it
		if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil {
			logrus.Warnf("Failed to delete task for container %s: %v", id[:12], err)
		}
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	logrus.Infof("Deleted container %s", id[:12])
	return nil
}

func (r *containerdRuntime) Exists(id string) bool {
	ctx := namespaces.WithNamespace(context.Background(), "budgie")
	_, err := r.client.LoadContainer(ctx, id)
	return err == nil
}

func (r *containerdRuntime) Status(ctx context.Context, id string) (string, error) {
	container, err := r.client.LoadContainer(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "stopped", nil
	}

	status, err := task.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get task status: %w", err)
	}

	return string(status.Status), nil
}

func GetDefaultRuntime() (Runtime, error) {
	address := os.Getenv("CONTAINERD_ADDRESS")
	if address == "" {
		address = "/run/containerd/containerd.sock"
	}

	return NewContainerdRuntime(address)
}
