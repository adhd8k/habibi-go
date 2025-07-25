package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"habibi-go/internal/database/repositories"
	"habibi-go/internal/models"
)

type SlashCommandService struct {
	sessionService *SessionService
	agentService   *ClaudeSessionService
	sessionRepo    *repositories.SessionRepository
	projectRepo    *repositories.ProjectRepository
}

func NewSlashCommandService(sessionService *SessionService, agentService *ClaudeSessionService, sessionRepo *repositories.SessionRepository, projectRepo *repositories.ProjectRepository) *SlashCommandService {
	return &SlashCommandService{
		sessionService: sessionService,
		agentService:   agentService,
		sessionRepo:    sessionRepo,
		projectRepo:    projectRepo,
	}
}

// GetAvailableCommands returns all available slash commands for a session
func (s *SlashCommandService) GetAvailableCommands(ctx context.Context, sessionID int) ([]models.SlashCommand, error) {
	// Get session to find project path
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	project, err := s.projectRepo.GetByID(session.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Start with built-in commands
	commands := s.getBuiltinCommands()

	// Scan project .claude/commands/
	projectCommandsPath := filepath.Join(project.Path, ".claude", "commands")
	
	// Also check the worktree path if this is a worktree session
	if session.WorktreePath != "" && session.WorktreePath != project.Path {
		worktreeCommandsPath := filepath.Join(session.WorktreePath, ".claude", "commands")
		worktreeCmds, err := s.scanDirectory(worktreeCommandsPath, false)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to scan worktree commands: %w", err)
		}
		commands = append(commands, worktreeCmds...)
	}
	
	projectCmds, err := s.scanDirectory(projectCommandsPath, false)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to scan project commands: %w", err)
	}
	commands = append(commands, projectCmds...)

	// Scan ~/.claude/commands/
	homeDir, err := os.UserHomeDir()
	if err == nil {
		homeCommandsPath := filepath.Join(homeDir, ".claude", "commands")
		homeCmds, err := s.scanDirectory(homeCommandsPath, false)
		if err != nil && !os.IsNotExist(err) {
			// Don't fail if home commands can't be read
			fmt.Printf("Warning: failed to scan home commands: %v\n", err)
		} else {
			commands = append(commands, homeCmds...)
		}
	}

	return commands, nil
}

// ExecuteCommand executes a slash command and returns the result
func (s *SlashCommandService) ExecuteCommand(ctx context.Context, sessionID int, command string, args string) (*models.CommandResult, error) {
	// Parse command name
	cmdName := strings.TrimPrefix(command, "/")
	
	// Check if it's a built-in command
	if result, handled := s.handleBuiltinCommand(ctx, sessionID, cmdName, args); handled {
		return result, nil
	}

	// For custom commands, we need to execute through Claude
	fullCommand := command
	if args != "" {
		fullCommand = command + " " + args
	}

	// Send to Claude as a special command
	err := s.agentService.SendMessage(ctx, sessionID, fullCommand)
	if err != nil {
		return &models.CommandResult{
			Type: models.CommandResultTypeError,
			Data: map[string]string{"message": fmt.Sprintf("Failed to execute command: %v", err)},
		}, nil
	}

	// The Claude response will come through the normal streaming mechanism
	return &models.CommandResult{
		Type: models.CommandResultTypeClaudeMessage,
		Data: map[string]string{"message": "Command sent to Claude"},
	}, nil
}

// getBuiltinCommands returns the list of built-in commands
func (s *SlashCommandService) getBuiltinCommands() []models.SlashCommand {
	return []models.SlashCommand{
		{
			Name:        "/clear",
			Description: "Clear the chat history",
			IsBuiltin:   true,
			Category:    "Chat",
		},
		{
			Name:        "/help",
			Description: "Show available commands and their descriptions",
			IsBuiltin:   true,
			Category:    "Help",
		},
		{
			Name:        "/status",
			Description: "Show current session and agent status",
			IsBuiltin:   true,
			Category:    "Info",
		},
	}
}

