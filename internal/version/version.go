package version

import "fmt"

// These values are populated by the linker using -ldflags "-X version.major=x -X version.minor=x -X version.patch -X version.commit=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" -X version.date="xxxx-xx-xx"
var (
	major  string
	minor  string
	patch  string
	commit string
	date   string
)

func String() string {
	if major == "" && minor == "" && patch == "" && commit == "" && date == "" {
		return "development"
	}

	return fmt.Sprintf("%s.%s.%s-%s (%s)", major, minor, patch, commit, date)
}
