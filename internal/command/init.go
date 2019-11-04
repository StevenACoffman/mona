package command

import "github.com/uw-labs/mona/internal/config"

// Init creates a new project and lock file in the provided working directory
// with the given name.
func Init(wd, name string) error {
	if err := config.NewProject(wd, name); err != nil {
		return err
	}

	return config.NewLockFile(wd, name)
}
