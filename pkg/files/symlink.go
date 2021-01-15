// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Symlink struct {
	path string
}

type SymlinkAllowOpts struct {
	AllowAll        bool
	AllowedDstPaths []string
}

var (
	symlinkPipeErrMsg = regexp.QuoteMeta("lstat /proc/NUM/fd/pipe:[NUM]: no such file or directory")
	symlinkPipeErr    = regexp.MustCompile("^" + strings.Replace(symlinkPipeErrMsg, "NUM", "\\d+", -1) + "$")
)

func (s Symlink) IsAllowed(opts SymlinkAllowOpts) error {
	if opts.AllowAll {
		return nil
	}

	dstPath, err := filepath.EvalSymlinks(s.path)
	if err != nil {
		// Note that on Linux resolving symlink /dev/fd/3 fails:
		// "lstat /proc/3719/fd/pipe:[903476724]: no such file or directory"
		// Since file doesnt actually exist on FS, it could not have been tricked to be included.
		if symlinkPipeErr.MatchString(err.Error()) {
			return nil
		}
		return fmt.Errorf("Eval symlink: %s", err)
	}

	for _, allowedDstPath := range opts.AllowedDstPaths {
		matched, err := s.isIn(dstPath, allowedDstPath)
		if matched || err != nil {
			return err
		}
	}

	return fmt.Errorf("Expected symlink file '%s' -> '%s' to be allowed, but was not", s.path, dstPath)
}

func (s Symlink) isIn(path, allowedPath string) (bool, error) {
	var err error

	// Abs runs clean on the result
	path, err = filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("Abs path '%s': %s", path, err)
	}

	allowedPath, err = filepath.Abs(allowedPath)
	if err != nil {
		return false, fmt.Errorf("Abs path '%s': %s", allowedPath, err)
	}

	pathPieces := s.pathPieces(path)
	allowedPathPieces := s.pathPieces(allowedPath)

	if len(allowedPathPieces) > len(pathPieces) {
		return false, nil
	}

	for i := range allowedPathPieces {
		if allowedPathPieces[i] != pathPieces[i] {
			return false, nil
		}
	}

	return true, nil
}

func (s Symlink) pathPieces(path string) []string {
	if path == string(filepath.Separator) {
		return []string{""}
	}
	return strings.Split(path, string(filepath.Separator))
}
