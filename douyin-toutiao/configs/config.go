package configs

import "os"

var (
	IsHeadless = true
)

func InitHeadless(isHeadless bool) {
	IsHeadless = isHeadless
}

func SetBinPath(binPath string) {
	if binPath != "" {
		os.Setenv("ROD_BROWSER_BIN", binPath)
	}
}
