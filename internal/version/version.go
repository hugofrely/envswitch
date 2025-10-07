package version

const (
	// DevVersion is the default version string for development builds
	DevVersion = "dev"
)

// Version information - set via ldflags during build
var (
	Version   = DevVersion
	GitCommit = "unknown"
	BuildDate = "unknown"
)
