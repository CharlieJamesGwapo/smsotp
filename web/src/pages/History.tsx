import { useState, useEffect } from 'react'
import api from '../api/client'
import toast from 'react-hot-toast'

interface Message {
  id: number
  phone: string
  body: string
  status: string
  gateway_used: string
  scheduled_at: string
  sent_at: string
  error_message: string
  created_at: string
  contact_name: string
}

export default function History() {
  const [messages, setMessages] = useState<Message[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [gatewayFilter, setGatewayFilter] = useState('')

  const load = () => {
    const params = new URLSearchParams({ page: String(page), per_page: '20' })
    if (status) params.set('status', status)
    if (gatewayFilter) params.set('gateway', gatewayFilter)
    api.get(`/messages?${params}`).then((r) => {
      setMessages(r.data.data || [])
      setTotal(r.data.total)
    })
  }

  useEffect(() => { load() }, [page, status, gatewayFilter])

  const retry = async (id: number) => {
    try {
      await api.post(`/messages/${id}/retry`)
      toast.success('Message retried successfully')
      load()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Retry failed')
    }
  }

  const cancel = async (id: number) => {
    if (!confirm('Cancel this scheduled message?')) return
    try {
      await api.delete(`/messages/${id}`)
      toast.success('Message cancelled')
      load()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Cancel failed')
    }
  }

  const statusBadge = (s: string) => {
    const colors: Record<string, string> = {
      sent: 'bg-green-100 text-green-700',
      failed: 'bg-red-100 text-red-700',
      pending: 'bg-yellow-100 text-yellow-700',
    }
    return <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[s] || 'bg-gray-100'}`}>{s}</span>
  }

  const totalPages = Math.ceil(total / 20)

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Message History ({total})</h2>
        <div className="flex gap-2">
          <select
            value={status}
            onChange={(e) => { setStatus(e.target.value); setPage(1) }}
            className="border rounded-lg px-3 py-2 text-sm"
          >
            <option value="">All Status</option>
            <option value="sent">Sent</option>
            <option value="failed">Failed</option>
            <option value="pending">Pending</option>
          </select>
          <select
            value={gatewayFilter}
            onChange={(e) => { setGatewayFilter(e.target.value); setPage(1) }}
            className="border rounded-lg px-3 py-2 text-sm"
          >
            <option value="">All Gateways</option>
            <option value="phone">Phone</option>
            <option value="semaphore">Semaphore</option>
          </select>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Contact</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Phone</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Message</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Status</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Gateway</th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Date</th>
              <th className="text-right px-4 py-3 text-sm font-medium text-gray-500">Actions</th>
            </tr>
          </thead>
          <tbody>
            {messages.map((m) => (
              <tr key={m.id} className="border-t hover:bg-gray-50">
                <td className="px-4 py-3 text-sm">{m.contact_name}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{m.phone}</td>
                <td className="px-4 py-3 text-sm text-gray-600 truncate max-w-52" title={m.body}>{m.body}</td>
                <td className="px-4 py-3">{statusBadge(m.status)}</td>
                <td className="px-4 py-3 text-sm text-gray-600">{m.gateway_used || '-'}</td>
                <td className="px-4 py-3 text-sm text-gray-500">{m.sent_at || m.scheduled_at || m.created_at}</td>
                <td className="px-4 py-3 text-right space-x-2">
                  {m.status === 'failed' && (
                    <button onClick={() => retry(m.id)} className="text-blue-600 text-sm hover:underline">Retry</button>
                  )}
                  {m.status === 'pending' && (
                    <button onClick={() => cancel(m.id)} className="text-red-600 text-sm hover:underline">Cancel</button>
                  )}
                </td>
              </tr>
            ))}
            {messages.length === 0 && (
              <tr><td colSpan={7} className="text-center py-8 text-gray-400">No messages found</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-4">
          {Array.from({ length: totalPages }, (_, i) => (
            <button
              key={i}
              onClick={() => setPage(i + 1)}
              className={`px-3 py-1 rounded ${page === i + 1 ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}
            >
              {i + 1}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
