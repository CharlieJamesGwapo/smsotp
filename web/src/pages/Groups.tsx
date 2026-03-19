import { useState, useEffect } from 'react'
import api from '../api/client'
import toast from 'react-hot-toast'

interface Group {
  id: number
  name: string
  description: string
  member_count: number
}

interface Contact {
  id: number
  name: string
  phone: string
}

export default function Groups() {
  const [groups, setGroups] = useState<Group[]>([])
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Group | null>(null)
  const [form, setForm] = useState({ name: '', description: '' })
  const [showMembers, setShowMembers] = useState<number | null>(null)
  const [members, setMembers] = useState<Contact[]>([])
  const [allContacts, setAllContacts] = useState<Contact[]>([])
  const [selectedAdd, setSelectedAdd] = useState<number[]>([])

  const load = () => {
    api.get('/groups?per_page=100').then((r) => setGroups(r.data.data || []))
  }

  useEffect(() => { load() }, [])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', description: '' })
    setShowModal(true)
  }

  const openEdit = (g: Group) => {
    setEditing(g)
    setForm({ name: g.name, description: g.description || '' })
    setShowModal(true)
  }

  const save = async () => {
    try {
      if (editing) {
        await api.put(`/groups/${editing.id}`, form)
        toast.success('Group updated')
      } else {
        await api.post('/groups', form)
        toast.success('Group created')
      }
      setShowModal(false)
      load()
    } catch (e: any) {
      toast.error(e.response?.data?.error || 'Error')
    }
  }

  const remove = async (id: number) => {
    if (!confirm('Delete this group?')) return
    await api.delete(`/groups/${id}`)
    toast.success('Group deleted')
    load()
  }

  const openMembers = async (gid: number) => {
    setShowMembers(gid)
    const [groupRes, contactsRes] = await Promise.all([
      api.get(`/groups/${gid}`),
      api.get('/contacts?per_page=100'),
    ])
    setMembers(groupRes.data.members || [])
    setAllContacts(contactsRes.data.data || [])
    setSelectedAdd([])
  }

  const addMembers = async () => {
    if (selectedAdd.length === 0) return
    await api.post(`/groups/${showMembers}/members`, { contact_ids: selectedAdd })
    toast.success('Members added')
    openMembers(showMembers!)
    load()
  }

  const removeMember = async (cid: number) => {
    await api.delete(`/groups/${showMembers}/members`, { data: { contact_ids: [cid] } })
    toast.success('Member removed')
    openMembers(showMembers!)
    load()
  }

  const nonMembers = allContacts.filter((c) => !members.some((m) => m.id === c.id))

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold">Groups</h2>
        <button onClick={openCreate} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          + Add Group
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {groups.map((g) => (
          <div key={g.id} className="bg-white rounded-xl p-5 shadow">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-bold text-lg">{g.name}</h3>
                <p className="text-gray-500 text-sm mt-1">{g.description || 'No description'}</p>
              </div>
              <span className="bg-blue-100 text-blue-700 px-2 py-1 rounded text-sm font-medium">
                {g.member_count} members
              </span>
            </div>
            <div className="flex gap-2 mt-4">
              <button onClick={() => openMembers(g.id)} className="text-sm text-blue-600 hover:underline">Members</button>
              <button onClick={() => openEdit(g)} className="text-sm text-gray-600 hover:underline">Edit</button>
              <button onClick={() => remove(g.id)} className="text-sm text-red-600 hover:underline">Delete</button>
            </div>
          </div>
        ))}
        {groups.length === 0 && (
          <p className="text-gray-400 col-span-3 text-center py-8">No groups yet</p>
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-[420px]">
            <h3 className="text-lg font-bold mb-4">{editing ? 'Edit Group' : 'Add Group'}</h3>
            <div className="space-y-3">
              <input
                placeholder="Group Name *"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
              />
              <textarea
                placeholder="Description"
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="w-full border rounded-lg px-3 py-2"
                rows={2}
              />
            </div>
            <div className="flex justify-end gap-2 mt-4">
              <button onClick={() => setShowModal(false)} className="px-4 py-2 bg-gray-200 rounded-lg">Cancel</button>
              <button onClick={save} className="px-4 py-2 bg-blue-600 text-white rounded-lg">Save</button>
            </div>
          </div>
        </div>
      )}

      {showMembers !== null && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-[560px] max-h-[80vh] overflow-y-auto">
            <h3 className="text-lg font-bold mb-4">Group Members</h3>

            <div className="mb-4">
              <h4 className="font-medium mb-2">Current Members ({members.length})</h4>
              {members.length === 0 ? (
                <p className="text-gray-400 text-sm">No members</p>
              ) : (
                <div className="space-y-1">
                  {members.map((m) => (
                    <div key={m.id} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                      <span>{m.name} ({m.phone})</span>
                      <button onClick={() => removeMember(m.id)} className="text-red-500 text-sm hover:underline">Remove</button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {nonMembers.length > 0 && (
              <div className="border-t pt-4">
                <h4 className="font-medium mb-2">Add Members</h4>
                <div className="space-y-1 max-h-40 overflow-y-auto">
                  {nonMembers.map((c) => (
                    <label key={c.id} className="flex items-center gap-2 px-3 py-1 hover:bg-gray-50 rounded cursor-pointer">
                      <input
                        type="checkbox"
                        checked={selectedAdd.includes(c.id)}
                        onChange={(e) => {
                          if (e.target.checked) setSelectedAdd([...selectedAdd, c.id])
                          else setSelectedAdd(selectedAdd.filter((x) => x !== c.id))
                        }}
                      />
                      <span>{c.name} ({c.phone})</span>
                    </label>
                  ))}
                </div>
                <button onClick={addMembers} className="mt-2 px-4 py-2 bg-green-600 text-white rounded-lg text-sm">
                  Add Selected
                </button>
              </div>
            )}

            <div className="flex justify-end mt-4">
              <button onClick={() => setShowMembers(null)} className="px-4 py-2 bg-gray-200 rounded-lg">Close</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
