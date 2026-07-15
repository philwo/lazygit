package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBranch_UpstreamRefNames(t *testing.T) {
	t.Run("remote upstream", func(t *testing.T) {
		b := &Branch{UpstreamRemote: "origin", UpstreamBranch: "main"}
		assert.Equal(t, "origin/main", b.ShortUpstreamRefName())
		assert.Equal(t, "refs/remotes/origin/main", b.FullUpstreamRefName())
	})

	t.Run("local upstream", func(t *testing.T) {
		b := &Branch{UpstreamRemote: ".", UpstreamBranch: "main"}
		/* EXPECTED:
		assert.Equal(t, "main", b.ShortUpstreamRefName())
		assert.Equal(t, "refs/heads/main", b.FullUpstreamRefName())
		ACTUAL: */
		assert.Equal(t, "./main", b.ShortUpstreamRefName())
		assert.Equal(t, "refs/remotes/./main", b.FullUpstreamRefName())
	})
}
