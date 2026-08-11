package branch

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var BranchesTreeView = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Toggle the local branches view between a flat list and a tree of stacked branches, and collapse/expand sub-stacks",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		// The default "date" order would tie here, because every branch below is
		// created within the same second.
		config.GetUserConfig().Git.LocalBranchSortOrder = "alphabetical"
	},
	SetupRepo: func(shell *Shell) {
		// Build a stack: feat1->main, feat1b->feat1, topic->main, hotfix->develop,
		// with main, develop, and master as roots. Under the alphabetical order
		// pinned above, main's children come out as [feat1, topic] (feat1 is not
		// the last child, which exercises the "│" vertical continuation under
		// it).
		shell.
			EmptyCommit("initial").
			NewBranch("main").
			EmptyCommit("main-1").
			NewBranchFrom("feat1", "main").
			EmptyCommit("feat1-1").
			NewBranchFrom("feat1b", "feat1").
			EmptyCommit("feat1b-1").
			NewBranchFrom("topic", "main").
			EmptyCommit("topic-1").
			NewBranchFrom("develop", "master").
			EmptyCommit("develop-1").
			NewBranchFrom("hotfix", "develop").
			EmptyCommit("hotfix-1").
			SetBranchUpstream("feat1", "main").
			SetBranchUpstream("feat1b", "feat1").
			SetBranchUpstream("topic", "main").
			SetBranchUpstream("hotfix", "develop").
			Checkout("main")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Branches().
			Focus().
			Lines(
				Contains("main").IsSelected(),
				Contains("develop"),
				Contains("feat1"),
				Contains("feat1b"),
				Contains("hotfix"),
				Contains("master"),
				Contains("topic"),
			).
			// Switch to tree mode: stacked branches nest under their parents
			// with connector glyphs, and the order comes from the structure
			// plus the configured sort order, which puts develop's stack above
			// main's.
			Press(keys.Branches.ToggleTreeView).
			Lines(
				Contains("develop"),
				Contains("└─ hotfix"),
				Contains("main").IsSelected(),
				Contains("├─ feat1"),
				Contains("│  └─ feat1b"),
				Contains("└─ topic"),
				Contains("master"),
			).
			// Collapse the selected branch's sub-stack: feat1b hides.
			NavigateToLine(Contains("├─ feat1")).
			Press(keys.Branches.ToggleBranchCollapsed).
			Lines(
				Contains("develop"),
				Contains("└─ hotfix"),
				Contains("main"),
				Contains("├─ feat1").IsSelected(),
				Contains("└─ topic"),
				Contains("master"),
			).
			// Expand it again: feat1b reappears.
			Press(keys.Branches.ToggleBranchCollapsed).
			Lines(
				Contains("develop"),
				Contains("└─ hotfix"),
				Contains("main"),
				Contains("├─ feat1").IsSelected(),
				Contains("│  └─ feat1b"),
				Contains("└─ topic"),
				Contains("master"),
			).
			// Collapse all sub-stacks from a leaf: only roots stay visible, and
			// the hidden selection moves to its nearest visible ancestor (main).
			NavigateToLine(Contains("feat1b")).
			Press(keys.Branches.CollapseAllBranches).
			Lines(
				Contains("develop"),
				Contains("main").IsSelected(),
				Contains("master"),
			).
			// Expand all sub-stacks: everything is visible again.
			Press(keys.Branches.ExpandAllBranches).
			Lines(
				Contains("develop"),
				Contains("└─ hotfix"),
				Contains("main").IsSelected(),
				Contains("├─ feat1"),
				Contains("│  └─ feat1b"),
				Contains("└─ topic"),
				Contains("master"),
			).
			// Back to flat mode: no connectors, selection preserved by name.
			NavigateToLine(Contains("feat1b")).
			Press(keys.Branches.ToggleTreeView).
			Lines(
				Contains("main"),
				Contains("develop"),
				Contains("feat1"),
				Contains("feat1b").IsSelected(),
				Contains("hotfix"),
				Contains("master"),
				Contains("topic"),
			)
	},
})
