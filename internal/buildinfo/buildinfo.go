package buildinfo

import "strings"

const repositoryURL = "https://github.com/torikushiii/OpenGachaCodes"

// CommitHash is populated at build time with -ldflags.
var CommitHash string

func UserAgent() string {
	if hash := strings.TrimSpace(CommitHash); hash != "" {
		return "OpenGachaCodes@" + hash + " (+" + repositoryURL + ")"
	}
	return "OpenGachaCodes (+" + repositoryURL + ")"
}
