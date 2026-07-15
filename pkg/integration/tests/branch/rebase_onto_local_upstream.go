package branch

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var RebaseOntoLocalUpstream = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Rebase the checked-out branch onto its upstream when the upstream is a local branch",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.
			EmptyCommit("one").
			NewBranch("main").
			EmptyCommit("two").
			NewBranchFrom("feat1", "main").
			EmptyCommit("feat1-commit").
			Checkout("main").
			EmptyCommit("three").
			Checkout("feat1").
			SetBranchUpstream("feat1", "main")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().Lines(
			Contains("feat1-commit"),
			Contains("two"),
			Contains("one"),
		)

		t.Views().Branches().
			Focus().
			Lines(
				Contains("feat1").IsSelected(),
				Contains("main"),
				Contains("master"),
			).
			Press(keys.Branches.SetUpstream).
			Tap(func() {
				t.ExpectPopup().Menu().
					Title(Equals("Upstream options")).
					Select(Contains("Rebase checked-out branch onto main...")).
					Confirm()
				t.ExpectPopup().Menu().
					Title(Equals("Rebase 'feat1'")).
					Select(Contains("Simple rebase onto 'main'")).
					Confirm()
			})

		t.Views().Commits().Lines(
			Contains("feat1-commit"),
			Contains("three"),
			Contains("two"),
			Contains("one"),
		)
	},
})
