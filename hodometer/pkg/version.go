package hodometer

var (
	BuildVersion string = "0.0.0"
	BuildTime    string
	GitBranch    string
	GitCommit    string
	ReleaseType  string
)

func GetBuildDetails() map[string]interface{} {
	return map[string]interface{}{
		"BuildVersion": BuildVersion,
		"BuildTime":    BuildTime,
		"GitBranch":    GitBranch,
		"GitCommit":    GitCommit,
		"ReleaseType":  ReleaseType,
	}
}
