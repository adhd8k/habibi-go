import { useEffect } from 'react'
import { Provider } from 'react-redux'
import { Routes, Route } from 'react-router-dom'
import { store } from './store'
import { useAppDispatch } from './hooks'
import { websocketConnect } from './middleware/websocket'
import { AuthModal } from '../features/auth/components/AuthModal'
import { ProjectListContainer } from '../features/projects/components/ProjectListContainer'
import { Layout } from '../components/Layout'
import { SessionManager } from '../components/SessionManager'
import { SessionView } from '../components/SessionView'
import { selectCredentials } from '../features/auth/slice/authSlice'
import { StoreSync } from '../shared/components/StoreSync'

function Dashboard() {
  return (
    <div className="h-full flex flex-col lg:flex-row gap-4 lg:gap-6 p-4 lg:p-6">
      <div className="w-full lg:w-96 bg-white rounded-lg shadow-sm flex-shrink-0">
        <SessionManager />
      </div>
      <div className="flex-1 bg-white rounded-lg shadow-sm flex min-h-0">
        <SessionView />
      </div>
    </div>
  )
}

function AppContent() {
  const dispatch = useAppDispatch()

  useEffect(() => {
    // Get WebSocket URL with auth if needed
    const getWebSocketUrl = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const credentials = selectCredentials(store.getState())
      
      if (credentials) {
        return `${protocol}//${credentials.username}:${credentials.password}@${host}/ws`
      }
      
      return `${protocol}//${host}/ws`
    }

    // Connect WebSocket on app start
    dispatch(websocketConnect(getWebSocketUrl()) as any)

    return () => {
      dispatch({ type: 'websocket/disconnect' } as any)
    }
  }, [dispatch])

  return (
    <>
      <StoreSync />
      <AuthModal />
      <Layout
        sidebar={
          <div className="h-full flex flex-col">
            <div className="p-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold">Navigation</h2>
            </div>
            <div className="flex-1 overflow-y-auto">
              <ProjectListContainer />
            </div>
          </div>
        }
      >
        <Routes>
          <Route path="/" element={<Dashboard />} />
        </Routes>
      </Layout>
    </>
  )
}

export default function App() {
  return (
    <Provider store={store}>
      <AppContent />
    </Provider>
  )
}