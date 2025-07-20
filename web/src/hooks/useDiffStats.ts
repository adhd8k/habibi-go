import { useQuery, useQueryClient } from '@tanstack/react-query'
import { sessionsApi, type DiffFile } from '../api/client'

export function useDiffStats(sessionId?: number) {
  const queryClient = useQueryClient()
  
  return useQuery({
    queryKey: ['session-diff-stats', sessionId],
    queryFn: async () => {
      if (!sessionId) return null
      
      // Try to get cached data from FileDiffs first
      const cachedDiffData = queryClient.getQueryData(['session-diffs', sessionId])
      let files: DiffFile[] = []
      
      if (cachedDiffData) {
        // Use cached data if available
        files = (cachedDiffData as any)?.data?.files || []
      } else {
        // Fetch fresh data if not cached
        const response = await sessionsApi.getDiffs(sessionId)
        files = response.data?.data?.files || []
      }
      
      const stats = {
        filesChanged: files.length,
        additions: files.reduce((sum: number, f: DiffFile) => sum + (f.additions || 0), 0),
        deletions: files.reduce((sum: number, f: DiffFile) => sum + (f.deletions || 0), 0),
      }
      
      return stats
    },
    enabled: !!sessionId,
  })
}