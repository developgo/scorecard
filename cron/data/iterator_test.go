// Copyright 2021 Security Scorecard Authors
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

package data

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v3/repos"
)

type outcome struct {
	expectedErr error
	repo        repos.RepoURL
	hasError    bool
}

// nolint: gocognit
func TestCsvIterator(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		filename string
		outcomes []outcome
	}{
		{
			name:     "Basic",
			filename: "testdata/basic.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner1",
						Repo:  "repo1",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner2",
						Repo:  "repo2",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:     "github.com",
						Owner:    "owner3",
						Repo:     "repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:     "Comment",
			filename: "testdata/comment.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner1",
						Repo:  "repo1",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner2",
						Repo:  "repo2",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:     "github.com",
						Owner:    "owner3",
						Repo:     "repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:     "FailingURLs",
			filename: "testdata/failing_urls.csv",
			outcomes: []outcome{
				{
					hasError:    true,
					expectedErr: repos.ErrorUnsupportedHost,
				},
				{
					hasError:    true,
					expectedErr: repos.ErrorInvalidURL,
				},
				{
					hasError:    true,
					expectedErr: repos.ErrorInvalidURL,
				},
			},
		},
		{
			name:     "EmptyRows",
			filename: "testdata/empty_row.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner1",
						Repo:  "repo1",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner2",
						Repo:  "repo2",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:     "github.com",
						Owner:    "owner3",
						Repo:     "repo3",
						Metadata: []string{"meta"},
					},
				},
			},
		},
		{
			name:     "ExtraColumns",
			filename: "testdata/extra_column.csv",
			outcomes: []outcome{
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner1",
						Repo:  "repo1",
					},
				},
				{
					hasError: false,
					repo: repos.RepoURL{
						Host:  "github.com",
						Owner: "owner2",
						Repo:  "repo2",
					},
				},
				{
					hasError: true,
				},
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			testFile, err := os.OpenFile(testcase.filename, os.O_RDONLY, 0o644)
			if err != nil {
				t.Errorf("failed to open %s: %v", testcase.filename, err)
			}
			defer testFile.Close()

			testReader, err := MakeIteratorFrom(testFile)
			if err != nil {
				t.Errorf("failed to create reader: %v", err)
			}
			for _, outcome := range testcase.outcomes {
				if !testReader.HasNext() {
					t.Error("expected outcome, got EOF")
				}
				repoURL, err := testReader.Next()
				if (err != nil) != outcome.hasError {
					t.Errorf("expected hasError: %t, got: %v", outcome.hasError, err)
				}
				if !outcome.hasError && !cmp.Equal(outcome.repo, repoURL) {
					t.Errorf("expected repoURL: %s, got %s", outcome.repo, repoURL)
				}
				if outcome.hasError && outcome.expectedErr != nil && !errors.Is(err, outcome.expectedErr) {
					t.Errorf("expected error: %v, got %v", outcome.expectedErr, err)
				}
			}
			if testReader.HasNext() {
				t.Error("actual reader has more repos than expected")
			}
		})
	}
}
