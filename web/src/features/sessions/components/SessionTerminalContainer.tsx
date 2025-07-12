interface SessionTerminalContainerProps {
  sessionId: number
}

export function SessionTerminalContainer({ sessionId: _sessionId }: SessionTerminalContainerProps) {
  // TODO: Migrate Terminal component
  // For now, using the legacy component
  const Terminal = require('../../../components/Terminal').Terminal
  return <Terminal />
}