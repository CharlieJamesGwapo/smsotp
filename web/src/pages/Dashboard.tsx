import { useState, useEffect } from 'react'
import api from '../api/client'

interface Stats {
  total_contacts: number
  total_groups: number
  total_templates: number
  messages_sent_today: number
  messages_sent_week: number
  messages_sent_month: number
  messages_failed: number
  messages_scheduled: number
  success_rate: number
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null)

  useEffect(() => {
    api.get('/dashboard/stats').then((r) => setStats(r.data))
  }, [])

  if (!stats) return <div className="text-center py-10">Loading...</div>

  const cards = [
    { label: 'Total Contacts', value: stats.total_contacts, color: 'bg-blue-500' },
    { label: 'Groups', value: stats.total_groups, color: 'bg-purple-500' },
    { label: 'Templates', value: stats.total_templates, color: 'bg-indigo-500' },
    { label: 'Sent Today', value: stats.messages_sent_today, color: 'bg-green-500' },
    { label: 'Sent This Week', value: stats.messages_sent_week, color: 'bg-teal-500' },
    { label: 'Sent This Month', value: stats.messages_sent_month, color: 'bg-cyan-500' },
    { label: 'Failed', value: stats.messages_failed, color: 'bg-red-500' },
    { label: 'Scheduled', value: stats.messages_scheduled, color: 'bg-yellow-500' },
    { label: 'Success Rate', value: `${stats.success_rate.toFixed(1)}%`, color: 'bg-emerald-500' },
  ]

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6">Dashboard</h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {cards.map((c) => (
          <div key={c.label} className={`${c.color} text-white rounded-xl p-6 shadow-lg`}>
            <p className="text-sm opacity-80">{c.label}</p>
            <p className="text-3xl font-bold mt-2">{c.value}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
