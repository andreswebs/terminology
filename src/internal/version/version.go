package version

import "runtime/debug"

var Override = ""

func Current() string {
	if Override != "" {
		return Override
	}
	// BuildInfo.Main.Version is populated by `go install ...@vX.Y.Z` but
	// reports "(devel)" for bare `go build`; integration coverage for the
	// real-version branch lives in E10's release tests.
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}
