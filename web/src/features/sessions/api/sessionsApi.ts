import { baseApi, ApiResponse } from '../../../services/api/baseApi'
import { 
  Session, 
  CreateSessionRequest, 
  SessionSchema,
  CreateSessionRequestSchema 
} from '../../../shared/types/schemas'

export const sessionsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getSessions: builder.query<Session[], number | void>({
      query: (projectId) => projectId 
        ? `/sessions?project_id=${projectId}` 
        : '/sessions',
      transformResponse: (response: ApiResponse<Session[]>) => {
        if (response.success && response.data) {
          return response.data.map(session => SessionSchema.parse(session))
        }
        return []
      },
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Session' as const, id })),
              { type: 'Session', id: 'LIST' },
            ]
          : [{ type: 'Session', id: 'LIST' }],
    }),

    getSession: builder.query<Session, number>({
      query: (id) => `/sessions/${id}`,
      transformResponse: (response: ApiResponse<Session>) => {
        if (response.success && response.data) {
          return SessionSchema.parse(response.data)
        }
        throw new Error(response.error || 'Session not found')
      },
      providesTags: (_result, _error, id) => [{ type: 'Session', id }],
    }),

    createSession: builder.mutation<Session, CreateSessionRequest>({
      query: (session) => ({
        url: '/sessions',
        method: 'POST',
        body: CreateSessionRequestSchema.parse(session),
      }),
      transformResponse: (response: ApiResponse<Session>) => {
        if (response.success && response.data) {
          return SessionSchema.parse(response.data)
        }
        throw new Error(response.error || 'Failed to create session')
      },
      invalidatesTags: [
        { type: 'Session', id: 'LIST' },
        { type: 'Project', id: 'LIST' }, // Projects show session count
      ],
    }),

    updateSession: builder.mutation<Session, { id: number; data: Partial<Session> }>({
      query: ({ id, data }) => ({
        url: `/sessions/${id}`,
        method: 'PUT',
        body: data,
      }),
      transformResponse: (response: ApiResponse<Session>) => {
        if (response.success && response.data) {
          return SessionSchema.parse(response.data)
        }
        throw new Error(response.error || 'Failed to update session')
      },
      invalidatesTags: (_result, _error, { id }) => [
        { type: 'Session', id },
        { type: 'Session', id: 'LIST' },
      ],
    }),

    deleteSession: builder.mutation<void, number>({
      query: (id) => ({
        url: `/sessions/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, id) => [
        { type: 'Session', id },
        { type: 'Session', id: 'LIST' },
        { type: 'Project', id: 'LIST' },
      ],
    }),

    getSessionDiffs: builder.query<any, number>({
      query: (id) => `/sessions/${id}/diffs`,
      transformResponse: (response: ApiResponse<any>) => {
        if (response.success && response.data) {
          return response.data
        }
        return { files: [] }
      },
    }),

    rebaseSession: builder.mutation<void, number>({
      query: (id) => ({
        url: `/sessions/${id}/rebase`,
        method: 'POST',
      }),
      invalidatesTags: (_result, _error, id) => [{ type: 'Session', id }],
    }),

    pushSession: builder.mutation<void, { id: number; remoteBranch?: string }>({
      query: ({ id, remoteBranch }) => ({
        url: `/sessions/${id}/push`,
        method: 'POST',
        body: { remote_branch: remoteBranch },
      }),
      invalidatesTags: (_result, _error, { id }) => [{ type: 'Session', id }],
    }),

    mergeSession: builder.mutation<void, { id: number; targetBranch?: string }>({
      query: ({ id, targetBranch }) => ({
        url: `/sessions/${id}/merge`,
        method: 'POST',
        body: { target_branch: targetBranch },
      }),
      invalidatesTags: (_result, _error, { id }) => [{ type: 'Session', id }],
    }),

    mergeSessionToOriginal: builder.mutation<void, number>({
      query: (id) => ({
        url: `/sessions/${id}/merge-to-original`,
        method: 'POST',
      }),
      invalidatesTags: (_result, _error, id) => [{ type: 'Session', id }],
    }),

    closeSession: builder.mutation<void, number>({
      query: (id) => ({
        url: `/sessions/${id}/close`,
        method: 'POST',
      }),
      invalidatesTags: (_result, _error, id) => [
        { type: 'Session', id },
        { type: 'Session', id: 'LIST' },
      ],
    }),

    openWithEditor: builder.mutation<void, number>({
      query: (id) => ({
        url: `/sessions/${id}/open-editor`,
        method: 'POST',
      }),
      // No cache invalidation needed for this action
    }),
  }),
})

export const {
  useGetSessionsQuery,
  useGetSessionQuery,
  useCreateSessionMutation,
  useUpdateSessionMutation,
  useDeleteSessionMutation,
  useGetSessionDiffsQuery,
  useRebaseSessionMutation,
  usePushSessionMutation,
  useMergeSessionMutation,
  useMergeSessionToOriginalMutation,
  useCloseSessionMutation,
  useOpenWithEditorMutation,
} = sessionsApi