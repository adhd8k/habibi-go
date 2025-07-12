import { Session } from '../../../shared/types/schemas'
import { ManageSession } from '../../../components/ManageSession'

interface SessionManageContainerProps {
  session: Session
}

export function SessionManageContainer({ session: _session }: SessionManageContainerProps) {
  return <ManageSession />
}