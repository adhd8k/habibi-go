package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"habibi-go/internal/util"
)

type GitService struct {
	gitUtil *util.GitUtil
}

func NewGitService() *GitService {
	return &GitService{
		gitUtil: util.NewGitUtil(),
	}
}

// CreateWorktree creates a new Git worktree for a session
func (s *GitService) CreateWorktree(projectPath, sessionName, branchName string) (string, error) {
	// Validate the project is a Git repository
	if err := s.gitUtil.ValidateRepository(projectPath); err != nil {
		return "", fmt.Errorf("invalid Git repository: %w", err)
	}
	
	// Create worktree path
	worktreePath := filepath.Join(projectPath, ".worktrees", sessionName)
	
	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return "", fmt.Errorf("worktree already exists: %s", worktreePath)
	}
	
	// Create worktree directory
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory: %w", err)
	}
	
	// Check if branch exists
	branchExists, err := s.gitUtil.BranchExists(projectPath, branchName)
	if err != nil {
		return "", fmt.Errorf("failed to check branch existence: %w", err)
	}
	
	// Create the worktree
	var cmd *exec.Cmd
	if branchExists {
		// Branch exists, create worktree from existing branch
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
	} else {
		// Branch doesn't exist, create new branch and worktree
		cmd = exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
	}
	
	cmd.Dir = projectPath
	
	if output, err := cmd.CombinedOutput(); err != nil {
		// Clean up partial creation
		os.RemoveAll(worktreePath)
		return "", fmt.Errorf("failed to create worktree: %w, output: %s", err, string(output))
	}
	
	return worktreePath, nil
}

// RemoveWorktree removes a Git worktree
func (s *GitService) RemoveWorktree(projectPath, worktreePath string) error {
	// Validate the project is a Git repository
	if err := s.gitUtil.ValidateRepository(projectPath); err != nil {
		return fmt.Errorf("invalid Git repository: %w", err)
	}
	
	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	// Remove the worktree from Git
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = projectPath
	
	if output, err := cmd.CombinedOutput(); err != nil {
		// If git worktree remove fails, try force removal
		forceCmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
		forceCmd.Dir = projectPath
		
		if forceOutput, forceErr := forceCmd.CombinedOutput(); forceErr != nil {
			return fmt.Errorf("failed to remove worktree: %w, output: %s, force output: %s", 
				err, string(output), string(forceOutput))
		}
	}
	
	// Clean up any remaining files
	if err := os.RemoveAll(worktreePath); err != nil {
		// Log warning but don't fail - the important part is Git cleanup
		fmt.Printf("Warning: failed to clean up worktree directory: %v\n", err)
	}
	
	return nil
}

// ListWorktrees returns all worktrees for a project
func (s *GitService) ListWorktrees(projectPath string) ([]WorktreeInfo, error) {
	// Validate the project is a Git repository
	if err := s.gitUtil.ValidateRepository(projectPath); err != nil {
		return nil, fmt.Errorf("invalid Git repository: %w", err)
	}
	
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = projectPath
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	
	return s.parseWorktreeList(string(output)), nil
}

// WorktreeInfo represents information about a Git worktree
type WorktreeInfo struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Bare   bool   `json:"bare"`
}

// parseWorktreeList parses the output of 'git worktree list --porcelain'
func (s *GitService) parseWorktreeList(output string) []WorktreeInfo {
	lines := strings.Split(output, "\n")
	var worktrees []WorktreeInfo
	var current WorktreeInfo
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = WorktreeInfo{}
			}
			continue
		}
		
		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch ")
		} else if strings.HasPrefix(line, "HEAD ") {
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		} else if line == "bare" {
			current.Bare = true
		}
	}
	
	// Add the last worktree if exists
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	
	return worktrees
}

// GetWorktreeStatus returns the status of a specific worktree
func (s *GitService) GetWorktreeStatus(worktreePath string) (*WorktreeStatus, error) {
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	status := &WorktreeStatus{
		Path: worktreePath,
	}
	
	// Get current branch
	if branch, err := s.gitUtil.GetCurrentBranch(worktreePath); err == nil {
		status.Branch = branch
	}
	
	// Get commit hash
	if commit, err := s.gitUtil.GetCommitHash(worktreePath); err == nil {
		status.Commit = commit
	}
	
	// Check for uncommitted changes
	if hasChanges, err := s.gitUtil.HasUncommittedChanges(worktreePath); err == nil {
		status.HasUncommittedChanges = hasChanges
	}
	
	// Get detailed status
	if gitStatus, err := s.gitUtil.GetStatus(worktreePath); err == nil {
		status.GitStatus = gitStatus
	}
	
	return status, nil
}

// WorktreeStatus represents the status of a Git worktree
type WorktreeStatus struct {
	Path                  string `json:"path"`
	Branch                string `json:"branch"`
	Commit                string `json:"commit"`
	HasUncommittedChanges bool   `json:"has_uncommitted_changes"`
	GitStatus             string `json:"git_status"`
}

// SwitchBranch switches the worktree to a different branch
func (s *GitService) SwitchBranch(worktreePath, branchName string) error {
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	// Check for uncommitted changes
	hasChanges, err := s.gitUtil.HasUncommittedChanges(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}
	
	if hasChanges {
		return fmt.Errorf("cannot switch branch: worktree has uncommitted changes")
	}
	
	// Check if branch exists
	branchExists, err := s.gitUtil.BranchExists(worktreePath, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	
	if branchExists {
		// Switch to existing branch
		if err := s.gitUtil.CheckoutBranch(worktreePath, branchName); err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}
	} else {
		// Create new branch
		if err := s.gitUtil.CreateBranch(worktreePath, branchName, ""); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}
	
	return nil
}

// PullLatest pulls the latest changes for the current branch
func (s *GitService) PullLatest(worktreePath string) error {
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	// Check for uncommitted changes
	hasChanges, err := s.gitUtil.HasUncommittedChanges(worktreePath)
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}
	
	if hasChanges {
		return fmt.Errorf("cannot pull: worktree has uncommitted changes")
	}
	
	// Fetch latest changes
	if err := s.gitUtil.FetchAll(worktreePath); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	
	// Pull changes
	cmd := exec.Command("git", "pull")
	cmd.Dir = worktreePath
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull: %w, output: %s", err, string(output))
	}
	
	return nil
}

// CleanupStaleWorktrees removes worktrees that no longer exist on disk
func (s *GitService) CleanupStaleWorktrees(projectPath string) error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = projectPath
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to prune worktrees: %w, output: %s", err, string(output))
	}
	
	return nil
}