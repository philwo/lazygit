// Package branchtree arranges a flat list of local branches into a forest of
// stacked branches, where a branch whose upstream is another local branch is
// nested under that parent. It computes the visible, tree-ordered subset (with
// collapse/expand applied) and the ASCII connector prefixes for rendering. It
// is deliberately GUI-free so it can be unit tested in isolation.
package branchtree

import (
	"cmp"
	"slices"
	"strings"

	"github.com/jesseduffield/generics/set"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
)

// Node is one branch in the arranged tree. Every node carries a real,
// selectable branch: roots, internal nodes, and leaves alike.
type Node struct {
	Branch      *models.Branch
	Depth       int    // 0 for roots
	Prefix      string // connector glyphs including trailing space; "" for roots
	HasChildren bool
	Collapsed   bool
}

// BranchTree owns the tree/flat toggle and the set of collapsed branch names.
// Everything is keyed by branch name, never by path, because branch names can
// contain slashes.
type BranchTree struct {
	showTree  bool
	collapsed *set.Set[string]
}

func New(showTree bool) *BranchTree {
	return &BranchTree{
		showTree:  showTree,
		collapsed: set.New[string](),
	}
}

func (self *BranchTree) InTreeMode() bool {
	return self.showTree
}

func (self *BranchTree) SetShowTree(showTree bool) {
	self.showTree = showTree
}

func (self *BranchTree) ToggleShowTree() {
	self.showTree = !self.showTree
}

func (self *BranchTree) IsCollapsed(name string) bool {
	return self.collapsed.Includes(name)
}

func (self *BranchTree) ToggleCollapsed(name string) {
	if self.collapsed.Includes(name) {
		self.collapsed.Remove(name)
	} else {
		self.collapsed.Add(name)
	}
}

func (self *BranchTree) Collapse(name string) {
	self.collapsed.Add(name)
}

func (self *BranchTree) Expand(name string) {
	self.collapsed.Remove(name)
}

func (self *BranchTree) ExpandAll() {
	self.collapsed.RemoveSlice(self.collapsed.ToSlice())
}

// CollapseAll collapses every branch that has at least one child, except the
// ancestors of the checked-out branch: the path down to it stays expanded so it
// remains visible. Collapsing a childless branch would be a no-op that only
// clutters the collapsed set.
func (self *BranchTree) CollapseAll(branches []*models.Branch) {
	parent := resolveParents(branches)

	keepExpanded := set.New[string]()
	for _, branch := range branches {
		if branch.Head {
			keepExpanded.Add(ancestorNames(parent, branch.Name)...)
			break
		}
	}

	for _, p := range parent {
		if !keepExpanded.Includes(p) {
			self.collapsed.Add(p)
		}
	}
}

// Build is the single source of truth: it returns the visible, tree-ordered
// nodes together with their connector prefixes, so the visible list and the
// prefixes can never drift. In flat mode it returns the input unchanged, at
// depth 0 with empty prefixes.
func (self *BranchTree) Build(branches []*models.Branch) []Node {
	if !self.showTree {
		nodes := make([]Node, len(branches))
		for i, branch := range branches {
			nodes[i] = Node{Branch: branch}
		}
		return nodes
	}

	// Walk the branches in sort order rather than model order: the model hoists
	// the checked-out branch to the front, which would otherwise move it ahead
	// of its siblings every time you check something out.
	sorted := slices.Clone(branches)
	slices.SortStableFunc(sorted, func(a *models.Branch, b *models.Branch) int {
		return cmp.Compare(a.SortIndex, b.SortIndex)
	})

	parent := resolveParents(sorted)

	children := make(map[string][]*models.Branch)
	roots := make([]*models.Branch, 0, len(sorted))
	for _, branch := range sorted {
		if p := parent[branch.Name]; p != "" {
			children[p] = append(children[p], branch)
		} else {
			roots = append(roots, branch)
		}
	}

	nodes := make([]Node, 0, len(branches))
	var dfs func(branch *models.Branch, depth int, ancestorIsLast []bool)
	dfs = func(branch *models.Branch, depth int, ancestorIsLast []bool) {
		kids := children[branch.Name]
		collapsed := self.collapsed.Includes(branch.Name)
		nodes = append(nodes, Node{
			Branch:      branch,
			Depth:       depth,
			Prefix:      buildPrefix(ancestorIsLast),
			HasChildren: len(kids) > 0,
			Collapsed:   collapsed,
		})
		if collapsed {
			return
		}
		for i, kid := range kids {
			// Copy into a fresh slice; sharing the backing array across
			// siblings would corrupt earlier siblings' connectors.
			childAncestors := make([]bool, depth+1)
			copy(childAncestors, ancestorIsLast)
			childAncestors[depth] = i == len(kids)-1
			dfs(kid, depth+1, childAncestors)
		}
	}
	for _, root := range roots {
		dfs(root, 0, nil)
	}

	return nodes
}

