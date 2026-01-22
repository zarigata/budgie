package pull

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
)

var (
	allTags  bool
	platform string
	quiet    bool
)

var pullCmd = &cobra.Command{
	Use:   "pull <image>",
	Short: "Pull an image from a registry",
	Long: `Pull an image or a repository from a registry.

Most of your images will be created on top of a base image from a registry.
Pull is used to download an image ahead of running a container.

Examples:
  budgie pull nginx
  budgie pull nginx:latest
  budgie pull docker.io/library/alpine:3.18
  budgie pull ghcr.io/myorg/myimage:v1.0`,
	Args: cobra.ExactArgs(1),
	RunE: pullImage,
}

func pullImage(cmd *cobra.Command, args []string) error {
	imageName := args[0]

	// Normalize image name (add docker.io/library/ if no registry specified)
	imageName = normalizeImageName(imageName)

	// Initialize command context
	cmdCtx, err := cmdutil.NewCommandContext()
	if err != nil {
		return err
	}

	ctx := context.Background()

	if !quiet {
		fmt.Printf("Pulling image: %s\n", imageName)
	}

	// Pull the image
	imageInfo, err := cmdCtx.Runtime.Pull(ctx, imageName)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	if quiet {
		// Just print the digest
		fmt.Println(imageInfo.ID)
	} else {
		fmt.Printf("Pulled: %s\n", imageInfo.Name)
		fmt.Printf("Digest: %s\n", imageInfo.ID)
		fmt.Printf("Size: %s\n", formatSize(imageInfo.Size))
	}

	return nil
}

// normalizeImageName adds default registry and library if not specified
func normalizeImageName(name string) string {
	// If no tag specified, add :latest
	if !strings.Contains(name, ":") && !strings.Contains(name, "@") {
		name = name + ":latest"
	}

	// If no registry specified, add docker.io
	parts := strings.Split(name, "/")
	if len(parts) == 1 {
		// Just image name, add docker.io/library/
		return "docker.io/library/" + name
	} else if len(parts) == 2 && !strings.Contains(parts[0], ".") {
		// user/image format, add docker.io/
		return "docker.io/" + name
	}

	return name
}

// formatSize formats bytes into human-readable size
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func GetPullCmd() *cobra.Command {
	return pullCmd
}

func init() {
	pullCmd.Flags().BoolVarP(&allTags, "all-tags", "a", false, "Download all tagged images in the repository")
	pullCmd.Flags().StringVar(&platform, "platform", "", "Set platform if server is multi-platform capable")
	pullCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress verbose output")
}
