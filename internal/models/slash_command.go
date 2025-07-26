package models

// SlashCommand represents a slash command available in a session
type SlashCommand struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Path        string   `json:"path,omitempty"`
	IsBuiltin   bool     `json:"is_builtin"`
	Category    string   `json:"category,omitempty"`
	Arguments   []string `json:"arguments,omitempty"`
}

// CommandResult represents the result of executing a slash command
type CommandResult struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// CommandResultType constants
const (
	CommandResultTypeClearChat     = "clear_chat"
	CommandResultTypeShowModal     = "show_modal"
	CommandResultTypeClaudeMessage = "claude_message"
	CommandResultTypeError         = "error"
	CommandResultTypeShowHelp      = "show_help"
	CommandResultTypeStatus        = "status"
	CommandResultTypeConfig        = "config"
	CommandResultTypeInfo          = "info"
	CommandResultTypeAction        = "action"
	CommandResultTypeVimMode       = "vim_mode"
	CommandResultTypeCompact       = "compact"
)