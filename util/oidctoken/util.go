package oidctoken

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func openBrowser(url string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		fmt.Fprintf(os.Stderr, "please open this link on your web browser\n%s\n", url)
	}
}

func inArray(what string, data []string) bool {
	for _, v := range data {
		if v == what {
			return true
		}
	}
	return false
}
