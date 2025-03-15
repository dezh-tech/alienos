package main

import "fmt"

// These constants follow the semantic versioning 2.0.0 spec.
// see: http://semver.org
var (
	major = 1
	minor = 4
	patch = 0
	meta  = ""
)

func StringVersion() string {
	v := fmt.Sprintf("alienos - %d.%d.%d", major, minor, patch)

	if meta != "" {
		v = fmt.Sprintf("%s-%s", v, meta)
	}

	return v
}
