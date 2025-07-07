import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useAppStore } from '../store'
import { sessionsApi } from '../api/client'

export function ManageSession() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { currentSession, setCurrentSession } = useAppStore()
  const [pushBranchName, setPushBranchName] = useState('')
  const [isRebasing, setIsRebasing] = useState(false)
  const [isPushing, setIsPushing] = useState(false)

  const rebaseMutation = useMutation({
    mutationFn: async () => {
      if (!currentSession) return
      setIsRebasing(true)
      const response = await sessionsApi.rebase(currentSession.id)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      setIsRebasing(false)
    },
    onError: () => {
      setIsRebasing(false)
    },
  })

  const pushMutation = useMutation({
    mutationFn: async (remoteBranch?: string) => {
      if (!currentSession) return
      setIsPushing(true)
      const response = await sessionsApi.push(currentSession.id, remoteBranch)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      setIsPushing(false)
      setPushBranchName('')
    },
    onError: () => {
      setIsPushing(false)
    },
  })

  const closeMutation = useMutation({
    mutationFn: async () => {
      if (!currentSession) return
      const response = await sessionsApi.close(currentSession.id)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      setCurrentSession(null)
      navigate('/')
    },
  })

  if (!currentSession) {
    return <div className="p-4 text-gray-500">No session selected</div>
  }

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <h2 className="text-xl font-semibold mb-6">Manage Session</h2>

      {/* Session Info */}
      <div className="bg-gray-50 rounded-lg p-4 mb-6">
        <h3 className="font-medium mb-2">Session Information</h3>
        <div className="space-y-1 text-sm">
          <div>
            <span className="text-gray-600">Name:</span>{' '}
            <span className="font-medium">{currentSession.name}</span>
          </div>
          <div>
            <span className="text-gray-600">Branch:</span>{' '}
            <span className="font-mono">{currentSession.branch_name}</span>
          </div>
          <div>
            <span className="text-gray-600">Status:</span>{' '}
            <span className={`
              px-2 py-0.5 rounded text-xs font-medium
              ${currentSession.status === 'active' ? 'bg-green-100 text-green-800' :
                'bg-gray-100 text-gray-800'}
            `}>
              {currentSession.status}
            </span>
          </div>
        </div>
      </div>

      {/* Rebase Section */}
      <div className="border rounded-lg p-4 mb-4">
        <h3 className="font-medium mb-2">Rebase from Original Branch</h3>
        <p className="text-sm text-gray-600 mb-3">
          Pull latest changes from the original branch and rebase your work on top.
        </p>
        <button
          onClick={() => rebaseMutation.mutate()}
          disabled={isRebasing}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {isRebasing ? 'Rebasing...' : 'Rebase'}
        </button>
        {rebaseMutation.isError && (
          <p className="mt-2 text-sm text-red-600">
            Failed to rebase. You may need to resolve conflicts manually.
          </p>
        )}
        {rebaseMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600">
            Successfully rebased!
          </p>
        )}
      </div>

      {/* Push Section */}
      <div className="border rounded-lg p-4 mb-4">
        <h3 className="font-medium mb-2">Push Changes</h3>
        <p className="text-sm text-gray-600 mb-3">
          Push your branch to the remote repository.
        </p>
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1">
              Remote Branch Name (optional)
            </label>
            <input
              type="text"
              value={pushBranchName}
              onChange={(e) => setPushBranchName(e.target.value)}
              placeholder={currentSession.branch_name}
              className="w-full px-3 py-2 border rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            <p className="text-xs text-gray-500 mt-1">
              Leave empty to push to the same branch name
            </p>
          </div>
          <button
            onClick={() => pushMutation.mutate(pushBranchName || undefined)}
            disabled={isPushing}
            className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
          >
            {isPushing ? 'Pushing...' : 'Push Branch'}
          </button>
        </div>
        {pushMutation.isError && (
          <p className="mt-2 text-sm text-red-600">
            Failed to push. Check your permissions and network connection.
          </p>
        )}
        {pushMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600">
            Successfully pushed to remote!
          </p>
        )}
      </div>

      {/* Close Session Section */}
      <div className="border border-red-200 bg-red-50 rounded-lg p-4">
        <h3 className="font-medium mb-2 text-red-900">Close Session</h3>
        <p className="text-sm text-red-700 mb-3">
          This will stop all agents and remove the worktree. Make sure to push any changes you want to keep.
        </p>
        <button
          onClick={() => {
            if (confirm('Are you sure you want to close this session? This will remove the worktree.')) {
              closeMutation.mutate()
            }
          }}
          disabled={closeMutation.isPending}
          className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
        >
          {closeMutation.isPending ? 'Closing...' : 'Close Session'}
        </button>
      </div>
    </div>
  )
}