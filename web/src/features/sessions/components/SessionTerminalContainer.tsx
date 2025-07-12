import { Terminal } from '../../../components/Terminal'

interface SessionTerminalContainerProps {
  sessionId: number
}

export function SessionTerminalContainer({ sessionId: _sessionId }: SessionTerminalContainerProps) {
  return <Terminal />
}