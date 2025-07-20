import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { projectsApi } from '../api/client'
import { CreateProjectRequest } from '../types'
import { Modal } from './ui/Modal'

interface CreateProjectModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess?: () => void
}

export function CreateProjectModal({ isOpen, onClose, onSuccess }: CreateProjectModalProps) {
  const queryClient = useQueryClient()
  const [newProject, setNewProject] = useState({
    name: '',
    path: '',
    setup_command: '',
  })

  const createMutation = useMutation({
    mutationFn: async (data: CreateProjectRequest) => {
      const response = await projectsApi.create(data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      handleClose()
      onSuccess?.()
    },
  })

  const handleCreateProject = () => {
    if (!newProject.name || !newProject.path) return

    createMutation.mutate({
      name: newProject.name,
      path: newProject.path,
      setup_command: newProject.setup_command || undefined,
    })
  }

  const handleClose = () => {
    setNewProject({ name: '', path: '', setup_command: '' })
    onClose()
  }

  const footer = (
    <>
      <button
        onClick={handleClose}
        className="px-3 py-1 bg-gray-300 dark:bg-gray-600 text-gray-700 dark:text-gray-200 rounded hover:bg-gray-400 dark:hover:bg-gray-500 text-sm"
      >
        Cancel
      </button>
      <button
        onClick={handleCreateProject}
        disabled={createMutation.isPending || !newProject.name || !newProject.path}
        className="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
      >
        {createMutation.isPending ? 'Creating...' : 'Create'}
      </button>
    </>
  )

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Create Local Project" footer={footer}>
      <div className="space-y-4">
        <div>
          <input
            type="text"
            placeholder="Project name"
            value={newProject.name}
            onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>

        <div className="relative">
          <input
            type="text"
            placeholder="Project path (e.g., /home/user/myproject)"
            value={newProject.path}
            onChange={(e) => setNewProject({ ...newProject, path: e.target.value })}
            className="w-full p-2 pr-10 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
          <button
            type="button"
            onClick={() => {
              // Note: Browser file API doesn't support folder path selection
              // This is just a hint to the user that they can browse
              alert('Please enter the full path to your project directory.\n\nExample:\n/home/user/projects/myproject\n\nNote: Browser security prevents automatic folder browsing.')
            }}
            className="absolute right-2 top-2 p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            title="Browse for folder"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </button>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Setup Script (optional)
          </label>
          <textarea
            placeholder="#!/bin/bash
# This script runs when creating a new session
# Available variables:
# $PROJECT_PATH - Main project directory
# $WORKTREE_PATH - Session worktree directory
# $SESSION_NAME - Name of the session
# $BRANCH_NAME - Git branch name

# Example:
cp $PROJECT_PATH/.env $WORKTREE_PATH/.env"
            value={newProject.setup_command}
            onChange={(e) => setNewProject({ ...newProject, setup_command: e.target.value })}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded font-mono text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
            rows={6}
          />
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
            This script will be executed each time a new session is created
          </p>
        </div>
      </div>
    </Modal>
  )
}