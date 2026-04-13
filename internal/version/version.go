package version

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	return "ralphx " + Version + " (commit=" + Commit + ", date=" + Date + ")"
}
