import { useQuery } from '@tanstack/react-query'
import { projectsApi } from '../api/client'
import { useAppStore } from '../store'
import { Project } from '../types'

export function ProjectList() {
  const { currentProject, setCurrentProject, setCurrentSession } = useAppStore()
  
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

  const handleSelectProject = (project: Project) => {
    setCurrentProject(project)
    setCurrentSession(null)
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
      <h2 className="text-lg font-semibold mb-4">Projects</h2>
      <div className="space-y-2">
        {projects?.map((project: Project) => (
          <button
            key={project.id}
            onClick={() => handleSelectProject(project)}
            className={`w-full text-left p-3 rounded-lg transition-colors ${
              currentProject?.id === project.id
                ? 'bg-blue-100 border-blue-500 border'
                : 'bg-gray-50 hover:bg-gray-100 border-gray-200 border'
            }`}
          >
            <div className="font-medium">{project.name}</div>
            <div className="text-sm text-gray-600">{project.path}</div>
            {project.repository_url && (
              <div className="text-xs text-gray-500 mt-1">
                {project.repository_url}
              </div>
            )}
          </button>
        ))}
      </div>
    </div>
  )
}