package editor

import (
	"os/exec"
)

type AppLauncher struct {
	AppName string
}

func (l *AppLauncher) Launch(path string) error {
	cmd := exec.Command("open", "-a", l.AppName, path)
	return cmd.Run()
}
