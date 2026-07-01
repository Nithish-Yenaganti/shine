package version

var (
	Version = "0.1.0"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	if Commit == "unknown" && Date == "unknown" {
		return Version
	}
	return Version + " (" + Commit + ", " + Date + ")"
}
