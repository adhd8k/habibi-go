package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitUtil provides utility functions for Git operations
type GitUtil struct{}

// NewGitUtil creates a new GitUtil instance
func NewGitUtil() *GitUtil {
	return &GitUtil{}
}

// IsGitRepository checks if the given path is a Git repository
func (g *GitUtil) IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	
	// Check if .git exists (either directory or file for worktrees)
	if stat, err := os.Stat(gitPath); err == nil {
		return stat.IsDir() || stat.Mode().IsRegular()
	}
	
	return false
}

// GetCurrentBranch returns the current branch name
func (g *GitUtil) GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "HEAD", nil // Detached HEAD state
	}
	
	return branch, nil
}

// GetRemoteBranches returns a list of remote branches
func (g *GitUtil) GetRemoteBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote branches: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	var branches []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "->") {
			// Remove "origin/" prefix if present
			if strings.HasPrefix(line, "origin/") {
				line = line[7:]
			}
			branches = append(branches, line)
		}
	}
	
	return branches, nil
}

// GetLocalBranches returns a list of local branches
func (g *GitUtil) GetLocalBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "branch")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get local branches: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	var branches []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove the "* " prefix for current branch
			if strings.HasPrefix(line, "* ") {
				line = line[2:]
			}
			branches = append(branches, line)
		}
	}
	
	return branches, nil
}

// BranchExists checks if a branch exists (local or remote)
func (g *GitUtil) BranchExists(repoPath, branchName string) (bool, error) {
	// Check local branches
	localBranches, err := g.GetLocalBranches(repoPath)
	if err == nil {
		for _, branch := range localBranches {
			if branch == branchName {
				return true, nil
			}
		}
	}
	
	// Check remote branches
	remoteBranches, err := g.GetRemoteBranches(repoPath)
	if err == nil {
		for _, branch := range remoteBranches {
			if branch == branchName {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// CreateBranch creates a new local branch
func (g *GitUtil) CreateBranch(repoPath, branchName, baseBranch string) error {
	if baseBranch == "" {
		baseBranch = "HEAD"
	}
	
	cmd := exec.Command("git", "checkout", "-b", branchName, baseBranch)
	cmd.Dir = repoPath
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}
	
	return nil
}

// CheckoutBranch switches to an existing branch
func (g *GitUtil) CheckoutBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	cmd.Dir = repoPath
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}
	
	return nil
}

// FetchAll fetches all remote references
func (g *GitUtil) FetchAll(repoPath string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	
	return nil
}

// GetCommitHash returns the current commit hash
func (g *GitUtil) GetCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// GetStatus returns the Git status
func (g *GitUtil) GetStatus(repoPath string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}
	
	return string(output), nil
}

// HasUncommittedChanges checks if there are uncommitted changes
func (g *GitUtil) HasUncommittedChanges(repoPath string) (bool, error) {
	status, err := g.GetStatus(repoPath)
	if err != nil {
		return false, err
	}
	
	return strings.TrimSpace(status) != "", nil
}

// GetWorktreeList returns a list of active worktrees
func (g *GitUtil) GetWorktreeList(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree list: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	var worktrees []string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			worktrees = append(worktrees, path)
		}
	}
	
	return worktrees, nil
}

// ValidateRepository performs basic validation on a Git repository
func (g *GitUtil) ValidateRepository(repoPath string) error {
	if !g.IsGitRepository(repoPath) {
		return fmt.Errorf("path is not a Git repository: %s", repoPath)
	}
	
	// Check if git command is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found: %w", err)
	}
	
	// Try to get current branch to verify repository is accessible
	if _, err := g.GetCurrentBranch(repoPath); err != nil {
		return fmt.Errorf("failed to access repository: %w", err)
	}
	
	return nil
}