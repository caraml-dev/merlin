// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlab

import (
	"github.com/xanzy/go-gitlab"
)

// Client for GitLab interface.
type Client interface {
	GetFileContent(opt GetFileContentOptions) (string, error)
	CreateFile(opt CreateFileOptions) error
	UpdateFile(opt UpdateFileOptions) error
}

type client struct {
	git *gitlab.Client
}

// NewClient initializes new GitLab client.
func NewClient(baseURL, token string) (Client, error) {
	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}

	return &client{
		git: git,
	}, nil
}

type GetFileContentOptions struct {
	Repository string
	Branch     string
	FileName   string
}

func (c *client) GetFileContent(opt GetFileContentOptions) (string, error) {
	getFile := &gitlab.GetFileOptions{
		Ref: &opt.Branch,
	}

	f, _, err := c.git.RepositoryFiles.GetFile(opt.Repository, opt.FileName, getFile)
	if err != nil {
		return "", err
	}

	return f.Content, nil
}

type CreateFileOptions struct {
	Repository    string
	Branch        string
	FileName      string
	Content       string
	CommitMessage string
	AuthorEmail   string
	AuthorName    string
}

func (c *client) CreateFile(opt CreateFileOptions) error {
	createFile := &gitlab.CreateFileOptions{
		Branch:        &opt.Branch,
		Content:       &opt.Content,
		CommitMessage: &opt.CommitMessage,
		AuthorEmail:   &opt.AuthorEmail,
		AuthorName:    &opt.AuthorName,
	}

	_, _, err := c.git.RepositoryFiles.CreateFile(opt.Repository, opt.FileName, createFile)
	return err
}

type UpdateFileOptions struct {
	Repository    string
	Branch        string
	FileName      string
	Content       string
	CommitMessage string
	AuthorEmail   string
	AuthorName    string
}

func (c *client) UpdateFile(opt UpdateFileOptions) error {
	updateFile := &gitlab.UpdateFileOptions{
		Branch:        &opt.Branch,
		Content:       &opt.Content,
		CommitMessage: &opt.CommitMessage,
		AuthorEmail:   &opt.AuthorEmail,
		AuthorName:    &opt.AuthorName,
	}

	_, _, err := c.git.RepositoryFiles.UpdateFile(opt.Repository, opt.FileName, updateFile)
	return err
}
