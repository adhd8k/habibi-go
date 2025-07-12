import { baseApi, ApiResponse } from '../../../services/api/baseApi'
import { 
  Project, 
  CreateProjectRequest, 
  ProjectSchema, 
  CreateProjectRequestSchema 
} from '../../../shared/types/schemas'

export const projectsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getProjects: builder.query<Project[], void>({
      query: () => '/projects',
      transformResponse: (response: ApiResponse<Project[]>) => {
        // Validate response data
        if (response.data) {
          return response.data.map(project => ProjectSchema.parse(project))
        }
        return []
      },
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Project' as const, id })),
              { type: 'Project', id: 'LIST' },
            ]
          : [{ type: 'Project', id: 'LIST' }],
    }),

    getProject: builder.query<Project, number>({
      query: (id) => `/projects/${id}`,
      transformResponse: (response: ApiResponse<Project>) => {
        if (response.data) {
          return ProjectSchema.parse(response.data)
        }
        throw new Error('Project not found')
      },
      providesTags: (_result, _error, id) => [{ type: 'Project', id }],
    }),

    createProject: builder.mutation<Project, CreateProjectRequest>({
      query: (project) => ({
        url: '/projects',
        method: 'POST',
        body: CreateProjectRequestSchema.parse(project),
      }),
      transformResponse: (response: ApiResponse<Project>) => {
        if (response.data) {
          return ProjectSchema.parse(response.data)
        }
        throw new Error('Failed to create project')
      },
      invalidatesTags: [{ type: 'Project', id: 'LIST' }],
    }),

    updateProject: builder.mutation<Project, { id: number; data: Partial<Project> }>({
      query: ({ id, data }) => ({
        url: `/projects/${id}`,
        method: 'PUT',
        body: data,
      }),
      transformResponse: (response: ApiResponse<Project>) => {
        if (response.data) {
          return ProjectSchema.parse(response.data)
        }
        throw new Error('Failed to update project')
      },
      invalidatesTags: (_result, _error, { id }) => [
        { type: 'Project', id },
        { type: 'Project', id: 'LIST' },
      ],
    }),

    deleteProject: builder.mutation<void, number>({
      query: (id) => ({
        url: `/projects/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, id) => [
        { type: 'Project', id },
        { type: 'Project', id: 'LIST' },
      ],
    }),

    discoverProjects: builder.mutation<Project[], { path?: string }>({
      query: (params) => ({
        url: '/projects/discover',
        method: 'POST',
        body: params,
      }),
      transformResponse: (response: ApiResponse<Project[]>) => {
        if (response.data) {
          return response.data.map(project => ProjectSchema.parse(project))
        }
        return []
      },
    }),

    getProjectStats: builder.query<Record<string, any>, void>({
      query: () => '/projects/stats',
      transformResponse: (response: ApiResponse<Record<string, any>>) => {
        return response.data || {}
      },
    }),
  }),
})

export const {
  useGetProjectsQuery,
  useGetProjectQuery,
  useCreateProjectMutation,
  useUpdateProjectMutation,
  useDeleteProjectMutation,
  useDiscoverProjectsMutation,
  useGetProjectStatsQuery,
} = projectsApi