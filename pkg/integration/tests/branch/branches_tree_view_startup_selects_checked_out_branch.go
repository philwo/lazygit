package branch

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var BranchesTreeViewStartupSelectsCheckedOutBranch = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Start up with the checked-out branch selected and scrolled into view, even when it sits far down the tree",
	ExtraCmdArgs: []string{},
	// A short terminal, so that the branches below don't all fit in the view.
	Width:  100,
	Height: 20,
	Skip:   false,
	SetupConfig: func(config *config.AppConfig) {
		// The default "date" order would tie here, because every branch below is
		// created within the same second.
		config.GetUserConfig().Git.LocalBranchSortOrder = "alphabetical"
		config.GetUserConfig().Gui.ShowBranchTree = true
	},
	SetupRepo: func(shell *Shell) {
		// Five stacks of three branches each, with the last one checked out, so
		// that it sits near the bottom of a list that is longer than the view.
		shell.EmptyCommit("initial")
		for i := 1; i <= 5; i++ {
			root := fmt.Sprintf("stack%d", i)
			shell.NewBranchFrom(root, "master").EmptyCommit(root)
			for _, suffix := range []string{"a", "b", "c"} {
				name := root + "-" + suffix
				shell.NewBranchFrom(name, root).
					EmptyCommit(name).
					SetBranchUpstream(name, root)
			}
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			OriginYAtLeast(1).
			ContainsLines(Contains("stack5-c").IsSelected())
	},
})
