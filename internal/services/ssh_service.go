package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"golang.org/x/crypto/ssh"
	"habibi-go/internal/models"
)

type SSHService struct {
	connections map[int]*SSHConnection // project ID -> connection
}

type SSHConnection struct {
	client   *ssh.Client
	config   *models.ProjectConfig
	host     string
	port     string
}

func NewSSHService() *SSHService {
	return &SSHService{
		connections: make(map[int]*SSHConnection),
	}
}

// Connect establishes an SSH connection for a project
func (s *SSHService) Connect(project *models.Project) error {
	// Parse SSH configuration from project config
	config, err := s.ParseProjectSSHConfig(project)
	if err != nil {
		return fmt.Errorf("failed to parse SSH config: %w", err)
	}
	
	if config.SSHHost == "" {
		return fmt.Errorf("SSH host not configured for project")
	}
	
	// Parse host and port
	host, port := s.parseHostPort(config.SSHHost, config.SSHPort)
	
	// Load SSH key
	key, err := s.loadSSHKey(config.SSHKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load SSH key: %w", err)
	}
	
	// Extract username from host
	parts := strings.Split(host, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid SSH host format, expected user@hostname")
	}
	username := parts[0]
	hostname := parts[1]
	
	// Create SSH client config
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Add proper host key verification
	}
	
	// Connect
	addr := fmt.Sprintf("%s:%s", hostname, port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	
	// Store connection
	s.connections[project.ID] = &SSHConnection{
		client: client,
		config: config,
		host:   host,
		port:   port,
	}
	
	return nil
}

// Disconnect closes the SSH connection for a project
func (s *SSHService) Disconnect(projectID int) error {
	conn, exists := s.connections[projectID]
	if !exists {
		return nil
	}
	
	err := conn.client.Close()
	delete(s.connections, projectID)
	return err
}

