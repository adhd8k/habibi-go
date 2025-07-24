import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { projectsApi, sessionsApi } from '../../../api/client'
import { useAppStore } from '../../../store'
import { Project } from '../../../types'
import { DropdownMenu } from '../../../components/ui/DropdownMenu'
import { CreateProjectModal } from './CreateProjectModal'
import { CreateSSHProjectModal } from './CreateSSHProjectModal'
import { EditProjectModal } from './EditProjectModal'
import { EditStartupScriptModal } from './EditStartupScriptModal'

export function ProjectManager() {
  const queryClient = useQueryClient()
  const { currentProject, setCurrentProject, setCurrentSession } = useAppStore()
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showSSHModal, setShowSSHModal] = useState(false)
  const [editProject, setEditProject] = useState<Project | null>(null)
  const [editScriptProject, setEditScriptProject] = useState<Project | null>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const [isCollapsed, setIsCollapsed] = useState(false)
  const itemsPerPage = 5

  const { data: projects, isLoading } = useQuery({
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

  // Fetch session counts for all projects
  const { data: sessionCounts } = useQuery({
    queryKey: ['session-counts', projects?.map((p: Project) => p.id)],
    queryFn: async () => {
      if (!projects || projects.length === 0) return {}
      
      const counts: Record<number, number> = {}
      
      // Fetch sessions for each project in parallel
      const sessionPromises = projects.map(async (project: Project) => {
        try {
          const response = await sessionsApi.list(project.id)
          const data = response.data as any
          let sessions = []
          if (data && data.data && Array.isArray(data.data)) {
            sessions = data.data
          } else if (Array.isArray(response.data)) {
            sessions = response.data
          }
          counts[project.id] = sessions.length
        } catch (error) {
          console.error(`Failed to fetch sessions for project ${project.id}:`, error)
          counts[project.id] = 0
        }
      })
      
      await Promise.all(sessionPromises)
      return counts
    },
    enabled: !!projects && projects.length > 0,
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

  const handleProjectSelect = (project: Project) => {
    setCurrentProject(project)
    // Clear current session when switching projects
    setCurrentSession(null)
  }

  const handleCreateSuccess = () => {
    // The modals will handle invalidating the project list
    // We could optionally set the newly created project as current here
  }

  // Calculate pagination
  const displayProjects = isCollapsed && currentProject 
    ? [currentProject] 
    : (projects || [])
  
  const totalPages = Math.ceil(displayProjects.length / itemsPerPage)
  const startIndex = (currentPage - 1) * itemsPerPage
  const endIndex = startIndex + itemsPerPage
  const currentProjects = displayProjects.slice(startIndex, endIndex)

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Projects</h2>
          {currentProject && (
            <button
              onClick={() => setIsCollapsed(!isCollapsed)}
              className="p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              title={isCollapsed ? "Show all projects" : "Show only active project"}
            >
              <svg className={`w-4 h-4 transition-transform ${isCollapsed ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>
          )}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
          >
            New Local
          </button>
          <button
            onClick={() => setShowSSHModal(true)}
            className="px-3 py-1 bg-purple-500 text-white rounded hover:bg-purple-600 text-sm"
          >
            New SSH
          </button>
        </div>
      </div>


      {isLoading ? (
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
        </div>
      ) : (
        <div className="space-y-2">
          {currentProjects.map((project: Project) => (
            <div key={project.id} className="relative group">
              <button
                onClick={() => handleProjectSelect(project)}
                className={`w-full text-left p-3 rounded-lg transition-colors ${
                  currentProject?.id === project.id
                    ? 'bg-blue-100 dark:bg-blue-900 border-blue-500 border'
                    : 'bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 border-gray-200 dark:border-gray-600 border'
                }`}
              >
                <div className="pr-8">
                  <div className="font-medium flex items-center gap-2 text-gray-900 dark:text-gray-100">
                    {project.name}
                    {project.config?.ssh_host && (
                      <span className="text-xs bg-purple-100 dark:bg-purple-800 text-purple-700 dark:text-purple-200 px-2 py-1 rounded">SSH</span>
                    )}
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-300 truncate">
                    {project.config?.ssh_host ? (
                      <>
                        <span className="text-purple-600 dark:text-purple-400">{project.config.ssh_host}</span>
                        <span className="text-gray-500 dark:text-gray-400">:</span>
                        <span>{project.path}</span>
                      </>
                    ) : (
                      project.path
                    )}
                  </div>
                  {project.config?.current_branch && (
                    <div className="text-xs mt-1 flex items-center justify-between">
                      <div className="flex items-center text-blue-600 dark:text-blue-400">
                        <span className="mr-1">🌿</span>
                        {project.config.current_branch}
                      </div>
                      {sessionCounts && (sessionCounts[project.id] || 0) > 0 && (
                        <div className="flex items-center text-gray-500 dark:text-gray-400">
                          <span className="w-2 h-2 bg-green-500 rounded-full mr-1"></span>
                          {sessionCounts[project.id]} session{sessionCounts[project.id] !== 1 ? 's' : ''}
                        </div>
                      )}
                    </div>
                  )}
                  {!project.config?.current_branch && sessionCounts && (sessionCounts[project.id] || 0) > 0 && (
                    <div className="text-xs text-gray-500 dark:text-gray-400 mt-1 flex items-center">
                      <span className="w-2 h-2 bg-green-500 rounded-full mr-1"></span>
                      {sessionCounts[project.id]} session{sessionCounts[project.id] !== 1 ? 's' : ''}
                    </div>
                  )}
                </div>
              </button>
              <div className="absolute top-3 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <DropdownMenu
                  items={[
                    {
                      label: 'Edit Project',
                      onClick: () => setEditProject(project),
                      icon: (
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      )
                    },
                    {
                      label: 'Edit Startup Script',
                      onClick: () => setEditScriptProject(project),
                      icon: (
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                      )
                    },
                    {
                      label: 'Delete Project',
                      onClick: () => {
                        if (confirm(`Are you sure you want to delete project "${project.name}"?`)) {
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
      )}

      {/* Pagination Controls */}
      {!isCollapsed && totalPages > 1 && (
        <div className="flex justify-center items-center gap-2 mt-4">
          <button
            onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
            disabled={currentPage === 1}
            className="px-3 py-1 text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Page {currentPage} of {totalPages}
          </span>
          <button
            onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
            disabled={currentPage === totalPages}
            className="px-3 py-1 text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
      )}

      <CreateProjectModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={handleCreateSuccess}
      />

      <CreateSSHProjectModal
        isOpen={showSSHModal}
        onClose={() => setShowSSHModal(false)}
        onSuccess={handleCreateSuccess}
      />

      <EditProjectModal
        isOpen={!!editProject}
        onClose={() => setEditProject(null)}
        project={editProject}
      />

      <EditStartupScriptModal
        isOpen={!!editScriptProject}
        onClose={() => setEditScriptProject(null)}
        project={editScriptProject}
      />
    </div>
  )
}