// handleBuiltinCommand handles built-in commands
func (s *SlashCommandService) handleBuiltinCommand(ctx context.Context, sessionID int, command string, args string) (*models.CommandResult, bool) {
	switch command {
	case "clear":
		// Clear chat history
		err := s.agentService.ClearChatHistory(ctx, sessionID)
		if err != nil {
			return &models.CommandResult{
				Type: models.CommandResultTypeError,
				Data: map[string]string{"message": fmt.Sprintf("Failed to clear chat: %v", err)},
			}, true
		}
		return &models.CommandResult{
			Type: models.CommandResultTypeClearChat,
			Data: nil,
		}, true

	case "help":
		// Get all available commands
		commands, err := s.GetAvailableCommands(ctx, sessionID)
		if err != nil {
			return &models.CommandResult{
				Type: models.CommandResultTypeError,
				Data: map[string]string{"message": fmt.Sprintf("Failed to get commands: %v", err)},
			}, true
		}
		
		// Group commands by category
		grouped := make(map[string][]models.SlashCommand)
		for _, cmd := range commands {
			category := cmd.Category
			if category == "" {
				category = "Custom"
			}
			grouped[category] = append(grouped[category], cmd)
		}
		
		return &models.CommandResult{
			Type: models.CommandResultTypeShowHelp,
			Data: map[string]interface{}{
				"commands": commands,
				"grouped":  grouped,
			},
		}, true

	case "status":
		// Get session and agent status
		session, err := s.sessionRepo.GetByID(sessionID)
		if err != nil {
			return &models.CommandResult{
				Type: models.CommandResultTypeError,
				Data: map[string]string{"message": fmt.Sprintf("Failed to get session: %v", err)},
			}, true
		}

		// For now, use a mock agent ID since sessions don't have agent IDs
		agent, err := s.agentService.GetAgent(ctx, 1)
		if err != nil {
			return &models.CommandResult{
				Type: models.CommandResultTypeError,
				Data: map[string]string{"message": fmt.Sprintf("Failed to get agent: %v", err)},
			}, true
		}

		return &models.CommandResult{
			Type: models.CommandResultTypeStatus,
			Data: map[string]interface{}{
				"session": session,
				"agent":   agent,
			},
		}, true

	default:
		return nil, false
	}
}

// scanDirectory scans a directory for slash command files
func (s *SlashCommandService) scanDirectory(dir string, nested bool) ([]models.SlashCommand, error) {
	var commands []models.SlashCommand

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return commands, nil
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Extract command name from path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Convert path to command name
		cmdName := strings.TrimSuffix(relPath, ".md")
		cmdName = strings.ReplaceAll(cmdName, string(os.PathSeparator), ":")
		
		// Extract description from file (first non-empty line after frontmatter)
		description := s.extractDescription(path)
		
		// Determine category from path
		category := ""
		if strings.Contains(cmdName, ":") {
			parts := strings.Split(cmdName, ":")
			category = strings.Title(parts[0])
		}

		commands = append(commands, models.SlashCommand{
			Name:        "/" + cmdName,
			Description: description,
			Path:        path,
			IsBuiltin:   false,
			Category:    category,
		})

		return nil
	})

	return commands, err
}

// extractDescription extracts the description from a command file
func (s *SlashCommandService) extractDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	frontmatterCount := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Handle frontmatter
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
			} else {
				frontmatterCount++
				if frontmatterCount >= 2 {
					inFrontmatter = false
				}
			}
			continue
		}
		
		// Look for description in frontmatter
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.TrimSpace(desc)
			desc = strings.Trim(desc, "\"'")
			return desc
		}
		
		// If not in frontmatter and line is not empty, use as description
		if !inFrontmatter && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			return strings.TrimSpace(line)
		}
	}
	
	return ""
}