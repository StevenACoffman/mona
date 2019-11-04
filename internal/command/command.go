// Package command contains methods called by the CLI to manage
// a mona project.
package command

import (
	"fmt"

	"github.com/apex/log"
	"github.com/hashicorp/go-multierror"

	"github.com/uw-labs/mona/internal/app"
	"github.com/uw-labs/mona/internal/config"
	"github.com/uw-labs/mona/internal/hash"
)

type Config struct {
	Project  *config.Project
	FailFast bool
}

type (
	changeType int
	rangeFn    func(*app.App, map[changeType]bool) error
	changedApp struct {
		app         *app.App
		newHash     string
		changeTypes map[changeType]bool
	}
)

const (
	changeTypeLint changeType = iota
	changeTypeTest
	changeTypeBuild
)

func getLockAndChangedApps(pj *config.Project) (lock *config.LockFile, out []changedApp, err error) {
	apps, err := app.FindApps("./", pj.Mod)
	if err != nil {
		return nil, nil, err
	}

	lock, err = config.LoadLockFile(pj.Location)
	if err != nil {
		return nil, nil, err
	}

	for _, appInfo := range apps {
		lockInfo, ok := lock.Apps[appInfo.Name]
		if !ok {
			lockInfo = &config.AppVersion{}
		}

		// GenerateString a new hash for the app directory
		exclude := append(pj.Exclude, appInfo.Exclude...)
		newHash, err := hash.GetForApp(appInfo, exclude...)
		if err != nil {
			return nil, nil, err
		}

		changes := make(map[changeType]bool, 3)

		if lockInfo.LintHash != newHash {
			changes[changeTypeLint] = true
		}
		if lockInfo.TestHash != newHash {
			changes[changeTypeTest] = true
		}
		if lockInfo.BuildHash != newHash {
			changes[changeTypeBuild] = true
		}
		out = append(out, changedApp{
			app:         appInfo,
			newHash:     newHash,
			changeTypes: changes,
		})
	}

	return lock, out, nil
}

func rangeChangedApps(cfg Config, cts []changeType, fn rangeFn) error {
	lock, changed, err := getLockAndChangedApps(cfg.Project)

	if err != nil || len(changed) == 0 {
		return err
	}

	var errs []error
	for _, change := range changed {
		if len(cts) == 1 && !change.changeTypes[cts[0]] {
			continue
		}
		if err := fn(change.app, change.changeTypes); err != nil {
			errs = append(errs, fmt.Errorf("app %s: %s", change.app.Name, err.Error()))

			if cfg.FailFast {
				return errs[0]
			}
			continue
		}

		lockInfo, modInLock := lock.Apps[change.app.Name]

		if !modInLock {
			log.Debugf("Detected new appInfo %s at %s, adding to lock file", change.app.Name, change.app.Location)

			if err := config.AddApp(lock, cfg.Project.Location, change.app.Name); err != nil {
				return err
			}

			lockInfo = lock.Apps[change.app.Name]
		}

		for _, ct := range cts {
			switch ct {
			case changeTypeLint:
				lockInfo.LintHash = change.newHash
			case changeTypeTest:
				lockInfo.TestHash = change.newHash
			case changeTypeBuild:
				lockInfo.BuildHash = change.newHash
			}
		}

		if err := config.UpdateLockFile(cfg.Project.Location, lock); err != nil {
			return err
		}
	}

	return multierror.Append(nil, errs...).ErrorOrNil()
}
