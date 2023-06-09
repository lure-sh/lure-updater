/*
 * LURE Updater - Automated updater bot for LURE packages
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package builtins

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.elara.ws/logger/log"
	"go.elara.ws/lure-updater/internal/config"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func updaterModule(cfg *config.Config) *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "updater",
		Members: starlark.StringDict{
			"repo_dir":           starlark.String(cfg.Git.RepoDir),
			"pull":               updaterPull(cfg),
			"push_changes":       updaterPushChanges(cfg),
			"get_package_file":   getPackageFile(cfg),
			"write_package_file": writePackageFile(cfg),
		},
	}
}

// repoMtx makes sure two starlark threads can
// never access the repo at the same time
var repoMtx = &sync.Mutex{}

func updaterPull(cfg *config.Config) *starlark.Builtin {
	return starlark.NewBuiltin("updater.pull", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		repoMtx.Lock()
		defer repoMtx.Unlock()

		repo, err := git.PlainOpen(cfg.Git.RepoDir)
		if err != nil {
			return nil, err
		}

		w, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		err = w.Pull(&git.PullOptions{Progress: os.Stderr})
		if err != git.NoErrAlreadyUpToDate && err != nil {
			return nil, err
		}

		return starlark.None, nil
	})
}

func updaterPushChanges(cfg *config.Config) *starlark.Builtin {
	return starlark.NewBuiltin("updater.push_changes", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var msg string
		err := starlark.UnpackArgs("updater.push_changes", args, kwargs, "msg", &msg)
		if err != nil {
			return nil, err
		}

		repoMtx.Lock()
		defer repoMtx.Unlock()

		repo, err := git.PlainOpen(cfg.Git.RepoDir)
		if err != nil {
			return nil, err
		}

		w, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		status, err := w.Status()
		if err != nil {
			return nil, err
		}

		if status.IsClean() {
			return starlark.None, nil
		}

		err = w.Pull(&git.PullOptions{Progress: os.Stderr})
		if err != git.NoErrAlreadyUpToDate && err != nil {
			return nil, err
		}

		_, err = w.Add(".")
		if err != nil {
			return nil, err
		}

		sig := &object.Signature{
			Name:  cfg.Git.Commit.Name,
			Email: cfg.Git.Commit.Email,
			When:  time.Now(),
		}

		h, err := w.Commit(msg, &git.CommitOptions{
			Author:    sig,
			Committer: sig,
		})
		if err != nil {
			return nil, err
		}

		log.Debug("Created new commit").Stringer("hash", h).Stringer("pos", thread.CallFrame(1).Pos).Send()

		err = repo.Push(&git.PushOptions{
			Progress: os.Stderr,
			Auth: &http.BasicAuth{
				Username: cfg.Git.Credentials.Username,
				Password: cfg.Git.Credentials.Password,
			},
		})
		if err != nil {
			return nil, err
		}

		log.Debug("Successfully pushed to repo").Stringer("pos", thread.CallFrame(1).Pos).Send()

		return starlark.None, nil
	})
}

func getPackageFile(cfg *config.Config) *starlark.Builtin {
	return starlark.NewBuiltin("updater.get_package_file", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var pkg, filename string
		err := starlark.UnpackArgs("updater.get_package_file", args, kwargs, "pkg", &pkg, "filename", &filename)
		if err != nil {
			return nil, err
		}

		repoMtx.Lock()
		defer repoMtx.Unlock()

		path := filepath.Join(cfg.Git.RepoDir, pkg, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		log.Debug("Got package file").Str("package", pkg).Str("filename", filename).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.String(data), nil
	})
}

func writePackageFile(cfg *config.Config) *starlark.Builtin {
	return starlark.NewBuiltin("updater.write_package_file", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var pkg, filename, content string
		err := starlark.UnpackArgs("updater.write_package_file", args, kwargs, "pkg", &pkg, "filename", &filename, "content", &content)
		if err != nil {
			return nil, err
		}

		repoMtx.Lock()
		defer repoMtx.Unlock()

		path := filepath.Join(cfg.Git.RepoDir, pkg, filename)
		fl, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fl, strings.NewReader(content))
		if err != nil {
			return nil, err
		}

		log.Debug("Wrote package file").Str("package", pkg).Str("filename", filename).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.None, nil
	})
}
