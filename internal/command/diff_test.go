package command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uw-labs/mona/internal/command"
)

func TestDiff(t *testing.T) {
	tt := []struct {
		Name          string
		AppDirs       []string
		ExpectedDiffs []string
	}{
		{
			Name:          "It should detect changes in apps",
			AppDirs:       []string{"test/a", "test/b"},
			ExpectedDiffs: []string{"a"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			pj := setupProject(t)
			setupApps(t, pj, tc.AppDirs...)

			defer deleteProjectFiles(t)
			defer deleteAppFiles(t, tc.AppDirs...)

			summary, err := command.Diff(pj)

			if err != nil {
				assert.Fail(t, err.Error())
				return
			}

			for _, exp := range tc.ExpectedDiffs {
				assert.Contains(t, summary.Build, exp)
				assert.Contains(t, summary.Test, exp)
				assert.Contains(t, summary.Lint, exp)
			}
		})
	}
}
