package branchtree

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// branch builds a local branch with a local upstream (remote ".") pointing at
// parent. An empty parent means the branch has no upstream (a root).
func branch(name string, parent string) *models.Branch {
	b := &models.Branch{Name: name}
	if parent != "" {
		b.UpstreamRemote = "."
		b.UpstreamBranch = parent
	}
	return b
}

// atSortIndex sets the branch's position in the loader's sort order, which the
// loader records before it moves the checked-out branch to the front.
func atSortIndex(b *models.Branch, index int) *models.Branch {
	b.SortIndex = index
	return b
}

// remoteBranch builds a branch tracking a real remote, which must never be
// treated as a stacked child.
func remoteBranch(name string, remote string, upstream string) *models.Branch {
	return &models.Branch{Name: name, UpstreamRemote: remote, UpstreamBranch: upstream}
}

type expectedNode struct {
	name   string
	depth  int
	prefix string
}

func toExpected(nodes []Node) []expectedNode {
	return lo.Map(nodes, func(node Node, _ int) expectedNode {
		return expectedNode{name: node.Branch.Name, depth: node.Depth, prefix: node.Prefix}
	})
}

func TestBuild(t *testing.T) {
	scenarios := []struct {
		name      string
		showTree  bool
		collapsed []string
		branches  []*models.Branch
		expected  []expectedNode
	}{
		{
			name:     "flat mode returns input order with no prefixes",
			showTree: false,
			branches: []*models.Branch{
				branch("main", ""),
				branch("feat1", "main"),
				branch("feat1b", "feat1"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 0, ""},
				{"feat1b", 0, ""},
			},
		},
		{
			name:     "linear stack main->feat1->feat1b",
			showTree: true,
			branches: []*models.Branch{
				branch("main", ""),
				branch("feat1", "main"),
				branch("feat1b", "feat1"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 1, "└─ "},
				{"feat1b", 2, "   └─ "},
			},
		},
		{
			name:     "product example: siblings, continuation, and multiple roots",
			showTree: true,
			branches: []*models.Branch{
				branch("main", ""),
				branch("feat1", "main"),
				branch("feat1b", "feat1"),
				branch("experiment", "main"),
				branch("develop", ""),
				branch("hotfix", "develop"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 1, "├─ "},
				{"feat1b", 2, "│  └─ "},
				{"experiment", 1, "└─ "},
				{"develop", 0, ""},
				{"hotfix", 1, "└─ "},
			},
		},
		{
			name:     "sibling under a last child uses space filler",
			showTree: true,
			branches: []*models.Branch{
				branch("main", ""),
				branch("a", "main"),
				branch("a1", "a"),
				branch("a2", "a"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"a", 1, "└─ "},
				{"a1", 2, "   ├─ "},
				{"a2", 2, "   └─ "},
			},
		},
		{
			name:     "sibling under a non-last child uses vertical continuation",
			showTree: true,
			branches: []*models.Branch{
				branch("main", ""),
				branch("a", "main"),
				branch("a1", "a"),
				branch("b", "main"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"a", 1, "├─ "},
				{"a1", 2, "│  └─ "},
				{"b", 1, "└─ "},
			},
		},
		{
			name:     "missing parent becomes a root",
			showTree: true,
			branches: []*models.Branch{
				branch("feat1", "main"),
				branch("feat1b", "feat1"),
			},
			expected: []expectedNode{
				{"feat1", 0, ""},
				{"feat1b", 1, "└─ "},
			},
		},
		{
			name:     "remote upstream is never a parent",
			showTree: true,
			branches: []*models.Branch{
				branch("main", ""),
				remoteBranch("feat1", "origin", "main"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 0, ""},
			},
		},
		{
			name:     "self-cycle becomes a root",
			showTree: true,
			branches: []*models.Branch{
				branch("main", ""),
				branch("self", "self"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"self", 0, ""},
			},
		},
		{
			name:     "two-cycle resolves to a deterministic forest",
			showTree: true,
			branches: []*models.Branch{
				branch("a", "b"),
				branch("b", "a"),
			},
			// a is processed first, so a's edge is dropped and a becomes the
			// root, with b nested beneath it.
			expected: []expectedNode{
				{"a", 0, ""},
				{"b", 1, "└─ "},
			},
		},
		{
			name:      "collapse hides descendants",
			showTree:  true,
			collapsed: []string{"feat1"},
			branches: []*models.Branch{
				branch("main", ""),
				branch("feat1", "main"),
				branch("feat1b", "feat1"),
				branch("experiment", "main"),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 1, "├─ "},
				{"experiment", 1, "└─ "},
			},
		},
		{
			name:     "sibling order follows the sort index, not the model order",
			showTree: true,
			// The loader moved the checked-out branch (feat1a) to the front of
			// the model, but its sort index still puts it last among feat1's
			// children.
			branches: []*models.Branch{
				atSortIndex(branch("feat1a", "feat1"), 3),
				atSortIndex(branch("main", ""), 0),
				atSortIndex(branch("feat1", "main"), 1),
				atSortIndex(branch("feat1b", "feat1"), 2),
				atSortIndex(branch("develop", ""), 4),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 1, "└─ "},
				{"feat1b", 2, "   ├─ "},
				{"feat1a", 2, "   └─ "},
				{"develop", 0, ""},
			},
		},
		{
			name:     "root order follows the sort index too",
			showTree: true,
			branches: []*models.Branch{
				atSortIndex(branch("develop", ""), 2),
				atSortIndex(branch("main", ""), 0),
				atSortIndex(branch("feat1", "main"), 1),
			},
			expected: []expectedNode{
				{"main", 0, ""},
				{"feat1", 1, "└─ "},
				{"develop", 0, ""},
			},
		},
		{
			name:     "flat mode keeps the checked-out branch at the top",
			showTree: false,
			branches: []*models.Branch{
				atSortIndex(branch("feat1a", "feat1"), 3),
				atSortIndex(branch("main", ""), 0),
				atSortIndex(branch("feat1", "main"), 1),
			},
			expected: []expectedNode{
				{"feat1a", 0, ""},
				{"main", 0, ""},
				{"feat1", 0, ""},
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			tree := New(s.showTree)
			for _, name := range s.collapsed {
				tree.Collapse(name)
			}
			assert.Equal(t, s.expected, toExpected(tree.Build(s.branches)))
		})
	}
}

