import { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { SessionManager } from './components/SessionManager'
import { SessionView } from './components/SessionView'
import { wsClient } from './api/websocket'
import { useSessionActivity } from './hooks/useSessionActivity'

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

function App() {
  // Hook to handle session activity updates
  useSessionActivity()
  
  useEffect(() => {
    // Connect WebSocket on app start
    wsClient.connect()

    return () => {
      wsClient.disconnect()
    }
  }, [])

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
      </Routes>
    </Layout>
  )
}

export default App