package context

import (
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/branchtree"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

type BranchesContext struct {
	*FilteredListViewModel[*models.Branch]
	*ListContextTrait
	tree *branchtree.BranchTree
}

var (
	_ types.IListContext    = (*BranchesContext)(nil)
	_ types.DiffableContext = (*BranchesContext)(nil)
)

func NewBranchesContext(c *ContextCommon) *BranchesContext {
	tree := branchtree.New(c.UserConfig().Gui.ShowBranchTree)

	viewModel := NewFilteredListViewModel(
		func() []*models.Branch { return tree.VisibleBranches(c.Model().Branches) },
		func(branch *models.Branch) []string {
			return []string{branch.Name}
		},
	)

	getDisplayStrings := func(_ int, _ int) [][]string {
		return presentation.GetBranchListDisplayStrings(
			viewModel.GetItems(),
			c.State().GetItemOperation,
			c.Model().PullRequestsMap,
			c.State().GetRepoState().GetScreenMode() != types.SCREEN_NORMAL,
			c.Modes().Diffing.Ref,
			c.Views().Branches.InnerWidth()+c.Views().Branches.OriginX(),
			c.Tr,
			c.UserConfig(),
			c.Model().Worktrees,
			treePrefixes(tree, viewModel, c.Model().Branches),
		)
	}

	self := &BranchesContext{
		tree:                  tree,
		FilteredListViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				View:                       c.Views().Branches,
				WindowName:                 "branches",
				Key:                        LOCAL_BRANCHES_CONTEXT_KEY,
				Kind:                       types.SIDE_CONTEXT,
				Focusable:                  true,
				NeedsRerenderOnWidthChange: types.NEEDS_RERENDER_ON_WIDTH_CHANGE_WHEN_WIDTH_CHANGES,
			})),
			ListRenderer: ListRenderer{
				list:              viewModel,
				getDisplayStrings: getDisplayStrings,
			},
			c: c,
		},
	}

	return self
}

// treePrefixes returns the map of branch name to connector prefix for the
// current view. It is nil in flat mode, and also while a search filter is
// active: fuzzy matching can drop intermediate parents, which would leave the
// connectors pointing at nothing, so we fall back to a flat rendering.
func treePrefixes(tree *branchtree.BranchTree, viewModel *FilteredListViewModel[*models.Branch], branches []*models.Branch) map[string]string {
	if !tree.InTreeMode() || viewModel.IsFiltering() {
		return nil
	}

	prefixes := make(map[string]string)
	for _, node := range tree.Build(branches) {
		prefixes[node.Branch.Name] = node.Prefix
	}
	return prefixes
}

func (self *BranchesContext) InTreeMode() bool {
	return self.tree.InTreeMode()
}

func (self *BranchesContext) ToggleShowTree() {
	self.tree.ToggleShowTree()
}

func (self *BranchesContext) ToggleBranchCollapsed(name string) {
	self.tree.ToggleCollapsed(name)
}

func (self *BranchesContext) CollapseAllBranches() {
	self.tree.CollapseAll(self.c.Model().Branches)
}

func (self *BranchesContext) ExpandAllBranches() {
	self.tree.ExpandAll()
}

// GetBranchNode returns the tree node for the named branch, if it is currently
// visible. The controller uses its depth and HasChildren for click handling.
func (self *BranchesContext) GetBranchNode(name string) (branchtree.Node, bool) {
	return self.tree.Node(self.c.Model().Branches, name)
}

// ExpandAncestorsOf expands any collapsed ancestors of the named branch so that
// the branch itself becomes visible.
func (self *BranchesContext) ExpandAncestorsOf(name string) {
	for _, ancestor := range self.tree.AncestorNames(self.c.Model().Branches, name) {
		self.tree.Expand(ancestor)
	}
}

// SelectBranchByNameOrAncestor selects the branch with the given name. If that
// branch is hidden under a collapsed ancestor, it selects the nearest visible
// ancestor instead. It does nothing when no candidate is visible.
func (self *BranchesContext) SelectBranchByNameOrAncestor(name string) {
	candidates := append([]string{name}, self.tree.AncestorNames(self.c.Model().Branches, name)...)
	items := self.GetItems()
	for _, candidate := range candidates {
		if _, idx, found := lo.FindIndexOf(items, func(b *models.Branch) bool {
			return b.Name == candidate
		}); found {
			self.SetSelectedLineIdx(idx)
			return
		}
	}
}

func (self *BranchesContext) GetSelectedRef() models.Ref {
	branch := self.GetSelected()
	if branch == nil {
		return nil
	}
	return branch
}

func (self *BranchesContext) GetDiffTerminals() []string {
	// for our local branches we want to include both the branch and its upstream
	branch := self.GetSelected()
	if branch != nil {
		names := []string{branch.ID()}
		if branch.IsTrackingRemote() {
			names = append(names, branch.ID()+"@{u}")
		}
		return names
	}
	return nil
}

func (self *BranchesContext) RefForAdjustingLineNumberInDiff() string {
	branch := self.GetSelected()
	if branch != nil {
		return branch.ID()
	}
	return ""
}

func (self *BranchesContext) ShowBranchHeadsInSubCommits() bool {
	return true
}
