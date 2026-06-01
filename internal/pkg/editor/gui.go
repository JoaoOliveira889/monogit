package editor

import (
	"fmt"
	"os/exec"
)

type GUILauncher struct {
	Spec CommandSpec
}

func (l *GUILauncher) Launch(path string) error {
	if l.Spec.Name == "" {
		return fmt.Errorf("empty editor command")
	}
	args := append(append([]string{}, l.Spec.Args...), path)
	cmd := exec.Command(l.Spec.Name, args...)
	return cmd.Start()
}
