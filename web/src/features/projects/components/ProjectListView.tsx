import { useState } from 'react'
import { Project } from '../../../shared/types/schemas'
import { CreateProjectForm } from './CreateProjectForm'

interface ProjectListViewProps {
  projects: Project[]
  currentProject: Project | null
  isLoading: boolean
  error: any
  isDeleting: boolean
  onSelectProject: (project: Project) => void
  onDeleteProject: (id: number) => void
}

export function ProjectListView({
  projects,
  currentProject,
  isLoading,
  error,
  isDeleting,
  onSelectProject,
  onDeleteProject,
}: ProjectListViewProps) {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [showSSHForm, setShowSSHForm] = useState(false)

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
          <CreateProjectForm 
            onSuccess={() => setShowCreateForm(false)}
            onCancel={() => setShowCreateForm(false)}
          />
        </div>
      )}

      {showSSHForm && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <h3 className="text-lg font-semibold mb-3">Add SSH Project</h3>
          {/* TODO: Add SSH form component */}
          <p className="text-gray-500">SSH project form coming soon...</p>
        </div>
      )}

      <div className="space-y-2">
        {projects.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            <p className="mb-2">No projects yet</p>
            <p className="text-sm">Click "New Project" to get started</p>
          </div>
        )}
        
        {projects.map((project) => (
          <div key={project.id} className="relative group">
            <button
              onClick={() => onSelectProject(project)}
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
            <button
              onClick={(e) => {
                e.stopPropagation()
                onDeleteProject(project.id)
              }}
              disabled={isDeleting}
              className="absolute top-3 right-2 p-1 text-red-500 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-50 rounded disabled:opacity-50"
              title="Delete project"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}