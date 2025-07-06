import { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { ProjectList } from './components/ProjectList'
import { SessionManager } from './components/SessionManager'
import { AgentControl } from './components/AgentControl'
import { wsClient } from './api/websocket'

function Dashboard() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6 p-6">
      <div className="bg-white rounded-lg shadow-sm">
        <SessionManager />
      </div>
      <div className="bg-white rounded-lg shadow-sm">
        <AgentControl />
      </div>
    </div>
  )
}

function App() {
  useEffect(() => {
    // Connect WebSocket on app start
    wsClient.connect()

    return () => {
      wsClient.disconnect()
    }
  }, [])

  return (
    <Layout
      sidebar={
        <div className="h-full flex flex-col">
          <div className="p-4 border-b border-gray-200">
            <h2 className="text-lg font-semibold">Navigation</h2>
          </div>
          <div className="flex-1 overflow-y-auto">
            <ProjectList />
          </div>
        </div>
      }
    >
      <Routes>
        <Route path="/" element={<Dashboard />} />
      </Routes>
    </Layout>
  )
}

export default App