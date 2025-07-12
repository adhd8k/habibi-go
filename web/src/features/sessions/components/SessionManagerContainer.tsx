import { useState } from 'react'
import { useAppSelector } from '../../../app/hooks'
import { selectCurrentProject } from '../../projects/slice/projectsSlice'
import { useGetSessionsQuery } from '../api/sessionsApi'
import { SessionManagerView } from './SessionManagerView'
import { CreateSessionModal } from './CreateSessionModal'

export function SessionManagerContainer() {
  const currentProject = useAppSelector(selectCurrentProject)
  const [showCreateModal, setShowCreateModal] = useState(false)
  
  const { data: sessions, isLoading, error } = useGetSessionsQuery(
    currentProject?.id,
    { skip: !currentProject }
  )

  const handleCreateSession = () => {
    setShowCreateModal(true)
  }

  const handleCloseModal = () => {
    setShowCreateModal(false)
  }

  return (
    <>
      <SessionManagerView
        sessions={sessions || []}
        currentProject={currentProject}
        isLoading={isLoading}
        error={error}
        onCreateSession={handleCreateSession}
      />
      {showCreateModal && currentProject && (
        <CreateSessionModal
          projectId={currentProject.id}
          defaultBranch={currentProject.default_branch}
          onClose={handleCloseModal}
        />
      )}
    </>
  )
}