package command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uw-labs/mona/internal/app"
	"github.com/uw-labs/mona/internal/command"
)

func TestAddApp(t *testing.T) {
	tt := []struct {
		Name        string
		AppName     string
		AppLocation string
	}{
		{
			Name:        "It should create an app",
			AppName:     "test",
			AppLocation: ".",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			pj := setupProject(t)
			defer deleteProjectFiles(t)
			defer deleteAppFile(t, tc.AppLocation)

			if err := command.AddApp(pj, tc.AppName, tc.AppLocation); err != nil {
				assert.Fail(t, err.Error())
				return
			}

			assert.FileExists(t, "app.yml")
			mod, err := app.LoadApp(tc.AppLocation)

			if err != nil {
				assert.Fail(t, err.Error())
				return
			}

			assert.Equal(t, tc.AppName, mod.Name)
		})
	}
}
