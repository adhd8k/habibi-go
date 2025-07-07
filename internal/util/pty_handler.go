package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"github.com/creack/pty"
)

// StartProcessWithPTY starts a process with a pseudo-terminal
func StartProcessWithPTY(cmd *exec.Cmd) (*PTYProcess, error) {
	// Start the command with a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}

	// Set the PTY size to a reasonable default
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		fmt.Printf("Warning: failed to set PTY size: %v\n", err)
	}

	return &PTYProcess{
		PTY:     ptmx,
		Cmd:     cmd,
		InputCh: make(chan string, 10),
		OutputCh: make(chan string, 100),
		ErrorCh: make(chan error, 1),
		StopCh:  make(chan struct{}),
	}, nil
}

// PTYProcess represents a process running with a PTY
type PTYProcess struct {
	PTY      *os.File
	Cmd      *exec.Cmd
	InputCh  chan string
	OutputCh chan string
	ErrorCh  chan error
	StopCh   chan struct{}
}

// Start begins handling I/O for the PTY process
func (p *PTYProcess) Start() {
	// Handle input
	go func() {
		for {
			select {
			case input := <-p.InputCh:
				if _, err := p.PTY.Write([]byte(input + "\n")); err != nil {
					select {
					case p.ErrorCh <- fmt.Errorf("failed to write to pty: %w", err):
					default:
					}
					return
				}
			case <-p.StopCh:
				return
			}
		}
	}()

	// Handle output
	go func() {
		scanner := bufio.NewScanner(p.PTY)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case p.OutputCh <- line:
			default:
				// Channel full, skip
			}
		}
		
		if err := scanner.Err(); err != nil && err != io.EOF {
			select {
			case p.ErrorCh <- fmt.Errorf("pty read error: %w", err):
			default:
			}
		}
		
		close(p.OutputCh)
	}()
}

// Stop stops the PTY process
func (p *PTYProcess) Stop() error {
	close(p.StopCh)
	p.PTY.Close()
	
	if p.Cmd.Process != nil {
		return p.Cmd.Process.Kill()
	}
	
	return nil
}