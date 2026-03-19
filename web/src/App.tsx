import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './context/AuthContext'
import Sidebar from './components/Sidebar'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Contacts from './pages/Contacts'
import Groups from './pages/Groups'
import Templates from './pages/Templates'
import SendSMS from './pages/SendSMS'
import History from './pages/History'
import Settings from './pages/Settings'

function Layout({ children }: { children: React.ReactNode }) {
  const { user } = useAuth()

  return (
    <div className="flex min-h-screen bg-gray-100">
      <Sidebar />
      <main className="flex-1 p-6 overflow-auto">
        {user?.default_password && (
          <div className="mb-4 p-3 bg-yellow-100 border border-yellow-400 text-yellow-800 rounded-lg">
            ⚠️ You are using the default password. Please change it in <a href="/settings" className="underline font-bold">Settings</a>.
          </div>
        )}
        {children}
      </main>
    </div>
  )
}

export default function App() {
  const { token } = useAuth()

  if (!token) {
    return (
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="*" element={<Navigate to="/login" />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route path="/login" element={<Navigate to="/dashboard" />} />
      <Route path="/dashboard" element={<Layout><Dashboard /></Layout>} />
      <Route path="/contacts" element={<Layout><Contacts /></Layout>} />
      <Route path="/groups" element={<Layout><Groups /></Layout>} />
      <Route path="/templates" element={<Layout><Templates /></Layout>} />
      <Route path="/messages/send" element={<Layout><SendSMS /></Layout>} />
      <Route path="/messages/history" element={<Layout><History /></Layout>} />
      <Route path="/settings" element={<Layout><Settings /></Layout>} />
      <Route path="*" element={<Navigate to="/dashboard" />} />
    </Routes>
  )
}