// VisibleBranches returns just the branches from Build, in visible tree order.
func (self *BranchTree) VisibleBranches(branches []*models.Branch) []*models.Branch {
	if !self.showTree {
		return branches
	}

	nodes := self.Build(branches)
	result := make([]*models.Branch, len(nodes))
	for i, node := range nodes {
		result[i] = node.Branch
	}
	return result
}

// Node returns the built node for the named branch, if it is currently visible.
func (self *BranchTree) Node(branches []*models.Branch, name string) (Node, bool) {
	for _, node := range self.Build(branches) {
		if node.Branch.Name == name {
			return node, true
		}
	}
	return Node{}, false
}

// AncestorNames returns the parent chain of the named branch, nearest ancestor
// first. Callers use it to expand collapsed ancestors so a branch becomes
// visible, or to find the nearest visible ancestor to select when a branch got
// hidden.
func (self *BranchTree) AncestorNames(branches []*models.Branch, name string) []string {
	return ancestorNames(resolveParents(branches), name)
}

// ancestorNames walks a precomputed parent map from name up to the root,
// returning the ancestors nearest first. The bounded seen set guards against a
// cycle that resolveParents did not break.
func ancestorNames(parent map[string]string, name string) []string {
	result := []string{}
	seen := set.New[string]()
	seen.Add(name)
	for cur := name; ; {
		p := parent[cur]
		if p == "" || seen.Includes(p) {
			break
		}
		result = append(result, p)
		seen.Add(p)
		cur = p
	}
	return result
}

// resolveParents computes the parent branch name for each branch, or "" for a
// root. A branch's parent is its local upstream: UpstreamRemote == "." and
// UpstreamBranch names a different, existing local branch. Cycles are broken
// deterministically so the result is always a forest.
func resolveParents(branches []*models.Branch) map[string]string {
	byName := make(map[string]*models.Branch, len(branches))
	for _, branch := range branches {
		byName[branch.Name] = branch
	}

	parent := make(map[string]string, len(branches))
	for _, branch := range branches {
		if branch.UpstreamRemote != "." || branch.UpstreamBranch == "" || branch.UpstreamBranch == branch.Name {
			continue
		}
		if _, ok := byName[branch.UpstreamBranch]; ok {
			parent[branch.Name] = branch.UpstreamBranch
		}
	}

	// Break cycles: walk up from each node; if the walk returns to the node
	// itself before reaching a root, drop that node's parent edge (promote it
	// to a root). Processing in input order with a bounded visited set makes
	// this deterministic.
	for _, branch := range branches {
		name := branch.Name
		if parent[name] == "" {
			continue
		}
		visited := set.New[string]()
		visited.Add(name)
		for cur := name; ; {
			p := parent[cur]
			if p == "" {
				break // reached a root: edge is valid
			}
			if p == name {
				delete(parent, name) // cycle back to the start: drop our edge
				break
			}
			if visited.Includes(p) {
				break // cycle downstream of us: keep our edge, it gets broken there
			}
			visited.Add(p)
			cur = p
		}
	}

	return parent
}

// buildPrefix turns the ancestor "is-last-child" flags into connector glyphs.
// Each level is a uniform 3 display columns, so the prefix width is 3*depth.
func buildPrefix(ancestorIsLast []bool) string {
	if len(ancestorIsLast) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, isLast := range ancestorIsLast[:len(ancestorIsLast)-1] {
		if isLast {
			sb.WriteString("   ")
		} else {
			sb.WriteString("│  ")
		}
	}
	if ancestorIsLast[len(ancestorIsLast)-1] {
		sb.WriteString("└─ ")
	} else {
		sb.WriteString("├─ ")
	}
	return sb.String()
}
