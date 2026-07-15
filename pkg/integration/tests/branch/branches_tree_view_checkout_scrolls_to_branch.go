package branch

import (
	"fmt"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var BranchesTreeViewCheckoutScrollsToBranch = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Checking out a branch far down the tree scrolls it into view rather than jumping to the top of the list",
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
		// Five stacks of three branches each, so the branch we check out at the
		// end sits near the bottom of a list that is longer than the view.
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
		shell.Checkout("master")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			Focus().
			NavigateToLine(Contains("stack5-c")).
			Press(keys.Universal.Select).
			Tap(func() {
				t.Views().Branches().
					OriginYAtLeast(1).
					ContainsLines(Contains("stack5-c").IsSelected())
			})
	},
})
