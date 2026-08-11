package branch

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var BranchesTreeViewCheckoutKeepsOrder = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Checking out a branch in the tree view leaves it in place instead of moving it ahead of its siblings",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		// The default "date" order would tie here, because every branch below is
		// created within the same second.
		config.GetUserConfig().Git.LocalBranchSortOrder = "alphabetical"
		config.GetUserConfig().Gui.ShowBranchTree = true
	},
	SetupRepo: func(shell *Shell) {
		shell.
			EmptyCommit("initial").
			NewBranch("main").
			EmptyCommit("main-1").
			NewBranchFrom("feat1", "main").
			EmptyCommit("feat1-1").
			NewBranchFrom("feat2", "main").
			EmptyCommit("feat2-1").
			NewBranchFrom("feat3", "main").
			EmptyCommit("feat3-1").
			SetBranchUpstream("feat1", "main").
			SetBranchUpstream("feat2", "main").
			SetBranchUpstream("feat3", "main").
			Checkout("main")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			Focus().
			Lines(
				Contains("main").IsSelected(),
				Contains("├─ feat1"),
				Contains("├─ feat2"),
				Contains("└─ feat3"),
				Contains("master"),
			).
			// Checking out the last sibling keeps the tree exactly as it was.
			NavigateToLine(Contains("feat3")).
			Press(keys.Universal.Select).
			Lines(
				Contains("main"),
				Contains("├─ feat1"),
				Contains("├─ feat2"),
				Contains("└─ feat3").IsSelected(),
				Contains("master"),
			)
	},
})
