// Package version reports the build version of the tool.
package version

import "runtime/debug"

// Override forces Current to return this value instead of the version derived
// from build info when it is non-empty.
var Override = ""

// Current returns the tool's version: Override if set, otherwise the module
// version from build info, falling back to "dev".
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
