import { useState, useEffect } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { sessionsApi, projectsApi } from '../../../api/client'
import { CreateSessionRequest } from '../../../types'
import { Modal } from '../../../components/ui/Modal'

interface CreateSessionModalProps {
  isOpen: boolean
  onClose: () => void
  projectId: number
  onSuccess?: (session: any) => void
}

export function CreateSessionModal({ isOpen, onClose, projectId, onSuccess }: CreateSessionModalProps) {
  const queryClient = useQueryClient()
  const [newSession, setNewSession] = useState({
    session_name: '',
    branch_name: '',
    base_branch: 'main',
  })
  const [showBranchSuggestions, setShowBranchSuggestions] = useState(false)

  const { data: branches } = useQuery({
    queryKey: ['branches', projectId],
    queryFn: async () => {
      const response = await projectsApi.getBranches(projectId)
      const data = response.data as any
      if (data && data.data) {
        return data.data
      }
      return response.data
    },
    enabled: isOpen && !!projectId,
  })

  const createMutation = useMutation({
    mutationFn: async (data: CreateSessionRequest) => {
      const response = await sessionsApi.create(data)
      return response.data
    },
    onSuccess: (response: any) => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      handleClose()
      
      // Extract session data from response
      const session = (response as any).data || response
      onSuccess?.(session)
    },
  })

  const handleCreateSession = () => {
    if (!newSession.session_name || !newSession.branch_name) return

    createMutation.mutate({
      project_id: projectId,
      name: newSession.session_name,
      branch_name: newSession.branch_name,
      base_branch: newSession.base_branch,
    })
  }

  const handleClose = () => {
    setNewSession({ session_name: '', branch_name: '', base_branch: 'main' })
    setShowBranchSuggestions(false)
    onClose()
  }

  // Reset form when modal closes
  useEffect(() => {
    if (!isOpen) {
      setNewSession({ session_name: '', branch_name: '', base_branch: 'main' })
      setShowBranchSuggestions(false)
    }
  }, [isOpen])

  const footer = (
    <>
      <button
        onClick={handleClose}
        className="px-3 py-1 bg-gray-300 dark:bg-gray-600 text-gray-700 dark:text-gray-200 rounded hover:bg-gray-400 dark:hover:bg-gray-500 text-sm"
      >
        Cancel
      </button>
      <button
        onClick={handleCreateSession}
        disabled={createMutation.isPending || !newSession.session_name || !newSession.branch_name}
        className="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
      >
        {createMutation.isPending ? 'Creating...' : 'Create'}
      </button>
    </>
  )

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Create New Session" footer={footer}>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Session Name
          </label>
          <input
            type="text"
            placeholder="Enter session name"
            value={newSession.session_name}
            onChange={(e) => setNewSession({ ...newSession, session_name: e.target.value })}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Branch Name
          </label>
          <input
            type="text"
            placeholder="Enter new branch name"
            value={newSession.branch_name}
            onChange={(e) => setNewSession({ ...newSession, branch_name: e.target.value })}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            This will create a new Git branch for your session
          </p>
        </div>

        <div className="relative">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Base Branch
          </label>
          <input
            type="text"
            placeholder="Base branch (e.g., main)"
            value={newSession.base_branch}
            onChange={(e) => setNewSession({ ...newSession, base_branch: e.target.value })}
            onFocus={() => setShowBranchSuggestions(true)}
            onBlur={() => setTimeout(() => setShowBranchSuggestions(false), 200)}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
          {showBranchSuggestions && branches && (
            <div className="absolute z-10 w-full mt-1 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded shadow-lg max-h-48 overflow-y-auto">
              {branches.local.length > 0 && (
                <>
                  <div className="px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-700">
                    Local Branches
                  </div>
                  {branches.local.map((branch: string) => (
                    <button
                      key={`local-${branch}`}
                      type="button"
                      onClick={() => {
                        setNewSession({ ...newSession, base_branch: branch })
                        setShowBranchSuggestions(false)
                      }}
                      className="w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2 text-gray-900 dark:text-gray-100"
                    >
                      <span className="text-blue-600 dark:text-blue-400">●</span>
                      {branch}
                    </button>
                  ))}
                </>
              )}
              {branches.remote.length > 0 && (
                <>
                  <div className="px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-700">
                    Remote Branches
                  </div>
                  {branches.remote.map((branch: string) => (
                    <button
                      key={`remote-${branch}`}
                      type="button"
                      onClick={() => {
                        setNewSession({ ...newSession, base_branch: branch })
                        setShowBranchSuggestions(false)
                      }}
                      className="w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2 text-gray-900 dark:text-gray-100"
                    >
                      <span className="text-green-600 dark:text-green-400">●</span>
                      {branch}
                    </button>
                  ))}
                </>
              )}
              {branches.local.length === 0 && branches.remote.length === 0 && (
                <div className="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
                  No branches found
                </div>
              )}
            </div>
          )}
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            The branch to create your new branch from
          </p>
        </div>
      </div>
    </Modal>
  )
}