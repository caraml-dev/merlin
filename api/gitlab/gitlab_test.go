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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_client(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/":
			fmt.Fprintf(w, "")
		case "/api/v4/projects/test/repository/files/README.md":
			fmt.Fprintf(w, `{
				"file_name": "README.md",
				"file_path": "README.md",
				"content": "README"
			}`)
		case "/api/v4/projects/test/repository/files/.gitignore":
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				fmt.Fprintf(w, `{
					"file_path": ".gitignore",
					"branch": "master"
				}`)
			} else if r.Method == http.MethodDelete {
				fmt.Fprintf(w, "")
			} else {
				fmt.Printf("Unregistered URL method %s for %s\n", r.Method, r.URL.Path)
			}
		default:
			fmt.Printf("Unregistered URL path %s\n", r.URL.Path)
		}
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "")
	assert.Nil(t, err)
	assert.NotNil(t, client)

	getOpt := GetFileContentOptions{
		Repository: "test",
		Branch:     "master",
		FileName:   "README.md",
	}
	content, err := client.GetFileContent(getOpt)
	assert.Nil(t, err)
	assert.NotEmpty(t, content)

	createOpt := CreateFileOptions{
		Repository:    "test",
		Branch:        "master",
		FileName:      ".gitignore",
		Content:       `*.exe`,
		CommitMessage: "Create",
		AuthorEmail:   "merlin-dev@gojek.com",
		AuthorName:    "merlin-dev@gojek.com",
	}
	err = client.CreateFile(createOpt)
	assert.Nil(t, err)

	updateOpt := UpdateFileOptions{
		Repository:    "test",
		Branch:        "master",
		FileName:      ".gitignore",
		Content:       `*.out`,
		CommitMessage: "Update",
		AuthorEmail:   "merlin-dev@gojek.com",
		AuthorName:    "merlin-dev@gojek.com",
	}
	err = client.UpdateFile(updateOpt)
	assert.Nil(t, err)
}
