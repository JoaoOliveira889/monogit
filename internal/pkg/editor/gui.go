package editor

import (
	"os/exec"
)

type GUILauncher struct {
	Editor string
}

func (l *GUILauncher) Launch(path string) error {
	cmd := exec.Command(l.Editor, path)
	return cmd.Start()
}
