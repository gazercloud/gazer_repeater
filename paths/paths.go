package paths

import "os"

func ProgramDataFolder() string {
	return systemSettingFolders
}

func FileIsExists(filePath string) (exists bool) {
	exists = true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}
	return
}
