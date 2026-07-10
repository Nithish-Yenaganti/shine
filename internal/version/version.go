package version

var (
	Version = "0.1.1"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	if Commit == "unknown" && Date == "unknown" {
		return Version
	}
	return Version + " (" + Commit + ", " + Date + ")"
}
