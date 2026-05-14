package editor

type Launcher interface {
	Launch(path string) error
}
