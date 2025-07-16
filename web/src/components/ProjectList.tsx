import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { projectsApi } from '../api/client'
import { useAppStore } from '../store'
import { Project, CreateProjectRequest } from '../types'
import { AddSSHProjectForm } from './AddSSHProjectForm'
import { DropdownMenu } from './ui/DropdownMenu'
import { EditStartupScriptModal } from './EditStartupScriptModal'

export function ProjectList() {
  const queryClient = useQueryClient()
  const { currentProject, setCurrentProject, setCurrentSession } = useAppStore()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [showSSHForm, setShowSSHForm] = useState(false)
  const [newProject, setNewProject] = useState({
    name: '',
    path: '',
    setup_command: '',
  })
  const [editScriptProject, setEditScriptProject] = useState<Project | null>(null)
  
  const { data: projects, isLoading, error } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await projectsApi.list()
      // Handle the wrapped response format {data: [...], success: true}
      const data = response.data as any
      if (data && data.data && Array.isArray(data.data)) {
        return data.data
      }
      // Fallback to direct array if API format changes
      return Array.isArray(response.data) ? response.data : []
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await projectsApi.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      // Clear current project if it was deleted
      if (currentProject?.id === deleteMutation.variables) {
        setCurrentProject(null)
        setCurrentSession(null)
      }
    },
  })

  const createMutation = useMutation({
    mutationFn: async (data: CreateProjectRequest) => {
      const response = await projectsApi.create(data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      setShowCreateForm(false)
      setNewProject({ name: '', path: '', setup_command: '' })
    },
  })

  const handleSelectProject = (project: Project) => {
    setCurrentProject(project)
    setCurrentSession(null)
  }

  const handleCreateProject = () => {
    if (!newProject.name || !newProject.path) return
    
    createMutation.mutate({
      name: newProject.name,
      path: newProject.path,
      setup_command: newProject.setup_command || undefined,
    })
  }

  if (isLoading) {
    return (
      <div className="p-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 text-red-600">
        Failed to load projects
      </div>
    )
  }

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">Projects</h2>
        <div className="flex gap-2">
          <button
            onClick={() => {
              setShowCreateForm(!showCreateForm)
              setShowSSHForm(false)
            }}
            className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
          >
            New Local
          </button>
          <button
            onClick={() => {
              setShowSSHForm(!showSSHForm)
              setShowCreateForm(false)
            }}
            className="px-3 py-1 bg-purple-500 text-white rounded hover:bg-purple-600 text-sm"
          >
            New SSH
          </button>
        </div>
      </div>

      {showCreateForm && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <input
            type="text"
            placeholder="Project name"
            value={newProject.name}
            onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <div className="relative mb-2">
            <input
              type="text"
              placeholder="Project path (e.g., /home/user/myproject)"
              value={newProject.path}
              onChange={(e) => setNewProject({ ...newProject, path: e.target.value })}
              className="w-full p-2 pr-10 border rounded"
            />
            <button
              type="button"
              onClick={() => {
                // Note: Browser file API doesn't support folder path selection
                // This is just a hint to the user that they can browse
                alert('Please enter the full path to your project directory.\n\nExample:\n/home/user/projects/myproject\n\nNote: Browser security prevents automatic folder browsing.')
              }}
              className="absolute right-2 top-2 p-1 text-gray-500 hover:text-gray-700"
              title="Browse for folder"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
            </button>
          </div>
          <div className="mb-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Setup Script (optional)
            </label>
            <textarea
              placeholder="#!/bin/bash\n# This script runs when creating a new session\n# Available variables:\n# $PROJECT_PATH - Main project directory\n# $WORKTREE_PATH - Session worktree directory\n# $SESSION_NAME - Name of the session\n# $BRANCH_NAME - Git branch name\n\n# Example:\ncp $PROJECT_PATH/.env $WORKTREE_PATH/.env"
              value={newProject.setup_command}
              onChange={(e) => setNewProject({ ...newProject, setup_command: e.target.value })}
              className="w-full p-2 border rounded font-mono text-sm"
              rows={6}
            />
            <p className="text-xs text-gray-500 mt-1">
              This script will be executed each time a new session is created
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleCreateProject}
              disabled={createMutation.isPending}
              className="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
            >
              Create
            </button>
            <button
              onClick={() => {
                setShowCreateForm(false)
                setNewProject({ name: '', path: '', setup_command: '' })
              }}
              className="px-3 py-1 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {showSSHForm && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <h3 className="text-lg font-semibold mb-3">Add SSH Project</h3>
          <AddSSHProjectForm 
            onSuccess={() => {
              setShowSSHForm(false)
            }}
            onCancel={() => {
              setShowSSHForm(false)
            }}
          />
        </div>
      )}

      <div className="space-y-2">
        {projects?.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            <p className="mb-2">No projects yet</p>
            <p className="text-sm">Click "New Project" to get started</p>
          </div>
        )}
        {projects?.map((project: Project) => (
          <div key={project.id} className="relative group">
            <button
              onClick={() => handleSelectProject(project)}
              className={`w-full text-left p-3 rounded-lg transition-colors ${
                currentProject?.id === project.id
                  ? 'bg-blue-100 border-blue-500 border'
                  : 'bg-gray-50 hover:bg-gray-100 border-gray-200 border'
              }`}
            >
              <div className="pr-8">
                <div className="font-medium flex items-center gap-2">
                  {project.name}
                  {project.config?.ssh_host && (
                    <span className="text-xs bg-purple-100 text-purple-700 px-2 py-1 rounded">SSH</span>
                  )}
                </div>
                <div className="text-sm text-gray-600 truncate">
                  {project.config?.ssh_host ? (
                    <>
                      <span className="text-purple-600">{project.config.ssh_host}</span>
                      <span className="text-gray-500">:</span>
                      <span>{project.path}</span>
                    </>
                  ) : (
                    project.path
                  )}
                </div>
                {project.config?.current_branch && (
                  <div className="text-xs text-blue-600 mt-1 flex items-center">
                    <span className="mr-1">🌿</span>
                    {project.config.current_branch}
                  </div>
                )}
              </div>
            </button>
            <div className="absolute top-3 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <DropdownMenu
                items={[
                  {
                    label: 'Edit Startup Script',
                    onClick: () => setEditScriptProject(project),
                    icon: (
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    )
                  },
                  {
                    label: 'Delete Project',
                    onClick: () => {
                      if (confirm(`Are you sure you want to delete project "${project.name}"? This will also delete all sessions.`)) {
                        deleteMutation.mutate(project.id)
                      }
                    },
                    danger: true,
                    icon: (
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    )
                  }
                ]}
              />
            </div>
          </div>
        ))}
      </div>
      
      <EditStartupScriptModal
        isOpen={!!editScriptProject}
        onClose={() => setEditScriptProject(null)}
        project={editScriptProject}
      />
    </div>
  )
}