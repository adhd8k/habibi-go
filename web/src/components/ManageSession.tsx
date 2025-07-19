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
  const [targetBranchName, setTargetBranchName] = useState('')
  const [isRebasing, setIsRebasing] = useState(false)
  const [isPushing, setIsPushing] = useState(false)
  const [isMerging, setIsMerging] = useState(false)
  const [isMergingToOriginal, setIsMergingToOriginal] = useState(false)

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

  const mergeMutation = useMutation({
    mutationFn: async (targetBranch?: string) => {
      if (!currentSession) return
      setIsMerging(true)
      const response = await sessionsApi.merge(currentSession.id, targetBranch)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      setIsMerging(false)
      setTargetBranchName('')
    },
    onError: () => {
      setIsMerging(false)
    },
  })

  const mergeToOriginalMutation = useMutation({
    mutationFn: async () => {
      if (!currentSession) return
      setIsMergingToOriginal(true)
      const response = await sessionsApi.mergeToOriginal(currentSession.id)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      setIsMergingToOriginal(false)
    },
    onError: () => {
      setIsMergingToOriginal(false)
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
    return <div className="p-4 text-gray-500 dark:text-gray-400">No session selected</div>
  }

  return (
    <div className="p-4 sm:p-6 max-w-2xl mx-auto overflow-y-auto">
      <h2 className="text-xl font-semibold mb-6 text-gray-900 dark:text-gray-100">Manage Session</h2>

      {/* Session Info */}
      <div className="bg-gray-50 dark:bg-gray-800 rounded-lg dark:border-r-gray-600 p-4 mb-6">
        <h3 className="font-medium mb-2 text-gray-900 dark:text-gray-100">Session Information</h3>
        <div className="space-y-1 text-sm">
          <div>
            <span className="text-gray-600 dark:text-gray-400">Name:</span>{' '}
            <span className="font-medium text-gray-900 dark:text-gray-100">{currentSession.name}</span>
          </div>
          <div>
            <span className="text-gray-600 dark:text-gray-400">Branch:</span>{' '}
            <span className="font-mono text-gray-900 dark:text-gray-100">{currentSession.branch_name}</span>
          </div>
          {currentSession.original_branch && (
            <div>
              <span className="text-gray-600 dark:text-gray-400">Original Branch:</span>{' '}
              <span className="font-mono text-gray-900 dark:text-gray-100">{currentSession.original_branch}</span>
            </div>
          )}
          <div>
            <span className="text-gray-600 dark:text-gray-400">Status:</span>{' '}
            <span className={`
              px-2 py-0.5 rounded text-xs font-medium
              ${currentSession.status === 'active' ? 'bg-green-100 dark:bg-green-800 text-green-800 dark:text-green-200' :
                'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200'}
            `}>
              {currentSession.status}
            </span>
          </div>
        </div>
      </div>

      {/* Rebase Section */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 mb-4">
        <h3 className="font-medium mb-2 text-gray-900 dark:text-gray-100">Rebase from Original Branch</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
          Pull latest changes from the original branch and rebase your work on top.
        </p>
        <button
          onClick={() => rebaseMutation.mutate()}
          disabled={isRebasing}
          className="px-4 py-2 bg-blue-500 dark:bg-blue-800 text-white rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {isRebasing ? 'Rebasing...' : 'Rebase'}
        </button>
        {rebaseMutation.isError && (
          <p className="mt-2 text-sm text-red-600 dark:text-red-400">
            Failed to rebase. You may need to resolve conflicts manually.
          </p>
        )}
        {rebaseMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">
            Successfully rebased!
          </p>
        )}
      </div>

      {/* Push Section */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 mb-4">
        <h3 className="font-medium mb-2 text-gray-900 dark:text-gray-100">Push Changes</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
          Push your branch to the remote repository.
        </p>
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1 text-gray-900 dark:text-gray-100">
              Remote Branch Name (optional)
            </label>
            <input
              type="text"
              value={pushBranchName}
              onChange={(e) => setPushBranchName(e.target.value)}
              placeholder={currentSession.branch_name}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Leave empty to push to the same branch name
            </p>
          </div>
          <button
            onClick={() => pushMutation.mutate(pushBranchName || undefined)}
            disabled={isPushing}
            className="px-4 py-2 bg-green-500 dark:bg-green-800 text-white rounded hover:bg-green-600 disabled:opacity-50"
          >
            {isPushing ? 'Pushing...' : 'Push Branch'}
          </button>
        </div>
        {pushMutation.isError && (
          <p className="mt-2 text-sm text-red-600 dark:text-red-400">
            Failed to push. Check your permissions and network connection.
          </p>
        )}
        {pushMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">
            Successfully pushed to remote!
          </p>
        )}
      </div>

      {/* Merge to Original Branch Section */}
      {currentSession.original_branch && (
        <div className="border border-blue-200 dark:border-blue-700 bg-blue-50 dark:bg-blue-900 rounded-lg p-4 mb-4">
          <h3 className="font-medium mb-2 text-blue-900 dark:text-blue-100">Merge to Original Branch</h3>
          <p className="text-sm text-blue-700 dark:text-blue-300 mb-3">
            Merge this session's changes back into the original branch: <span className="font-mono">{currentSession.original_branch}</span>
          </p>
          <button
            onClick={() => {
              if (confirm(`Are you sure you want to merge this session into ${currentSession.original_branch}? This will merge changes into the original branch.`)) {
                mergeToOriginalMutation.mutate()
              }
            }}
            disabled={isMergingToOriginal}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
          >
            {isMergingToOriginal ? 'Merging...' : `Merge to ${currentSession.original_branch}`}
          </button>
          {mergeToOriginalMutation.isError && (
            <p className="mt-2 text-sm text-red-600 dark:text-red-400">
              Failed to merge to original branch. Check for conflicts or ensure the branch exists.
            </p>
          )}
          {mergeToOriginalMutation.isSuccess && (
            <p className="mt-2 text-sm text-green-600 dark:text-green-400">
              Successfully merged into {currentSession.original_branch}!
            </p>
          )}
        </div>
      )}

      {/* Merge Section */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 mb-4">
        <h3 className="font-medium mb-2 text-gray-900 dark:text-gray-100">Merge into Branch</h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">
          Merge this session's changes into the original branch or a different target branch.
        </p>
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1 text-gray-900 dark:text-gray-100">
              Target Branch (optional)
            </label>
            <input
              type="text"
              value={targetBranchName}
              onChange={(e) => setTargetBranchName(e.target.value)}
              placeholder="main (default)"
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-blue-500 dark:focus:border-blue-400 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            />
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
              Leave empty to merge into the project's default branch
            </p>
          </div>
          <button
            onClick={() => {
              if (confirm('Are you sure you want to merge this session? This will merge changes into the target branch.')) {
                mergeMutation.mutate(targetBranchName || undefined)
              }
            }}
            disabled={isMerging}
            className="px-4 py-2 bg-purple-500 dark:bg-purple-800 text-white rounded hover:bg-purple-600 disabled:opacity-50"
          >
            {isMerging ? 'Merging...' : 'Merge Session'}
          </button>
        </div>
        {mergeMutation.isError && (
          <p className="mt-2 text-sm text-red-600 dark:text-red-400">
            Failed to merge. Check for conflicts or ensure the target branch exists.
          </p>
        )}
        {mergeMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">
            Successfully merged into target branch!
          </p>
        )}
      </div>

      {/* Close Session Section */}
      <div className="border border-red-200 dark:border-red-700 bg-red-50 dark:bg-red-900 rounded-lg p-4">
        <h3 className="font-medium mb-2 text-red-900 dark:text-red-100">Close Session</h3>
        <p className="text-sm text-red-700 dark:text-red-300 mb-3">
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
