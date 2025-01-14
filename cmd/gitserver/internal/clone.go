package internal

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// MaybeStartClone checks if a given repository is cloned on disk. If not, it starts
// cloning the repository in the background and returns a NotFound error, if no current
// clone operation is running for that repo yet. If it is already cloning, a NotFound
// error with CloneInProgress: true is returned.
// Note: If disableAutoGitUpdates is set in the site config, no operation is taken and
// a NotFound error is returned.
func (s *Server) MaybeStartClone(ctx context.Context, repo api.RepoName) (notFound *protocol.NotFoundPayload, cloned bool) {
	dir := gitserverfs.RepoDirFromName(s.reposDir, repo)
	if repoCloned(dir) {
		return nil, true
	}

	if conf.Get().DisableAutoGitUpdates {
		s.logger.Debug("not cloning on demand as DisableAutoGitUpdates is set")
		return &protocol.NotFoundPayload{}, false
	}

	cloneProgress, err := s.CloneRepo(ctx, repo, CloneOptions{})
	if err != nil {
		s.logger.Debug("error starting repo clone", log.String("repo", string(repo)), log.Error(err))
		return &protocol.NotFoundPayload{CloneInProgress: false}, false
	}

	return &protocol.NotFoundPayload{
		CloneInProgress: true,
		CloneProgress:   cloneProgress,
	}, false
}
