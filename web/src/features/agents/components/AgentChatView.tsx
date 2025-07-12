import { Agent, ChatMessage } from '../../../shared/types/schemas'
import { AgentControl } from '../../../components/AgentControl'

interface AgentChatViewProps {
  agent: Agent | null
  messages: ChatMessage[]
  onSendMessage: (message: string) => void
}

export function AgentChatView({ agent: _agent, messages: _messages, onSendMessage: _onSendMessage }: AgentChatViewProps) {
  // TODO: Pass props to AgentControl when it's migrated
  // For now, the legacy component will use its own state
  return <AgentControl />
}