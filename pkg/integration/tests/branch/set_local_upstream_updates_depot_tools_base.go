package branch

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SetLocalUpstreamUpdatesDepotToolsBase = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Reparenting a branch onto another local branch migrates depot_tools' rebase-update base metadata",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.
			EmptyCommit("initial commit").
			NewBranch("main").
			NewBranchFrom("old-parent", "main").
			EmptyCommit("old parent commit").
			NewBranchFrom("feature", "old-parent").
			EmptyCommit("feature commit").
			RunCommand([]string{"git", "branch", "--set-upstream-to=old-parent", "feature"}).
			RunShellCommand("git config branch.feature.base $(git rev-parse old-parent)").
			RunShellCommand("git config branch.feature.base-upstream old-parent")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			Focus().
			Press(keys.Universal.NextScreenMode).
			SelectedLines(
				Contains("feature").Contains(". old-parent"),
			).
			Press(keys.Branches.SetUpstream).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Upstream options")).
					Select(Contains(" Set upstream of selected branch")).
					Confirm()

				t.ExpectPopup().Prompt().
					Title(Equals("Enter upstream as '<remote> <branchname>' or '<local branchname>'")).
					Type("main").
					Confirm()
			}).
			SelectedLines(
				Contains("feature").Contains(". main"),
			)

		// The recorded base still marks the start of the branch's own commits,
		// and base-upstream follows the new parent.
		t.Git().
			ConfigValue("branch.feature.base", t.Git().GetCommitHash("old-parent")).
			ConfigValue("branch.feature.base-upstream", "main")
	},
})
