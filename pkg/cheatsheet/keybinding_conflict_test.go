package cheatsheet

import (
	"fmt"
	"testing"

	"github.com/jesseduffield/lazygit/pkg/app"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// A binding on a view wins over a binding on the global context, so the global
// action becomes unreachable in that panel. That is often what we want, but it
// is silent: nothing tells you that the key you picked already means something
// everywhere else.
//
// Every deliberate override is listed here, keyed by "<view>:<key>". The value
// is a note for the reader and is not asserted on. An override that is missing
// from this map fails the test, which turns picking such a key into a decision
// someone has to make on purpose.
//
// <esc> is exempt: backing out is defined per panel, and the global binding is
// only the fallback for panels that have nothing more specific to do.
var expectedGlobalOverrides = map[string]string{
	"commitDescription:<ctrl+s>": "confirms the commit message while writing it",
	"commits:<ctrl+r>":           "resets the cherry-picked commits selection",
	"commits:R":                  "rewords a commit with the editor",
	"commits:p":                  "picks a commit during an interactive rebase",
	"localBranches:R":            "renames a branch",
	"mergeConflicts:z":           "undoes a conflict resolution rather than a git command",
	"reflogCommits:<ctrl+r>":     "resets the cherry-picked commits selection",
	"subCommits:<ctrl+r>":        "resets the cherry-picked commits selection",
	"tags:P":                     "pushes the selected tag rather than the current branch",
}

const escapeKeyLabel = "<esc>"

// globalOverrides returns the keys, in "<view>:<key>" form, that a view binds
// while the global context binds them too.
func globalOverrides(bindings []*types.Binding) map[string]bool {
	globalKeys := map[gocui.Key]bool{}
	for _, binding := range bindings {
		if binding.ViewName == "" {
			for _, key := range binding.Keys {
				globalKeys[key] = true
			}
		}
	}

	overrides := map[string]bool{}
	for _, binding := range bindings {
		if binding.ViewName == "" {
			continue
		}
		for _, key := range binding.Keys {
			label := keyLabels([]gocui.Key{key})
			if globalKeys[key] && label != escapeKeyLabel {
				overrides[fmt.Sprintf("%s:%s", binding.ViewName, label)] = true
			}
		}
	}

	return overrides
}

func TestViewKeybindingsDontShadowGlobalsByAccident(t *testing.T) {
	appConfig := config.NewDummyAppConfig()
	common, err := app.NewCommon(appConfig)
	assert.NoError(t, err)
	testApp, err := app.NewApp(appConfig, nil, common)
	assert.NoError(t, err)

	actual := lo.Keys(globalOverrides(testApp.Gui.GetCheatsheetKeybindings()))
	expected := lo.Keys(expectedGlobalOverrides)

	assert.ElementsMatch(t, expected, actual,
		"A view keybinding shadows a global one. If that is intended, add it to "+
			"expectedGlobalOverrides; otherwise pick a key that is free in that panel.")
}
