import { Session } from '../../../shared/types/schemas'

interface SessionManageContainerProps {
  session: Session
}

export function SessionManageContainer({ session: _session }: SessionManageContainerProps) {
  // TODO: Migrate ManageSession component
  // For now, using the legacy component
  const ManageSession = require('../../../components/ManageSession').ManageSession
  return <ManageSession />
}