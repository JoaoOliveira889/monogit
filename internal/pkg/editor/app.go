package editor

import (
	"fmt"
	"os/exec"
)

type AppLauncher struct {
	AppName string
}

func (l *AppLauncher) Launch(path string) error {
	if err := ValidateAppName(l.AppName); err != nil {
		return fmt.Errorf("invalid app name: %w", err)
	}
	cmd := exec.Command("open", "-a", l.AppName, path)
	return cmd.Run()
}
