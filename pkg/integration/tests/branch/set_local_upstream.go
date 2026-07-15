package branch

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SetLocalUpstream = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Set a local branch as the upstream of another local branch",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.
			EmptyCommit("initial commit").
			NewBranch("main").
			NewBranchFrom("feature", "main")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			Focus().
			Press(keys.Universal.NextScreenMode).
			SelectedLines(
				Contains("feature").DoesNotContain(". main"),
			).
			Press(keys.Branches.SetUpstream).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Upstream options")).
					Select(Contains(" Set upstream of selected branch")).
					Confirm()

				t.ExpectPopup().Prompt().
					/* EXPECTED:
					Title(Equals("Enter upstream as '<remote> <branchname>' or '<local branchname>'")).
					SuggestionTopLines(Equals("main")).
					ACTUAL: */
					Title(Equals("Enter upstream as '<remote> <branchname>'")).
					Type("main").
					Confirm()
			}).
			/* EXPECTED:
			SelectedLines(
				Contains("feature").Contains(". main"),
			)
			ACTUAL: */
			Tap(func() {
				t.ExpectPopup().Alert().
					Title(Equals("Error")).
					Content(Equals("Invalid upstream. Must be in the format '<remote> <branchname>'")).
					Confirm()
			})
	},
})
