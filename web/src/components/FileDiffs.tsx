import { useQuery } from '@tanstack/react-query'
import { useAppStore } from '../store'
import { sessionsApi, type DiffFile } from '../api/client'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'

export function FileDiffs() {
  const { currentSession } = useAppStore()

  const { data: diffsData, isLoading, error, refetch } = useQuery({
    queryKey: ['session-diffs', currentSession?.id],
    queryFn: async () => {
      if (!currentSession) return null
      const response = await sessionsApi.getDiffs(currentSession.id)
      console.log('Full axios response:', response)
      console.log('Response data:', response.data)
      // Axios gives us response.data which contains {success: true, data: {files: [...]}}
      return response.data
    },
    enabled: !!currentSession,
    staleTime: Infinity, // Never consider the data stale
    gcTime: Infinity, // Never garbage collect
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  })

  if (!currentSession) {
    return <div className="p-4 text-gray-500 dark:text-gray-400">No session selected</div>
  }

  if (isLoading) {
    return (
      <div className="p-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-4"></div>
          <div className="h-32 bg-gray-100 dark:bg-gray-700 rounded mb-2"></div>
          <div className="h-32 bg-gray-100 dark:bg-gray-700 rounded"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="text-red-600 dark:text-red-400 mb-2">Failed to load diffs</div>
        <button
          onClick={() => refetch()}
          className="text-sm px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Retry
        </button>
      </div>
    )
  }

  console.log('Query result diffsData:', diffsData)
  const diffFiles: DiffFile[] = diffsData?.data?.files || []
  console.log('Extracted diff files:', diffFiles)
  const hasChanges = diffFiles.length > 0

  return (
    <div className="h-full w-full flex flex-col">
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Branch Changes</h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              Showing uncommitted changes and commits since branch creation
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {diffFiles.length} file{diffFiles.length !== 1 ? 's' : ''} changed
            </span>
            <button
              onClick={() => refetch()}
              className="text-sm px-3 py-1 bg-gray-200 dark:bg-gray-700 rounded hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-900 dark:text-gray-100"
            >
              Refresh
            </button>
          </div>
        </div>
        {hasChanges && (
          <div className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            <span className="font-semibold">
              {diffFiles.length} files
            </span>
            {' ('}
            <span className="text-green-600 dark:text-green-400">
              +{diffFiles.reduce((sum: number, f: DiffFile) => sum + f.additions, 0)}
            </span>
            {' / '}
            <span className="text-red-600 dark:text-red-400">
              -{diffFiles.reduce((sum: number, f: DiffFile) => sum + f.deletions, 0)}
            </span>
            {') '}
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        {!hasChanges ? (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">
            <p className="text-lg mb-2">No changes to show</p>
            <p className="text-sm">No changes since branch was created</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {diffFiles.map((file: DiffFile) => (
              <div key={file.path} className="p-4">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className={`
                      text-xs px-2 py-1 rounded font-medium
                      ${file.status === 'added' ? 'bg-green-100 dark:bg-green-800 text-green-800 dark:text-green-200' :
                        file.status === 'deleted' ? 'bg-red-100 dark:bg-red-800 text-red-800 dark:text-red-200' :
                          'bg-yellow-100 dark:bg-yellow-800 text-yellow-800 dark:text-yellow-200'}
                    `}>
                      {file.status}
                    </span>
                    <span className="font-mono text-sm text-gray-900 dark:text-gray-100">{file.path}</span>
                  </div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">
                    <span className="text-green-600 dark:text-green-400">+{file.additions}</span>
                    {' / '}
                    <span className="text-red-600 dark:text-red-400">-{file.deletions}</span>
                  </div>
                </div>
                <div className="bg-gray-900 rounded overflow-hidden">
                  <SyntaxHighlighter
                    language="diff"
                    style={vscDarkPlus}
                    customStyle={{
                      margin: 0,
                      fontSize: '0.875rem',
                      lineHeight: '1.5',
                    }}
                  >
                    {file.diff}
                  </SyntaxHighlighter>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
