import { NavLink } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const links = [
  { to: '/dashboard', label: 'Dashboard', icon: '📊' },
  { to: '/contacts', label: 'Contacts', icon: '👥' },
  { to: '/groups', label: 'Groups', icon: '📁' },
  { to: '/templates', label: 'Templates', icon: '📝' },
  { to: '/messages/send', label: 'Send SMS', icon: '📤' },
  { to: '/messages/history', label: 'History', icon: '📋' },
  { to: '/settings', label: 'Settings', icon: '⚙️' },
]

export default function Sidebar() {
  const { user, logout } = useAuth()

  return (
    <div className="w-64 bg-gray-900 text-white min-h-screen flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <h1 className="text-xl font-bold">SMS Notifier</h1>
        <p className="text-gray-400 text-sm mt-1">IT CEBU PARK</p>
      </div>

      <nav className="flex-1 p-4 space-y-1">
        {links.map((l) => (
          <NavLink
            key={l.to}
            to={l.to}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                isActive
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-300 hover:bg-gray-800 hover:text-white'
              }`
            }
          >
            <span>{l.icon}</span>
            <span>{l.label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-gray-700">
        <p className="text-sm text-gray-400 mb-2">Logged in as <strong>{user?.username}</strong></p>
        <button
          onClick={logout}
          className="w-full py-2 bg-red-600 hover:bg-red-700 rounded-lg text-sm transition-colors"
        >
          Logout
        </button>
      </div>
    </div>
  )
}
