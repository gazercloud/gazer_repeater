//go:build !windows && !darwin
// +build !windows,!darwin

package paths

var hasVendorName = true
var systemSettingFolders string
var globalSettingFolder string
var cacheFolder string

func init() {
	// TODO:
}
