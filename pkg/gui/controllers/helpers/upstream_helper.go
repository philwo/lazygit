package helpers

import (
	"errors"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type UpstreamHelper struct {
	c *HelperCommon

	getUpstreamSuggestionsFunc func(string) func(string) []*types.Suggestion
}

func NewUpstreamHelper(
	c *HelperCommon,
	getUpstreamSuggestionsFunc func(string) func(string) []*types.Suggestion,
) *UpstreamHelper {
	return &UpstreamHelper{
		c:                          c,
		getUpstreamSuggestionsFunc: getUpstreamSuggestionsFunc,
	}
}

func (self *UpstreamHelper) ParseUpstream(upstream string) (string, string, error) {
	split := strings.Split(upstream, " ")
	if len(split) == 1 && split[0] != "" {
		return ".", split[0], nil
	}
	if len(split) != 2 {
		return "", "", errors.New(self.c.Tr.InvalidUpstream)
	}

	return split[0], split[1], nil
}

func (self *UpstreamHelper) promptForUpstream(currentBranch *models.Branch, initialContent string, onConfirm func(string) error) error {
	self.c.Prompt(types.PromptOpts{
		Title:               self.c.Tr.EnterUpstream,
		InitialContent:      initialContent,
		FindSuggestionsFunc: self.getUpstreamSuggestionsFunc(currentBranch.Name),
		HandleConfirm:       onConfirm,
	})

	return nil
}

func (self *UpstreamHelper) PromptForUpstreamWithInitialContent(currentBranch *models.Branch, onConfirm func(string) error) error {
	suggestedRemote := self.GetSuggestedRemote()
	initialContent := suggestedRemote + " " + currentBranch.Name

	return self.promptForUpstream(currentBranch, initialContent, onConfirm)
}

func (self *UpstreamHelper) PromptForUpstreamWithoutInitialContent(currentBranch *models.Branch, onConfirm func(string) error) error {
	return self.promptForUpstream(currentBranch, "", onConfirm)
}

func (self *UpstreamHelper) GetSuggestedRemote() string {
	return getSuggestedRemote(self.c.Model().Remotes)
}

func getSuggestedRemote(remotes []*models.Remote) string {
	if len(remotes) == 0 {
		return "origin"
	}

	for _, remote := range remotes {
		if remote.Name == "origin" {
			return remote.Name
		}
	}

	return remotes[0].Name
}
