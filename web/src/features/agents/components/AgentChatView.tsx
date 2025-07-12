import { Agent } from '../../../shared/types/schemas'

interface AgentChatViewProps {
  agent: Agent
}

export function AgentChatView({ agent: _agent }: AgentChatViewProps) {
  // TODO: Migrate AgentControl/ClaudeChat component
  // For now, using the legacy component
  const AgentControl = require('../../../components/AgentControl').AgentControl
  return <AgentControl />
}