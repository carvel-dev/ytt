// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Source interface {
	Description() string
	RelativePath() (string, error)
	Bytes() ([]byte, error)
}

var _ []Source = []Source{BytesSource{}, StdinSource{},
	LocalSource{}, HTTPSource{}, &CachedSource{}}

type BytesSource struct {
	path string
	data []byte
}

func NewBytesSource(path string, data []byte) BytesSource { return BytesSource{path, data} }

func (s BytesSource) Description() string           { return s.path }
func (s BytesSource) RelativePath() (string, error) { return s.path, nil }
func (s BytesSource) Bytes() ([]byte, error)        { return s.data, nil }

type StdinSource struct {
	bytes []byte
	err   error
}

func NewStdinSource() StdinSource {
	// only read stdin once
	bs, err := ReadStdin()
	return StdinSource{bs, err}
}

func (s StdinSource) Description() string           { return "stdin.yml" }
func (s StdinSource) RelativePath() (string, error) { return "stdin.yml", nil }
func (s StdinSource) Bytes() ([]byte, error)        { return s.bytes, s.err }

type LocalSource struct {
	path string
	dir  string
}

func NewLocalSource(path, dir string) LocalSource { return LocalSource{path, dir} }

func (s LocalSource) Description() string { return fmt.Sprintf("file '%s'", s.path) }

func (s LocalSource) RelativePath() (string, error) {
	if s.dir == "" {
		return filepath.Base(s.path), nil
	}

	cleanPath, err := filepath.Abs(filepath.Clean(s.path))
	if err != nil {
		return "", err
	}

	cleanDir, err := filepath.Abs(filepath.Clean(s.dir))
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(cleanPath, cleanDir) {
		result := strings.TrimPrefix(cleanPath, cleanDir)
		result = strings.TrimPrefix(result, string(os.PathSeparator))
		return result, nil
	}

	return "", fmt.Errorf("unknown relative path for %s", s.path)
}

// Bytes returns  bytes of the read file
func (s LocalSource) Bytes() ([]byte, error) { return os.ReadFile(s.path) }

type HTTPSource struct {
	url    string
	Client *http.Client
}

// NewHTTPSource returns a new source of type HTTP
func NewHTTPSource(path string) HTTPSource { return HTTPSource{path, &http.Client{}} }

func (s HTTPSource) Description() string {
	return fmt.Sprintf("HTTP URL '%s'", s.url)
}

func (s HTTPSource) RelativePath() (string, error) { return path.Base(s.url), nil }

func (s HTTPSource) Bytes() ([]byte, error) {
	resp, err := s.Client.Get(s.url)
	if err != nil {
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("Requesting URL '%s': %s", s.url, resp.Status)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading URL '%s': %s", s.url, err)
	}

	return result, nil
}

type CachedSource struct {
	src Source

	bytesFetched bool
	bytes        []byte
	bytesErr     error
}

func NewCachedSource(src Source) *CachedSource { return &CachedSource{src: src} }

func (s *CachedSource) Description() string           { return s.src.Description() }
func (s *CachedSource) RelativePath() (string, error) { return s.src.RelativePath() }

func (s *CachedSource) Bytes() ([]byte, error) {
	if s.bytesFetched {
		return s.bytes, s.bytesErr
	}

	s.bytesFetched = true
	s.bytes, s.bytesErr = s.src.Bytes()

	return s.bytes, s.bytesErr
}
