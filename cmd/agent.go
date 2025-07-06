package cmd

import (
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent management commands",
	Long:  `Start, stop, and monitor coding agents.`,
}

var agentListCmd = &cobra.Command{
	Use:   "list [session]",
	Short: "List agents in a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement agent list
	},
}

var agentStartCmd = &cobra.Command{
	Use:   "start [session] [agent-type]",
	Short: "Start a new agent",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement agent start
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentStartCmd)
}