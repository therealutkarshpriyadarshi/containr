// Package version provides version information for containr
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version
	Version = "1.0.0"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildDate is the build date
	BuildDate = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()

	// Platform is the OS/Arch combination
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info represents version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  Platform,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("containr version %s\n"+
		"  Git commit: %s\n"+
		"  Build date: %s\n"+
		"  Go version: %s\n"+
		"  Platform:   %s",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.Platform)
}

// Short returns a short version string
func (i Info) Short() string {
	return fmt.Sprintf("containr %s (%s)", i.Version, i.GitCommit[:7])
}

// UserAgent returns a user agent string
func (i Info) UserAgent() string {
	return fmt.Sprintf("containr/%s (%s)", i.Version, i.Platform)
}
