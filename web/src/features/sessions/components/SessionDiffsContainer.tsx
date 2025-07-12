import { useGetSessionDiffsQuery } from '../api/sessionsApi'
import { LoadingSpinner } from '../../../shared/components/LoadingSpinner'
import { ErrorMessage } from '../../../shared/components/ErrorMessage'

interface SessionDiffsContainerProps {
  sessionId: number
}

export function SessionDiffsContainer({ sessionId }: SessionDiffsContainerProps) {
  const { data: _diffs, isLoading, error } = useGetSessionDiffsQuery(sessionId)

  if (isLoading) {
    return <LoadingSpinner className="mt-8" />
  }

  if (error) {
    return (
      <ErrorMessage
        message="Failed to load diffs"
        className="m-4"
      />
    )
  }

  // TODO: Migrate FileDiffs component
  // For now, using the legacy component
  const FileDiffs = require('../../../components/FileDiffs').FileDiffs
  return <FileDiffs />
}