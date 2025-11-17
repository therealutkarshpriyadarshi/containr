package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/errors"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
	"github.com/therealutkarshpriyadarshi/containr/pkg/registry"
)

// images command
var (
	imagesAll     bool
	imagesQuiet   bool
	imagesNoTrunc bool
)

var imagesCmd = &cobra.Command{
	Use:   "images [flags]",
	Short: "List images",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement image listing from local storage
		fmt.Println("IMAGE\t\tTAG\t\tIMAGE ID\tCREATED\t\tSIZE")
		fmt.Println("(Image listing not yet implemented)")
		return nil
	},
}

func init() {
	imagesCmd.Flags().BoolVarP(&imagesAll, "all", "a", false, "Show all images")
	imagesCmd.Flags().BoolVarP(&imagesQuiet, "quiet", "q", false, "Only show image IDs")
	imagesCmd.Flags().BoolVar(&imagesNoTrunc, "no-trunc", false, "Don't truncate output")
}

// rmi command
var (
	rmiForce bool
)

var rmiCmd = &cobra.Command{
	Use:   "rmi [flags] IMAGE [IMAGE...]",
	Short: "Remove one or more images",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, image := range args {
			fmt.Printf("Removing image %s...\n", image)
			// TODO: Implement image removal
			fmt.Println("(Image removal not yet implemented)")
		}
		return nil
	},
}

func init() {
	rmiCmd.Flags().BoolVarP(&rmiForce, "force", "f", false, "Force removal")
}

// import command
var importCmd = &cobra.Command{
	Use:   "import FILE|- [REPOSITORY[:TAG]]",
	Short: "Import the contents from a tarball to create a filesystem image",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tarball := args[0]
		var repository string
		if len(args) > 1 {
			repository = args[1]
		}

		fmt.Printf("Importing from %s", tarball)
		if repository != "" {
			fmt.Printf(" as %s", repository)
		}
		fmt.Println()
		fmt.Println("(Import not yet fully implemented)")

		return nil
	},
}

// export command
var exportCmd = &cobra.Command{
	Use:   "export [flags] CONTAINER",
	Short: "Export a container's filesystem as a tar archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := args[0]
		fmt.Printf("Exporting container %s\n", container)
		fmt.Println("(Export not yet fully implemented)")
		return nil
	},
}

// pull command
var (
	pullQuiet bool
)

var pullCmd = &cobra.Command{
	Use:   "pull [flags] IMAGE[:TAG|@DIGEST]",
	Short: "Pull an image from a registry",
	Long: `Downloads an image from a container registry.

Supports Docker Hub and OCI-compatible registries.`,
	Example: `  containr pull alpine
  containr pull alpine:3.14
  containr pull ubuntu:latest
  containr pull gcr.io/project/image:tag`,
	Args: cobra.ExactArgs(1),
	RunE: pullImage,
}

func init() {
	pullCmd.Flags().BoolVarP(&pullQuiet, "quiet", "q", false, "Suppress verbose output")
}

func pullImage(cmd *cobra.Command, args []string) error {
	log := logger.New("pull")
	imageRef := args[0]

	log.Infof("Pulling image %s", imageRef)

	// Parse image reference
	ref, err := registry.ParseImageReference(imageRef)
	if err != nil {
		return errors.Wrap(errors.ErrInvalidArgument, "invalid image reference", err).
			WithField("image", imageRef)
	}

	fmt.Printf("Pulling %s...\n", ref.String())

	// Create registry client
	client := registry.DefaultClient()

	// Pull image
	opts := &registry.PullOptions{
		DestDir: fmt.Sprintf("/var/lib/containr/images/%s", ref.Repository),
		Verbose: !pullQuiet,
	}

	if err := client.Pull(ref, opts); err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to pull image", err).
			WithField("image", ref.String()).
			WithHint("Check your network connection and registry access")
	}

	fmt.Printf("Successfully pulled %s\n", ref.String())

	// Extract to rootfs (optional)
	if !pullQuiet {
		fmt.Println("Extracting image layers...")
	}

	rootfsDir := fmt.Sprintf("/var/lib/containr/rootfs/%s-%s", ref.Repository, ref.Tag)
	if err := registry.ExtractImageToRootFS(opts.DestDir, rootfsDir); err != nil {
		log.WithError(err).Warn("Failed to extract image")
		fmt.Fprintf(os.Stderr, "Warning: Failed to extract image: %v\n", err)
	} else if !pullQuiet {
		fmt.Printf("Image extracted to %s\n", rootfsDir)
	}

	log.Info("Image pull complete")
	return nil
}

// Helper functions for image management

func formatImageSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	} else if bytes < GB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
}

func printImageTable(images []ImageInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REPOSITORY\tTAG\tIMAGE ID\tCREATED\tSIZE")

	for _, img := range images {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			img.Repository, img.Tag, img.ID, img.Created, formatImageSize(img.Size))
	}

	w.Flush()
}

// ImageInfo represents image information for display
type ImageInfo struct {
	Repository string
	Tag        string
	ID         string
	Created    string
	Size       int64
}
