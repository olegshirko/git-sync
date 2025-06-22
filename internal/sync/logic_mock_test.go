package sync

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// MockRepositoryManager - мок для repository.Manager
type MockRepositoryManager struct {
	MockCreateTempRepoPath func(repoName string) string
	MockCleanTempDir       func() error
	MockClone              func(repoURL, path, token, sshKeyPath string) (*git.Repository, error)
	MockPull               func(repo *git.Repository, token, sshKeyPath string) error
	MockPush               func(repo *git.Repository, token, sshKeyPath string) error
}

func (m *MockRepositoryManager) CreateTempRepoPath(repoName string) string {
	if m.MockCreateTempRepoPath != nil {
		return m.MockCreateTempRepoPath(repoName)
	}
	return ""
}

func (m *MockRepositoryManager) CleanTempDir() error {
	if m.MockCleanTempDir != nil {
		return m.MockCleanTempDir()
	}
	return nil
}

func (m *MockRepositoryManager) Clone(repoURL, path, token, sshKeyPath string) (*git.Repository, error) {
	if m.MockClone != nil {
		return m.MockClone(repoURL, path, token, sshKeyPath)
	}
	return nil, nil
}

func (m *MockRepositoryManager) Pull(repo *git.Repository, token, sshKeyPath string) error {
	if m.MockPull != nil {
		return m.MockPull(repo, token, sshKeyPath)
	}
	return nil
}

func (m *MockRepositoryManager) Push(repo *git.Repository, token, sshKeyPath string) error {
	if m.MockPush != nil {
		return m.MockPush(repo, token, sshKeyPath)
	}
	return nil
}

// MockGitRepository - мок для *git.Repository
type MockGitRepository struct {
	MockWorktree     func() (*git.Worktree, error)
	MockRemote       func(name string) (*git.Remote, error)
	MockCreateRemote func(config *gitconfig.RemoteConfig) (*git.Remote, error)
	MockBranches     func() (*object.ReferenceIter, error)
	MockReference    func(name plumbing.ReferenceName, resolve bool) (*plumbing.Reference, error)
	MockPush         func(o *git.PushOptions) error
}

func (m *MockGitRepository) Worktree() (*git.Worktree, error) {
	if m.MockWorktree != nil {
		return m.MockWorktree()
	}
	return nil, nil
}

func (m *MockGitRepository) Remote(name string) (*git.Remote, error) {
	if m.MockRemote != nil {
		return m.MockRemote(name)
	}
	return nil, nil
}

func (m *MockGitRepository) CreateRemote(config *gitconfig.RemoteConfig) (*git.Remote, error) {
	if m.MockCreateRemote != nil {
		return m.MockCreateRemote(config)
	}
	return nil, nil
}

func (m *MockGitRepository) Branches() (*object.ReferenceIter, error) {
	if m.MockBranches != nil {
		return m.MockBranches()
	}
	return nil, nil
}

func (m *MockGitRepository) Reference(name plumbing.ReferenceName, resolve bool) (*plumbing.Reference, error) {
	if m.MockReference != nil {
		return m.MockReference(name, resolve)
	}
	return nil, nil
}

func (m *MockGitRepository) Push(o *git.PushOptions) error {
	if m.MockPush != nil {
		return m.MockPush(o)
	}
	return nil
}

// MockGitWorktree - мок для *git.Worktree
type MockGitWorktree struct {
	MockPull     func(o *git.PullOptions) error
	MockCheckout func(opts *git.CheckoutOptions) error
	MockReset    func(opts *git.ResetOptions) error
}

func (m *MockGitWorktree) Pull(o *git.PullOptions) error {
	if m.MockPull != nil {
		return m.MockPull(o)
	}
	return nil
}

func (m *MockGitWorktree) Checkout(opts *git.CheckoutOptions) error {
	if m.MockCheckout != nil {
		return m.MockCheckout(opts)
	}
	return nil
}

func (m *MockGitWorktree) Reset(opts *git.ResetOptions) error {
	if m.MockReset != nil {
		return m.MockReset(opts)
	}
	return nil
}

// MockGitRemote - мок для *git.Remote
type MockGitRemote struct {
	MockConfig func() *gitconfig.RemoteConfig
	MockFetch  func(o *git.FetchOptions) error
}

func (m *MockGitRemote) Config() *gitconfig.RemoteConfig {
	if m.MockConfig != nil {
		return m.MockConfig()
	}
	return &gitconfig.RemoteConfig{}
}

func (m *MockGitRemote) Fetch(o *git.FetchOptions) error {
	if m.MockFetch != nil {
		return m.MockFetch(o)
	}
	return nil
}

// MockReferenceIter - мок для object.ReferenceIter
type MockReferenceIter struct {
	Refs  []*plumbing.Reference
	Index int
	Mu    sync.Mutex
}

func (m *MockReferenceIter) Next() (*plumbing.Reference, error) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if m.Index >= len(m.Refs) {
		return nil, object.ErrIterDone
	}
	ref := m.Refs[m.Index]
	m.Index++
	return ref, nil
}

