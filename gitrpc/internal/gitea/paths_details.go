// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package gitea

import (
	"context"
	"errors"
	"fmt"

	"github.com/harness/gitness/gitrpc/internal/types"

	gogitplumbing "github.com/go-git/go-git/v5/plumbing"
	gogitfilemode "github.com/go-git/go-git/v5/plumbing/filemode"
	gogitobject "github.com/go-git/go-git/v5/plumbing/object"
)

// PathsDetails returns additional details about provided the paths.
func (g Adapter) PathsDetails(ctx context.Context,
	repoPath string,
	ref string,
	paths []string,
) ([]types.PathDetails, error) {
	repo, refCommit, err := g.getGoGitCommit(ctx, repoPath, ref)
	if err != nil {
		return nil, err
	}

	refSHA := refCommit.Hash.String()

	tree, err := refCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for the commit: %w", err)
	}

	results := make([]types.PathDetails, len(paths))

	for i, path := range paths {
		results[i].Path = path

		if len(path) > 0 {
			entry, err := tree.FindEntry(path)
			if errors.Is(err, gogitobject.ErrDirectoryNotFound) || errors.Is(err, gogitobject.ErrEntryNotFound) {
				return nil, types.ErrPathNotFound
			} else if err != nil {
				return nil, fmt.Errorf("can't find path entry %s: %w", path, err)
			}

			if entry.Mode == gogitfilemode.Regular || entry.Mode == gogitfilemode.Executable {
				blobObj, err := repo.Object(gogitplumbing.BlobObject, entry.Hash)
				if err != nil {
					return nil, fmt.Errorf("failed to get blob object size for the path %s and hash %s: %w",
						path, entry.Hash.String(), err)
				}

				results[i].Size = blobObj.(*gogitobject.Blob).Size
			}
		}

		commitEntry, err := g.lastCommitCache.Get(ctx, makeCommitEntryKey(repoPath, refSHA, path))
		if err != nil {
			return nil, fmt.Errorf("failed to find last commit for path %s: %w", path, err)
		}

		results[i].LastCommit = commitEntry
	}

	return results, nil
}
