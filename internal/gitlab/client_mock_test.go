package gitlab

import (
	"fmt"
	"net/http"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// MockClient implements the Client interface for testing purposes.
type MockClient struct {
	MockGetProject          func(projectID interface{}) (*gitlab.Project, *gitlab.Response, error)
	MockGetFile             func(projectID int, filePath string, options *gitlab.GetFileOptions) (*gitlab.File, *gitlab.Response, error)
	MockCreateFile          func(projectID int, filePath string, opt *gitlab.CreateFileOptions) (*gitlab.File, *gitlab.Response, error)
	MockUpdateFile          func(projectID int, filePath string, opt *gitlab.UpdateFileOptions) (*gitlab.File, *gitlab.Response, error)
	MockDeleteFile          func(projectID int, filePath string, opt *gitlab.DeleteFileOptions) (*gitlab.Response, error)
	MockListProjectBranches func(projectID int, opt *gitlab.ListBranchesOptions) ([]*gitlab.Branch, *gitlab.Response, error)
	MockGetBranch           func(projectID int, branch string) (*gitlab.Branch, *gitlab.Response, error)
	MockCreateBranch        func(projectID int, opt *gitlab.CreateBranchOptions) (*gitlab.Branch, *gitlab.Response, error)
	MockDeleteBranch        func(projectID int, branch string) (*gitlab.Response, error)
	MockGetCommit           func(projectID int, sha string) (*gitlab.Commit, *gitlab.Response, error)
	MockListProjectCommits  func(projectID int, opt *gitlab.ListCommitsOptions) ([]*gitlab.Commit, *gitlab.Response, error)
	MockCreateCommit        func(projectID int, opt *gitlab.CreateCommitOptions) (*gitlab.Commit, *gitlab.Response, error)
	MockGetBranchHeadCommitID func(branch string) (string, error)
}

// NewMockClient creates a new MockClient with default (panic) implementations.
func NewMockClient() *MockClient {
	return &MockClient{
		MockGetProject: func(projectID interface{}) (*gitlab.Project, *gitlab.Response, error) {
			panic("GetProject not mocked")
		},
		MockGetFile: func(projectID int, filePath string, options *gitlab.GetFileOptions) (*gitlab.File, *gitlab.Response, error) {
			panic("GetFile not mocked")
		},
		MockCreateFile: func(projectID int, filePath string, opt *gitlab.CreateFileOptions) (*gitlab.File, *gitlab.Response, error) {
			panic("CreateFile not mocked")
		},
		MockUpdateFile: func(projectID int, filePath string, opt *gitlab.UpdateFileOptions) (*gitlab.File, *gitlab.Response, error) {
			panic("UpdateFile not mocked")
		},
		MockDeleteFile: func(projectID int, filePath string, opt *gitlab.DeleteFileOptions) (*gitlab.Response, error) {
			panic("DeleteFile not mocked")
		},
		MockListProjectBranches: func(projectID int, opt *gitlab.ListBranchesOptions) ([]*gitlab.Branch, *gitlab.Response, error) {
			panic("ListProjectBranches not mocked")
		},
		MockGetBranch: func(projectID int, branch string) (*gitlab.Branch, *gitlab.Response, error) {
			panic("GetBranch not mocked")
		},
		MockCreateBranch: func(projectID int, opt *gitlab.CreateBranchOptions) (*gitlab.Branch, *gitlab.Response, error) {
			panic("CreateBranch not mocked")
		},
		MockDeleteBranch: func(projectID int, branch string) (*gitlab.Response, error) {
			panic("DeleteBranch not mocked")
		},
		MockGetCommit: func(projectID int, sha string) (*gitlab.Commit, *gitlab.Response, error) {
			panic("GetCommit not mocked")
		},
		MockListProjectCommits: func(projectID int, opt *gitlab.ListCommitsOptions) ([]*gitlab.Commit, *gitlab.Response, error) {
			panic("ListProjectCommits not mocked")
		},
		MockCreateCommit: func(projectID int, opt *gitlab.CreateCommitOptions) (*gitlab.Commit, *gitlab.Response, error) {
			panic("CreateCommit not mocked")
		},
		MockGetBranchHeadCommitID: func(branch string) (string, error) {
			panic("GetBranchHeadCommitID not mocked")
		},
	}
}


func (m *MockClient) GetProject(projectID interface{}) (*gitlab.Project, *gitlab.Response, error) {
	return m.MockGetProject(projectID)
}

func (m *MockClient) GetFile(projectID int, filePath string, options *gitlab.GetFileOptions) (*gitlab.File, *gitlab.Response, error) {
	return m.MockGetFile(projectID, filePath, options)
}

func (m *MockClient) CreateFile(projectID int, filePath string, opt *gitlab.CreateFileOptions) (*gitlab.File, *gitlab.Response, error) {
	return m.MockCreateFile(projectID, filePath, opt)
}

func (m *MockClient) UpdateFile(projectID int, filePath string, opt *gitlab.UpdateFileOptions) (*gitlab.File, *gitlab.Response, error) {
	return m.MockUpdateFile(projectID, filePath, opt)
}

func (m *MockClient) DeleteFile(projectID int, filePath string, opt *gitlab.DeleteFileOptions) (*gitlab.Response, error) {
	return m.MockDeleteFile(projectID, filePath, opt)
}

func (m *MockClient) ListProjectBranches(projectID int, opt *gitlab.ListBranchesOptions) ([]*gitlab.Branch, *gitlab.Response, error) {
	return m.MockListProjectBranches(projectID, opt)
}

func (m *MockClient) GetBranch(projectID int, branch string) (*gitlab.Branch, *gitlab.Response, error) {
	return m.MockGetBranch(projectID, branch)
}

func (m *MockClient) CreateBranch(projectID int, opt *gitlab.CreateBranchOptions) (*gitlab.Branch, *gitlab.Response, error) {
	return m.MockCreateBranch(projectID, opt)
}

func (m *MockClient) DeleteBranch(projectID int, branch string) (*gitlab.Response, error) {
	return m.MockDeleteBranch(projectID, branch)
}

func (m *MockClient) GetCommit(projectID int, sha string) (*gitlab.Commit, *gitlab.Response, error) {
	return m.MockGetCommit(projectID, sha)
}

func (m *MockClient) ListProjectCommits(projectID int, opt *gitlab.ListCommitsOptions) ([]*gitlab.Commit, *gitlab.Response, error) {
	return m.MockListProjectCommits(projectID, opt)
}

func (m *MockClient) CreateCommit(projectID int, opt *gitlab.CreateCommitOptions) (*gitlab.Commit, *gitlab.Response, error) {
	return m.MockCreateCommit(projectID, opt)
}

func (m *MockClient) GetBranchHeadCommitID(branch string) (string, error) {
	return m.MockGetBranchHeadCommitID(branch)
}

// Helper function to create a gitlab.Response
func NewGitlabResponse(statusCode int) *gitlab.Response {
	return &gitlab.Response{
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
}

// Helper function to create a gitlab.Project
func NewGitlabProject(id int, name string, defaultBranch string) *gitlab.Project {
	return &gitlab.Project{
		ID:            id,
		Name:          name,
		DefaultBranch: defaultBranch,
		PathWithNamespace: fmt.Sprintf("group/%s", name),
	}
}