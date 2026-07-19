package patch_building

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var DiscardLinesFromCommitWithActivePatch = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Discard lines from a commit with the 'd' shortcut while a custom patch is already active",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "1st line\n2nd line\n3rd line\n")
		shell.Commit("initial commit")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().
			Focus().
			Lines(
				Contains("initial commit").IsSelected(),
			).
			PressEnter()

		t.Views().CommitFiles().
			IsFocused().
			Lines(
				Contains("A file1").IsSelected(),
			).
			PressEnter()

		// Toggle the second line into the patch, then discard it with 'd'
		t.Views().PatchBuilding().
			IsFocused().
			SelectNextItem().
			SelectedLines(
				Contains("+2nd line"),
			).
			PressPrimaryAction().
			SelectPreviousItem().
			SelectedLines(
				Contains("+2nd line"),
			).
			Press(keys.Universal.Remove)

		t.ExpectPopup().Confirmation().
			Title(Equals("Discard lines from commit")).
			Content(Contains("Are you sure you want to discard the selected lines from this commit?")).
			Confirm()

		/* EXPECTED:
		t.Views().CommitFiles().
			IsFocused().
			Lines(
				Contains("A file1").IsSelected(),
			).
			PressEscape()

		t.Views().Main().ContainsLines(
			Equals("+1st line"),
			Equals("+3rd line"),
		)
		ACTUAL: */
		t.ExpectPopup().Alert().
			Title(Equals("Error")).
			Content(Contains("bad revision")).
			Confirm()
	},
})