func (m *MockReferenceIter) ForEach(f func(*plumbing.Reference) error) error {
	for _, ref := range m.Refs {
		if err := f(ref); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockReferenceIter) Close() {
	// Do nothing
}

// MockAuthMethod - мок для transport.AuthMethod
type MockAuthMethod struct {
	NameFunc       func() string
	StringFunc     func() string
	SetContextFunc func(c context.Context)
}

func (m *MockAuthMethod) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock-auth"
}

func (m *MockAuthMethod) String() string {
	if m.StringFunc != nil {
		return m.StringFunc()
	}
	return "mock-auth-string"
}

func (m *MockAuthMethod) SetContext(c context.Context) {
	if m.SetContextFunc != nil {
		m.SetContextFunc(c)
	}
}

// Helper function to create a mock SSH auth method that returns an error
func NewMockSSHErrorAuth(err error) transport.AuthMethod {
	return &MockAuthMethod{
		NameFunc:   func() string { return "ssh" },
		StringFunc: func() string { return "ssh-mock-error" },
		SetContextFunc: func(c context.Context) {
			// Simulate error during context setting if needed
			if err != nil {
				// In a real scenario, you might want to store the error
				// or make it accessible for assertions.
				// For now, we just acknowledge it.
			}
		},
	}
}

// Helper function to create a mock HTTP auth method
func NewMockHTTPAuth(username, password string) transport.AuthMethod {
	return &http.BasicAuth{Username: username, Password: password}
}

// Helper function to create a mock SSH auth method
func NewMockSSHAuth(sshKeyPath string) (transport.AuthMethod, error) {
	// For testing purposes, we can return a dummy SSH auth method
	// or simulate the NewPublicKeysFromFile behavior.
	// For now, let's return a simple mock.
	if sshKeyPath == "/invalid/path/to/ssh/key" {
		return nil, fmt.Errorf("не удалось создать SSH-аутентификацию: %w", os.ErrNotExist)
	}
	return &ssh.PublicKeys{}, nil // Return a dummy PublicKeys struct
}

// MockFileInfo - мок для fs.FileInfo
type MockFileInfo struct {
	mockName string
	size     int64
	mode     fs.FileMode
	modTime  time.Time
	isDir    bool
	sys      interface{}
}

func (m MockFileInfo) Name() string       { return m.mockName }
func (m MockFileInfo) Size() int64        { return m.size }
func (m MockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m MockFileInfo) ModTime() time.Time { return m.modTime }
func (m MockFileInfo) IsDir() bool        { return m.isDir }
func (m MockFileInfo) Sys() interface{}   { return m.sys }

// MockFile - мок для os.File
type MockFile struct {
	mockName  string
	ReadFunc  func(p []byte) (n int, err error)
	CloseFunc func() error
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}
	return 0, nil
}

func (m *MockFile) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockFile) Stat() (fs.FileInfo, error) {
	return MockFileInfo{mockName: m.mockName}, nil
}

func (m *MockFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, nil
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (m *MockFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (m *MockFile) Truncate(size int64) error {
	return nil
}

func (m *MockFile) Sync() error {
	return nil
}

func (m *MockFile) Fd() uintptr {
	return 0
}

func (m *MockFile) Chdir() error {
	return nil
}

func (m *MockFile) Chmod(mode fs.FileMode) error {
	return nil
}

func (m *MockFile) Chown(uid, gid int) error {
	return nil
}

func (m *MockFile) Name() string {
	return m.mockName
}

// MockOsStat - мок для os.Stat
var MockOsStat func(name string) (fs.FileInfo, error)

// MockOsRemoveAll - мок для os.RemoveAll
var MockOsRemoveAll func(path string) error

// MockOsOpenFile - мок для os.OpenFile
var MockOsOpenFile func(name string, flag int, perm fs.FileMode) (*os.File, error)

// MockSshNewPublicKeysFromFile - мок для ssh.NewPublicKeysFromFile
var MockSshNewPublicKeysFromFile func(user, path, password string) (transport.AuthMethod, error)

func init() {
	// Reset mocks to nil before each test run
	MockOsStat = nil
	MockOsRemoveAll = nil
	MockOsOpenFile = nil
	MockSshNewPublicKeysFromFile = nil
}

// Set up mock functions for os package
func SetOsMocks(stat func(name string) (fs.FileInfo, error), removeAll func(path string) error, openFile func(name string, flag int, perm fs.FileMode) (*os.File, error)) {
	MockOsStat = stat
	MockOsRemoveAll = removeAll
	MockOsOpenFile = openFile
}

// Set up mock function for ssh.NewPublicKeysFromFile
func SetSshNewPublicKeysFromFileMock(f func(user, path, password string) (transport.AuthMethod, error)) {
	MockSshNewPublicKeysFromFile = f
}

// Restore original os functions after tests
func RestoreOsFunctions() {
	MockOsStat = nil
	MockOsRemoveAll = nil
	MockOsOpenFile = nil
	MockSshNewPublicKeysFromFile = nil
}
