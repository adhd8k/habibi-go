package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcessManager provides utilities for managing system processes
type ProcessManager struct{}

// NewProcessManager creates a new ProcessManager instance
func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID        int       `json:"pid"`
	Command    string    `json:"command"`
	Args       []string  `json:"args"`
	WorkingDir string    `json:"working_dir"`
	StartTime  time.Time `json:"start_time"`
	Status     string    `json:"status"`
	CPUPercent float64   `json:"cpu_percent"`
	MemoryMB   int64     `json:"memory_mb"`
}

// StartProcess starts a new process with the given command and arguments
func (pm *ProcessManager) StartProcess(command string, args []string, workingDir string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = workingDir
	
	// Set up pipes for communication
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}
	
	return cmd, nil
}

// GetProcessInfo retrieves information about a process by PID
func (pm *ProcessManager) GetProcessInfo(pid int) (*ProcessInfo, error) {
	if !pm.ProcessExists(pid) {
		return nil, fmt.Errorf("process %d does not exist", pid)
	}
	
	info := &ProcessInfo{
		PID: pid,
	}
	
	// Get process command line
	if cmdline, err := pm.getProcessCmdline(pid); err == nil {
		if len(cmdline) > 0 {
			info.Command = cmdline[0]
			if len(cmdline) > 1 {
				info.Args = cmdline[1:]
			}
		}
	}
	
	// Get working directory
	if wd, err := pm.getProcessWorkingDir(pid); err == nil {
		info.WorkingDir = wd
	}
	
	// Get process status
	if status, err := pm.getProcessStatus(pid); err == nil {
		info.Status = status
	}
	
	// Get resource usage
	if cpu, mem, err := pm.getProcessResourceUsage(pid); err == nil {
		info.CPUPercent = cpu
		info.MemoryMB = mem
	}
	
	return info, nil
}

// ProcessExists checks if a process with the given PID exists
func (pm *ProcessManager) ProcessExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	// Check if /proc/PID exists (Linux)
	if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); err == nil {
		return true
	}
	
	// Fallback: try to send signal 0 (doesn't actually send a signal)
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// KillProcess terminates a process by PID
func (pm *ProcessManager) KillProcess(pid int) error {
	if !pm.ProcessExists(pid) {
		return fmt.Errorf("process %d does not exist", pid)
	}
	
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}
	
	// Try graceful termination first (SIGTERM)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// If SIGTERM fails, force kill (SIGKILL)
		if err := process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	}
	
	return nil
}

// KillProcessGroup terminates a process group by PID
func (pm *ProcessManager) KillProcessGroup(pid int) error {
	if !pm.ProcessExists(pid) {
		return fmt.Errorf("process %d does not exist", pid)
	}
	
	// Kill the entire process group
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		// If SIGTERM fails, force kill
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to kill process group %d: %w", pid, err)
		}
	}
	
	return nil
}

// WaitForProcessExit waits for a process to exit with a timeout
func (pm *ProcessManager) WaitForProcessExit(cmd *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	
	go func() {
		done <- cmd.Wait()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		// Timeout reached, kill the process
		if cmd.Process != nil {
			pm.KillProcessGroup(cmd.Process.Pid)
		}
		return fmt.Errorf("process did not exit within timeout")
	}
}

// CreatePipes creates pipes for process communication
func (pm *ProcessManager) CreatePipes() (io.ReadCloser, io.WriteCloser, io.ReadCloser, error) {
	// Create stdin pipe
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	
	// Create stdout pipe
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		stdinReader.Close()
		stdinWriter.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	// Create stderr pipe
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		stdinReader.Close()
		stdinWriter.Close()
		stdoutReader.Close()
		stdoutWriter.Close()
		return nil, nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// For this simple implementation, we'll return stdout reader, stdin writer, stderr reader
	// Close the writer ends that the parent process doesn't need
	stdoutWriter.Close()
	stderrWriter.Close()
	stdinReader.Close()
	
	return stdoutReader, stdinWriter, stderrReader, nil
}

// StreamOutput streams output from a reader line by line
func (pm *ProcessManager) StreamOutput(reader io.Reader, outputChan chan<- string) {
	scanner := bufio.NewScanner(reader)
	
	for scanner.Scan() {
		line := scanner.Text()
		select {
		case outputChan <- line:
		default:
			// Channel full, skip this line
		}
	}
	
	close(outputChan)
}

// getProcessCmdline reads the command line of a process from /proc/PID/cmdline
func (pm *ProcessManager) getProcessCmdline(pid int) ([]string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return nil, err
	}
	
	// Command line arguments are separated by null bytes
	cmdline := strings.Split(string(data), "\x00")
	
	// Remove empty strings
	var result []string
	for _, arg := range cmdline {
		if arg != "" {
			result = append(result, arg)
		}
	}
	
	return result, nil
}

// getProcessWorkingDir reads the working directory of a process
func (pm *ProcessManager) getProcessWorkingDir(pid int) (string, error) {
	wd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return "", err
	}
	return wd, nil
}

// getProcessStatus reads the status of a process from /proc/PID/stat
func (pm *ProcessManager) getProcessStatus(pid int) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return "", err
	}
	
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return "", fmt.Errorf("invalid stat format")
	}
	
	// Field 2 is the state (3rd field, 0-indexed)
	state := fields[2]
	
	switch state {
	case "R":
		return "running", nil
	case "S":
		return "sleeping", nil
	case "D":
		return "disk_sleep", nil
	case "Z":
		return "zombie", nil
	case "T":
		return "stopped", nil
	default:
		return "unknown", nil
	}
}

// getProcessResourceUsage gets CPU and memory usage for a process
func (pm *ProcessManager) getProcessResourceUsage(pid int) (float64, int64, error) {
	// Read /proc/PID/stat for CPU info
	statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, err
	}
	
	statFields := strings.Fields(string(statData))
	if len(statFields) < 24 {
		return 0, 0, fmt.Errorf("invalid stat format")
	}
	
	// Read /proc/PID/status for memory info
	statusData, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, 0, err
	}
	
	var memoryKB int64
	statusLines := strings.Split(string(statusData), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if kb, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					memoryKB = kb
					break
				}
			}
		}
	}
	
	// Convert KB to MB
	memoryMB := memoryKB / 1024
	
	// For CPU percentage, we'd need to calculate over time
	// For now, return 0 as placeholder (would need historical data)
	cpuPercent := 0.0
	
	return cpuPercent, memoryMB, nil
}

// MonitorProcess monitors a process and returns resource usage updates
func (pm *ProcessManager) MonitorProcess(pid int, interval time.Duration) (<-chan *ProcessInfo, chan<- struct{}) {
	infoChan := make(chan *ProcessInfo, 10)
	stopChan := make(chan struct{})
	
	go func() {
		defer close(infoChan)
		
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				if !pm.ProcessExists(pid) {
					return // Process no longer exists
				}
				
				info, err := pm.GetProcessInfo(pid)
				if err != nil {
					continue // Skip this iteration on error
				}
				
				select {
				case infoChan <- info:
				default:
					// Channel full, skip this update
				}
			}
		}
	}()
	
	return infoChan, stopChan
}