// ExecuteSetupCommand runs the project setup command with environment variables
func (s *SSHService) ExecuteSetupCommand(project *models.Project, worktreePath string) (string, error) {
	conn, err := s.getConnection(project)
	if err != nil {
		return "", err
	}
	
	// Create a new session
	session, err := conn.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()
	
	// Prepare environment variables
	envVars := s.prepareEnvironmentVars(conn.config, project.Path, worktreePath)
	
	// Build command with environment variables
	var cmdBuilder strings.Builder
	for key, value := range envVars {
		cmdBuilder.WriteString(fmt.Sprintf("export %s=%q; ", key, value))
	}
	
	// Add the actual setup command
	if conn.config.RemoteSetupCmd != "" {
		// Replace variables in the setup command
		setupCmd := s.expandVariables(conn.config.RemoteSetupCmd, envVars)
		cmdBuilder.WriteString(setupCmd)
	}
	
	// Execute command
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	
	err = session.Run(cmdBuilder.String())
	if err != nil {
		return "", fmt.Errorf("setup command failed: %w\nstderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// ExecuteCommand runs a command on the remote server
func (s *SSHService) ExecuteCommand(projectID int, command string) (string, error) {
	conn, exists := s.connections[projectID]
	if !exists {
		return "", fmt.Errorf("no SSH connection for project %d", projectID)
	}
	
	// Create a new session
	session, err := conn.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()
	
	// Execute command
	var stdout bytes.Buffer
	session.Stdout = &stdout
	
	err = session.Run(command)
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	
	return stdout.String(), nil
}

// CreateRemoteWorktree creates a worktree on the remote server
func (s *SSHService) CreateRemoteWorktree(project *models.Project, branchName string, worktreePath string) error {
	conn, err := s.getConnection(project)
	if err != nil {
		return err
	}
	
	// Navigate to project directory and create worktree
	cmd := fmt.Sprintf("cd %s && git worktree add %s %s", 
		conn.config.RemoteProjectPath, 
		worktreePath, 
		branchName)
	
	_, err = s.ExecuteCommand(project.ID, cmd)
	return err
}

// StreamClaudeOutput starts a Claude process on the remote server and streams its output
func (s *SSHService) StreamClaudeOutput(projectID int, worktreePath string, args []string) (io.ReadCloser, error) {
	conn, exists := s.connections[projectID]
	if !exists {
		return nil, fmt.Errorf("no SSH connection for project %d", projectID)
	}
	
	// Create a new session
	session, err := conn.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	
	// Get stdout pipe
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	
	// Build claude command
	claudeCmd := fmt.Sprintf("cd %s && claude %s", worktreePath, strings.Join(args, " "))
	
	// Start the command
	err = session.Start(claudeCmd)
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}
	
	// Return a wrapper that closes the session when done
	return &sshStreamReader{
		reader:  stdout,
		session: session,
	}, nil
}

// Helper methods

// ParseProjectSSHConfig extracts SSH configuration from a project
func (s *SSHService) ParseProjectSSHConfig(project *models.Project) (*models.ProjectConfig, error) {
	configData, ok := project.Config["ssh"]
	if !ok {
		// Try to get it from the root config
		var config models.ProjectConfig
		if project.Config["ssh_host"] != nil {
			config.SSHHost, _ = project.Config["ssh_host"].(string)
			config.SSHPort, _ = project.Config["ssh_port"].(int)
			config.SSHKeyPath, _ = project.Config["ssh_key_path"].(string)
			config.RemoteProjectPath, _ = project.Config["remote_project_path"].(string)
			config.RemoteSetupCmd, _ = project.Config["remote_setup_cmd"].(string)
			
			if envVars, ok := project.Config["environment_vars"].(map[string]interface{}); ok {
				config.EnvironmentVars = make(map[string]string)
				for k, v := range envVars {
					config.EnvironmentVars[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		return &config, nil
	}
	
	// Parse nested SSH config
	var config models.ProjectConfig
	if sshConfig, ok := configData.(map[string]interface{}); ok {
		config.SSHHost, _ = sshConfig["host"].(string)
		config.SSHPort, _ = sshConfig["port"].(int)
		config.SSHKeyPath, _ = sshConfig["key_path"].(string)
		config.RemoteProjectPath, _ = sshConfig["remote_project_path"].(string)
		config.RemoteSetupCmd, _ = sshConfig["remote_setup_cmd"].(string)
		
		if envVars, ok := sshConfig["environment_vars"].(map[string]interface{}); ok {
			config.EnvironmentVars = make(map[string]string)
			for k, v := range envVars {
				config.EnvironmentVars[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	
	return &config, nil
}

func (s *SSHService) parseHostPort(host string, port int) (string, string) {
	if port == 0 {
		port = 22
	}
	return host, fmt.Sprintf("%d", port)
}

func (s *SSHService) loadSSHKey(keyPath string) (ssh.Signer, error) {
	// Expand home directory if needed
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}
	
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}
	
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}
	
	return signer, nil
}

func (s *SSHService) getConnection(project *models.Project) (*SSHConnection, error) {
	conn, exists := s.connections[project.ID]
	if !exists {
		// Try to connect
		if err := s.Connect(project); err != nil {
			return nil, err
		}
		conn = s.connections[project.ID]
	}
	return conn, nil
}

func (s *SSHService) prepareEnvironmentVars(config *models.ProjectConfig, projectPath, worktreePath string) map[string]string {
	vars := make(map[string]string)
	
	// Copy custom environment variables
	for k, v := range config.EnvironmentVars {
		vars[k] = v
	}
	
	// Add standard variables
	vars["PROJECT_PATH"] = projectPath
	vars["WORKTREE_PATH"] = worktreePath
	vars["REMOTE_PROJECT_PATH"] = config.RemoteProjectPath
	
	return vars
}

func (s *SSHService) expandVariables(command string, vars map[string]string) string {
	result := command
	for key, value := range vars {
		result = strings.ReplaceAll(result, "$"+key, value)
		result = strings.ReplaceAll(result, "${"+key+"}", value)
	}
	return result
}

// sshStreamReader wraps an SSH stdout reader and closes the session when done
type sshStreamReader struct {
	reader  io.Reader
	session *ssh.Session
}

func (r *sshStreamReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *sshStreamReader) Close() error {
	return r.session.Close()
}