package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// GetWorkingTreeDiff returns the diff of the working tree and commits not in base branch
func (s *GitService) GetWorkingTreeDiff(worktreePath, baseBranch string) ([]DiffFile, error) {
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	
	var diffFiles []DiffFile
	
	// Determine the comparison base - try to find the most recent common ancestor
	var compareBase string
	
	// If a base branch is provided, try it first
	if baseBranch != "" {
		// Try with origin prefix first, then without
		for _, branchVariant := range []string{"origin/" + baseBranch, baseBranch} {
			cmd := exec.Command("git", "merge-base", "HEAD", branchVariant)
			cmd.Dir = worktreePath
			if mergeBaseOutput, err := cmd.Output(); err == nil {
				compareBase = strings.TrimSpace(string(mergeBaseOutput))
				break
			} else {
			}
		}
	}
	
	// If that didn't work, try to find which branch actually contains our merge base
	// by checking all remote branches
	if compareBase == "" {
		cmd := exec.Command("git", "branch", "-r")
		cmd.Dir = worktreePath
		if branchesOutput, err := cmd.Output(); err == nil {
			branches := strings.Split(string(branchesOutput), "\n")
			for _, branch := range branches {
				branch = strings.TrimSpace(branch)
				if branch == "" || strings.Contains(branch, "HEAD ->") {
					continue
				}
				
				cmd = exec.Command("git", "merge-base", "HEAD", branch)
				cmd.Dir = worktreePath
				if mergeBaseOutput, err := cmd.Output(); err == nil {
					testBase := strings.TrimSpace(string(mergeBaseOutput))
					
					// Check if this branch's tip is the same as our HEAD (meaning we're ON this branch)
					cmd = exec.Command("git", "rev-parse", branch)
					cmd.Dir = worktreePath
					if branchTipOutput, err := cmd.Output(); err == nil {
						branchTip := strings.TrimSpace(string(branchTipOutput))
						cmd = exec.Command("git", "rev-parse", "HEAD")
						cmd.Dir = worktreePath
						if headOutput, err := cmd.Output(); err == nil {
							head := strings.TrimSpace(string(headOutput))
							if head != branchTip {
								// We're not on this branch, so it could be our base
								if compareBase == "" || testBase != head {
									compareBase = testBase
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Fallback to common default branches
	if compareBase == "" {
		for _, defaultBranch := range []string{"origin/main", "origin/master", "main", "master"} {
			cmd := exec.Command("git", "merge-base", "HEAD", defaultBranch)
			cmd.Dir = worktreePath
			if mergeBaseOutput, err := cmd.Output(); err == nil {
				compareBase = strings.TrimSpace(string(mergeBaseOutput))
				break
			} else {
			}
		}
	}
	
	// If still no base found, use HEAD as comparison (will show only uncommitted changes)
	if compareBase == "" {
		compareBase = "HEAD"
	}
	
	
	// Check if current HEAD is the same as merge base (no changes)
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = worktreePath
	if headOutput, err := cmd.Output(); err == nil {
		currentHead := strings.TrimSpace(string(headOutput))
		if currentHead == compareBase && compareBase != "HEAD" {
			// No commits ahead, check for uncommitted changes only
			cmd = exec.Command("git", "status", "--porcelain")
			cmd.Dir = worktreePath
			if statusOutput, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(statusOutput))) == 0 {
				// No uncommitted changes either
				return diffFiles, nil
			} else {
			}
		}
	} else {
	}
	
	// Get diff against the comparison base
	cmd = exec.Command("git", "diff", "--name-status", compareBase)
	cmd.Dir = worktreePath
	diffOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}
	
	// Parse diff output
	lines := strings.Split(string(diffOutput), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		status := parts[0]
		path := parts[1]
		
		var fileStatus string
		switch status[0] {
		case 'A':
			fileStatus = "added"
		case 'D':
			fileStatus = "deleted"
		case 'M':
			fileStatus = "modified"
		default:
			fileStatus = "modified"
		}
		
		// Get full diff for this file
		cmd = exec.Command("git", "diff", compareBase, "--", path)
		cmd.Dir = worktreePath
		fileDiffOutput, _ := cmd.Output()
		
		// Get stats
		cmd = exec.Command("git", "diff", "--numstat", compareBase, "--", path)
		cmd.Dir = worktreePath
		numstatOutput, _ := cmd.Output()
		
		var additions, deletions int
		if len(numstatOutput) > 0 {
			parts := strings.Fields(string(numstatOutput))
			if len(parts) >= 3 {
				additions, _ = strconv.Atoi(parts[0])
				deletions, _ = strconv.Atoi(parts[1])
			}
		}
		
		diffFile := DiffFile{
			Path:      path,
			Status:    fileStatus,
			Diff:      string(fileDiffOutput),
			Additions: additions,
			Deletions: deletions,
		}
		
		if diffFile.Diff != "" || diffFile.Status == "deleted" {
			diffFiles = append(diffFiles, diffFile)
		}
	}
	
	return diffFiles, nil
}

// GetCurrentBranch returns the current branch for a project
func (s *GitService) GetCurrentBranch(projectPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// RebaseWorktree rebases the current branch onto another branch
func (s *GitService) RebaseWorktree(worktreePath, targetBranch string) error {
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	// Fetch latest changes
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = worktreePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	
	// Perform rebase
	cmd = exec.Command("git", "rebase", fmt.Sprintf("origin/%s", targetBranch))
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a conflict
		if strings.Contains(string(output), "conflict") {
			return fmt.Errorf("rebase conflict: %s", string(output))
		}
		return fmt.Errorf("failed to rebase: %w", err)
	}
	
	return nil
}

// PushBranch pushes a branch to remote
func (s *GitService) PushBranch(worktreePath, localBranch, remoteBranch string) error {
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree does not exist: %s", worktreePath)
	}
	
	// Push to remote
	cmd := exec.Command("git", "push", "origin", fmt.Sprintf("%s:%s", localBranch, remoteBranch))
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push: %s", string(output))
	}
	
	return nil
}


// DiffFile represents a file in a git diff
type DiffFile struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Diff      string `json:"diff"`
	isUntracked bool // internal field, not exported to JSON
}