package images

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/zarigata/budgie/internal/cmdutil"
)

var (
	all       bool
	digests   bool
	noTrunc   bool
	quiet     bool
	format    string
	filterArg string
)

var imagesCmd = &cobra.Command{
	Use:     "images [OPTIONS] [REPOSITORY[:TAG]]",
	Aliases: []string{"image", "ls-images", "list-images"},
	Short:   "List images",
	Long: `List images stored locally.

By default, intermediate image layers are not shown. Use -a to show all images.

Examples:
  budgie images
  budgie images nginx
  budgie images --digests
  budgie images -q`,
	RunE: listImages,
}

func listImages(cmd *cobra.Command, args []string) error {
	cmdCtx, err := cmdutil.NewCommandContext()
	if err != nil {
		return err
	}

	ctx := context.Background()

	images, err := cmdCtx.Runtime.ListImages(ctx)
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	// Filter by repository if specified
	var repoFilter string
	if len(args) > 0 {
		repoFilter = args[0]
	}

	// Filter images
	var filtered []*struct {
		Repository string
		Tag        string
		ID         string
		Created    time.Time
		Size       int64
	}

	for _, img := range images {
		repo, tag := parseImageName(img.Name)

		// Apply repository filter
		if repoFilter != "" && !strings.Contains(repo, repoFilter) {
			continue
		}

		filtered = append(filtered, &struct {
			Repository string
			Tag        string
			ID         string
			Created    time.Time
			Size       int64
		}{
			Repository: repo,
			Tag:        tag,
			ID:         img.ID,
			Created:    img.CreatedAt,
			Size:       img.Size,
		})
	}

	// Handle JSON format
	if format == "json" {
		return printJSON(filtered)
	}

	// Quiet mode - just print IDs
	if quiet {
		for _, img := range filtered {
			id := img.ID
			if !noTrunc && len(id) > 12 {
				// Extract first 12 chars of the actual digest (after sha256:)
				if strings.HasPrefix(id, "sha256:") {
					id = id[7:19]
				} else if len(id) > 12 {
					id = id[:12]
				}
			}
			fmt.Println(id)
		}
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Header
	if digests {
		fmt.Fprintln(w, "REPOSITORY\tTAG\tDIGEST\tIMAGE ID\tCREATED\tSIZE")
	} else {
		fmt.Fprintln(w, "REPOSITORY\tTAG\tIMAGE ID\tCREATED\tSIZE")
	}

	for _, img := range filtered {
		id := img.ID
		digest := ""

		if strings.HasPrefix(id, "sha256:") {
			digest = id
			if !noTrunc {
				id = id[7:19] // sha256: is 7 chars, then take 12 chars
			}
		} else if !noTrunc && len(id) > 12 {
			id = id[:12]
		}

		created := formatTimeAgo(img.Created)
		size := formatSize(img.Size)

		if digests {
			if !noTrunc && len(digest) > 20 {
				digest = digest[:20] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				img.Repository, img.Tag, digest, id, created, size)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				img.Repository, img.Tag, id, created, size)
		}
	}

	return w.Flush()
}

// parseImageName splits an image name into repository and tag
func parseImageName(name string) (string, string) {
	// Handle digest references
	if idx := strings.Index(name, "@"); idx != -1 {
		return name[:idx], "<none>"
	}

	// Split by last colon (for tag)
	if idx := strings.LastIndex(name, ":"); idx != -1 {
		// Make sure this colon is not part of a port number
		afterColon := name[idx+1:]
		if !strings.Contains(afterColon, "/") {
			return name[:idx], afterColon
		}
	}

	return name, "latest"
}

// formatTimeAgo formats a time as a human-readable "ago" string
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "Less than a minute ago"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
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

func printJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func GetImagesCmd() *cobra.Command {
	return imagesCmd
}

func init() {
	imagesCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all images (default hides intermediate images)")
	imagesCmd.Flags().BoolVar(&digests, "digests", false, "Show digests")
	imagesCmd.Flags().BoolVar(&noTrunc, "no-trunc", false, "Don't truncate output")
	imagesCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only show image IDs")
	imagesCmd.Flags().StringVar(&format, "format", "", "Output format (json)")
	imagesCmd.Flags().StringVarP(&filterArg, "filter", "f", "", "Filter output based on conditions")
}