func TestVisibleBranches(t *testing.T) {
	branches := []*models.Branch{
		branch("main", ""),
		branch("feat1", "main"),
		branch("feat1b", "feat1"),
	}

	tree := New(true)
	names := lo.Map(tree.VisibleBranches(branches), func(b *models.Branch, _ int) string { return b.Name })
	assert.Equal(t, []string{"main", "feat1", "feat1b"}, names)

	tree.Collapse("feat1")
	names = lo.Map(tree.VisibleBranches(branches), func(b *models.Branch, _ int) string { return b.Name })
	assert.Equal(t, []string{"main", "feat1"}, names)
}

func TestCollapseOfChildlessBranchIsNoOp(t *testing.T) {
	branches := []*models.Branch{
		branch("main", ""),
		branch("feat1", "main"),
	}

	tree := New(true)
	// feat1 has no children, so collapsing it doesn't hide anything.
	tree.Collapse("feat1")
	names := lo.Map(tree.VisibleBranches(branches), func(b *models.Branch, _ int) string { return b.Name })
	assert.Equal(t, []string{"main", "feat1"}, names)
}

func TestCollapseAllOnlyCollapsesBranchesWithChildren(t *testing.T) {
	branches := []*models.Branch{
		branch("main", ""),
		branch("feat1", "main"),
		branch("feat1b", "feat1"),
		branch("develop", ""),
	}

	tree := New(true)
	tree.CollapseAll(branches)

	assert.True(t, tree.IsCollapsed("main"))
	assert.True(t, tree.IsCollapsed("feat1"))
	assert.False(t, tree.IsCollapsed("feat1b"))  // leaf
	assert.False(t, tree.IsCollapsed("develop")) // childless root

	names := lo.Map(tree.VisibleBranches(branches), func(b *models.Branch, _ int) string { return b.Name })
	assert.Equal(t, []string{"main", "develop"}, names)

	tree.ExpandAll()
	assert.False(t, tree.IsCollapsed("main"))
	assert.False(t, tree.IsCollapsed("feat1"))
	names = lo.Map(tree.VisibleBranches(branches), func(b *models.Branch, _ int) string { return b.Name })
	assert.Equal(t, []string{"main", "feat1", "feat1b", "develop"}, names)
}

func TestCollapseAllKeepsPathToCheckedOutBranch(t *testing.T) {
	feat1b := branch("feat1b", "feat1")
	feat1b.Head = true
	branches := []*models.Branch{
		branch("main", ""),
		branch("feat1", "main"),
		feat1b,
		branch("topic", "main"),
		branch("develop", ""),
		branch("hotfix", "develop"),
	}

	tree := New(true)
	tree.CollapseAll(branches)

	// The path down to the checked-out branch stays expanded so it stays
	// visible; every other sub-stack collapses.
	assert.False(t, tree.IsCollapsed("main"))
	assert.False(t, tree.IsCollapsed("feat1"))
	assert.True(t, tree.IsCollapsed("develop"))

	names := lo.Map(tree.VisibleBranches(branches), func(b *models.Branch, _ int) string { return b.Name })
	assert.Equal(t, []string{"main", "feat1", "feat1b", "topic", "develop"}, names)
}

func TestAncestorNames(t *testing.T) {
	branches := []*models.Branch{
		branch("main", ""),
		branch("feat1", "main"),
		branch("feat1b", "feat1"),
	}

	tree := New(true)
	assert.Equal(t, []string{"feat1", "main"}, tree.AncestorNames(branches, "feat1b"))
	assert.Equal(t, []string{"main"}, tree.AncestorNames(branches, "feat1"))
	assert.Equal(t, []string{}, tree.AncestorNames(branches, "main"))
}
