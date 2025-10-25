import { defineStore } from 'pinia'
import { useAuthStore } from '@/store/auth'
import { fetchHistory } from '@/api/chat'
import { fetchGroupMessages } from '@/api/groups'
import { useNotificationStore } from '@/store/notification'

const normalizeAvatar = (path) => {
	if (!path) return ''
	if (/^https?:/i.test(path)) return path
	const normalized = path.startsWith('/') ? path : `/${path}`
	return `${window.location.protocol}//${window.location.host}${normalized}`
}

export const useChatStore = defineStore('chat', {
	state: () => ({
		socket: null,
		connected: false,
		messageQueue: [], // Queue messages before socket opens
		contacts: [],
		conversations: {},
		activeContactId: null,
		errors: [],
		typingUsers: {},
		currentUserId: null,
		// batching for incoming realtime messages
		_incomingBuffer: [],
		_incomingFlushTimer: null,
		// group chat storage: keyed by group id
		groupConversations: {},
		_activeGroupId: null,
	}),
	getters: {
		activeConversation(state) {
			const key = state.activeContactId ? String(state.activeContactId) : ''
			return key && state.conversations[key] ? state.conversations[key] : []
		},
		activeContact(state) {
			const key = state.activeContactId ? String(state.activeContactId) : ''
			return state.contacts.find((c) => c.id === key) || null
		},
	},
	actions: {
		connect() {
			if (this.socket && this.connected) return

			// connect to backend websocket (backend runs on :8080)
			const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			const url = `${protocol}//${window.location.host}/ws`;
			this.socket = new WebSocket(url)

			console.log('chat: connecting to', url)

			this.socket.onopen = () => {
				this.connected = true
				this.currentUserId = this.getCurrentUserId()

				// Flush queued messages
				while (this.messageQueue.length) {
					const msg = this.messageQueue.shift()
					this.socket.send(JSON.stringify(msg))
				}

				// Request user list after connection established
				this.requestUserList()

				console.log('chat: connected as user', this.currentUserId)
			}

			this.socket.onmessage = (event) => {
				let payload
				try {
					payload = JSON.parse(event.data)
				} catch (err) {
					console.error('chat: invalid payload', err)
					return
				}
				// forward certain payloads to notification store for realtime UX
				const notifTypes = new Set(['group_invite','group_invite_response','group_join_request','group_join_response','group_event','new_message','group_message','new_follower','follow_request','follow_request_accepted','follow_request_declined'])
				if (payload && payload.type && notifTypes.has(payload.type)) {
					try {
						const notifStore = useNotificationStore()
						const tempId = -Date.now()
						notifStore.list.unshift({
							id: tempId,
							recipient_id: Number(this.getCurrentUserId()) || 0,
							actor_id: 0,
							type: payload.type,
							data: JSON.stringify(payload.data || {}),
							is_read: false,
							created_at: new Date().toISOString(),
						})
					} catch (e) {
						console.error('Failed to push realtime notification', e)
					}
				}
				this.handleMessage(payload)
			}

			this.socket.onclose = () => {
				console.log('chat: disconnected at socket.onclose')
				this.connected = false
				this.socket = null
			}

			this.socket.onerror = (err) => {
				console.error('chat socket error', err)
				this.pushError('Connection problem. Please try again.')
			}
		},

		disconnect() {
			console.log('chat: disconnecting')
			if (this.socket) this.socket.close()
			this.socket = null
			this.connected = false
			this.messageQueue = []
		},

		getCurrentUserId() {
			const auth = useAuthStore()
			return auth.user ? String(auth.user.user_id) : this.currentUserId || ''
		},

		requestUserList() {
			const payload = { type: 'user_list_request' }
			this.sendMessage(payload)
		},

		sendMessage(payload) {
			const body = { ...payload }
			if (!body.type) body.type = 'message'
			if (!body.receiver_id) body.receiver_id = this.activeContactId
			if (!body.receiver_id) {
				this.pushError('Choose someone to chat with.')
				return false
			}
			if (!body.content || !body.content.trim()) return false

			body.receiver_id = String(body.receiver_id)
			body.content = body.content.trim()

			if (this.socket && this.socket.readyState === WebSocket.OPEN) {
				this.socket.send(JSON.stringify(body))
			} else {
				// Queue the message if socket not open
				this.messageQueue.push(body)
			}
			return true
		},

		handleMessage(msg) {
			switch (msg.type) {
				case 'user_list':
					this.handleUserList(msg)
					break
				case 'message':
					this.bufferIncomingMessage(msg)
					break
				case 'typing':
					this.typingUsers[String(msg.sender_id)] = true
					break
				case 'stop_typing':
					delete this.typingUsers[String(msg.sender_id)]
					break
				case 'group_message':
					this.handleGroupMessage(msg)
					break
				case 'error':
					this.pushError(msg.content || 'Unable to deliver message.')
					break
				case 'new_message_notification':
					break
				default:
					break
			}
		},

		handleGroupMessage(msg) {
			// msg should contain group_id, sender_id, content, id, sender_name
			const gid = msg.group_id ? String(msg.group_id) : null
			if (!gid) return
			if (!this.groupConversations[gid]) this.groupConversations[gid] = []
			const me = this.getCurrentUserId()
			this.groupConversations[gid].push({
				id: msg.id || `${Date.now()}-${this.groupConversations[gid].length}`,
				content: msg.content,
				outgoing: String(msg.sender_id) === me,
				senderName: msg.sender_name || '',
				timestamp: msg.created_at || new Date().toISOString(),
			})
		},

		async fetchGroupHistory(groupId) {
			if (!groupId) return []
			try {
				const { data } = await fetchGroupMessages(groupId)
				// ensure array exists
				this.groupConversations[String(groupId)] = []
				const me = this.getCurrentUserId()
				if (Array.isArray(data)) {
					data.forEach((m) => {
						this.groupConversations[String(groupId)].push({
							id: m.id,
							content: m.content,
							outgoing: String(m.sender_id) === me,
							senderName: m.sender_name || '',
							timestamp: m.created_at || new Date().toISOString(),
						})
					})
				}
				return this.groupConversations[String(groupId)]
			} catch (err) {
				console.error('Failed to load group history', err)
				return []
			}
		},

		async loadMoreGroupHistory(groupId, limit = 50) {
			if (!groupId) return 0
			const key = String(groupId)
			const conv = this.groupConversations[key] || []
			let beforeId = null
			if (conv.length > 0) {
				beforeId = conv[0].id
			}
			try {
				const { data } = await fetchGroupMessages(groupId, { before_id: beforeId, limit })
				if (!Array.isArray(data) || data.length === 0) return 0
				// prepend older messages
				const me = this.getCurrentUserId()
				const items = data.map((m) => ({
					id: m.id,
					content: m.content,
					outgoing: String(m.sender_id) === me,
					senderName: m.sender_name || '',
					timestamp: m.created_at || new Date().toISOString(),
				}))
				this.groupConversations[key] = items.concat(this.groupConversations[key] || [])
				return items.length
			} catch (err) {
				console.error('Failed to load more group history', err)
				return 0
			}
		},

		sendGroupMessage(payload) {
			const body = { ...payload }
			if (!body.type) body.type = 'group_message'
			if (!body.group_id) {
				this.pushError('Group id required')
				return false
			}
			if (!body.content || !body.content.trim()) return false
			body.content = body.content.trim()
			if (this.socket && this.socket.readyState === WebSocket.OPEN) {
				this.socket.send(JSON.stringify(body))
			} else {
				this.messageQueue.push(body)
			}
			return true
		},

		bufferIncomingMessage(msg) {
			// Add to buffer and schedule flush: batch updates to avoid UI jank and
			// allow scrolling through history when many messages arrive.
			this._incomingBuffer.push(msg)
			if (!this._incomingFlushTimer) {
				this._incomingFlushTimer = setTimeout(() => this.flushIncomingBuffer(), 300)
			}
			// if buffer grows large, flush immediately
			if (this._incomingBuffer.length >= 20) this.flushIncomingBuffer()
		},

		flushIncomingBuffer() {
			if (!this._incomingBuffer.length) return
			const buf = this._incomingBuffer.splice(0)
			if (this._incomingFlushTimer) {
				clearTimeout(this._incomingFlushTimer)
				this._incomingFlushTimer = null
			}
			// process buffered messages in one batch
			const me = this.getCurrentUserId()
			buf.forEach((msg) => {
				const sender = String(msg.sender_id)
				const receiver = String(msg.receiver_id)
				const otherId = sender === me ? receiver : sender
				if (!otherId) return
				const conversation = this.ensureConversation(otherId)
				conversation.push({
					id: msg.id || `${Date.now()}-${conversation.length}`,
					content: msg.content,
					outgoing: sender === me,
					senderName: msg.sender_name || '',
					timestamp: msg.created_at || new Date().toISOString(),
				})
				if (this.activeContactId === otherId) this.markRead(otherId)
				else this.incrementUnread(otherId)
				if (!this.contacts.find((c) => c.id === otherId)) {
					this.contacts.push({
						id: otherId,
						nickname: msg.sender_name || `User ${otherId}`,
						displayName: msg.sender_name || `User ${otherId}`,
						avatar: '',
						isOnline: true,
						unread: sender === me ? 0 : 1,
					})
				}
			})
		},

		handleUserList(msg) {
			let users = []
			if (Array.isArray(msg.users)) {
				users = msg.users
			} else if (msg.content) {
				try {
					users = JSON.parse(msg.content)
				} catch (err) {
					console.error('chat: cannot parse user list', err)
					return
				}
			}
			const normalized = users.map((user) => {
				const id = String(user.id)
				const existing = this.contacts.find((c) => c.id === id)
				return {
					id,
					nickname: user.nickname || user.display_name || `User ${id}`,
					displayName: user.display_name || user.nickname || `User ${id}`,
					avatar: normalizeAvatar(user.avatar),
					isOnline: !!user.is_online,
					unread: existing ? existing.unread : 0,
				}
			})
			this.contacts = normalized

			if (this.activeContactId && !this.contacts.find((c) => c.id === this.activeContactId)) {
				const fallback = this.contacts.length ? this.contacts[0].id : null
				this.activeContactId = null
				if (fallback) this.setActiveContact(fallback)
			} else if (!this.activeContactId && this.contacts.length) {
				this.setActiveContact(this.contacts[0].id)
			}
		},

		ensureConversation(id) {
			const key = String(id)
			if (!this.conversations[key]) this.conversations[key] = []
			return this.conversations[key]
		},

		handleChatMessage(msg) {
			const me = this.getCurrentUserId()
			const sender = String(msg.sender_id)
			const receiver = String(msg.receiver_id)
			const otherId = sender === me ? receiver : sender
			if (!otherId) return

			const conversation = this.ensureConversation(otherId)
			conversation.push({
				id: msg.id || `${Date.now()}-${conversation.length}`,
				content: msg.content,
				outgoing: sender === me,
				senderName: msg.sender_name || '',
				timestamp: msg.created_at || new Date().toISOString(),
			})

			if (this.activeContactId === otherId) this.markRead(otherId)
			else this.incrementUnread(otherId)

			if (!this.contacts.find((c) => c.id === otherId)) {
				this.contacts.push({
					id: otherId,
					nickname: msg.sender_name || `User ${otherId}`,
					displayName: msg.sender_name || `User ${otherId}`,
					avatar: '',
					isOnline: true,
					unread: sender === me ? 0 : 1,
				})
			}
		},

		async setActiveContact(id) {
			this.activeContactId = id ? String(id) : null
			if (this.activeContactId) {
				this.ensureConversation(this.activeContactId)
				// load recent history from server (first page)
				try {
					const { data } = await fetchHistory(this.activeContactId, 0)
						// data is expected to be an array; be defensive and treat non-arrays as empty
						const conv = this.ensureConversation(this.activeContactId)
						conv.splice(0, conv.length) // clear existing
						const me = this.getCurrentUserId()
						const messages = Array.isArray(data) ? data : []
						if (!Array.isArray(data) && data != null) {
							// only log unexpected non-null non-array responses for debugging
							console.warn('fetchHistory returned non-array (unexpected):', data)
						}
						messages.forEach((m) => {
							conv.push({
								id: m.id,
								content: m.content,
								outgoing: String(m.sender_id) === me,
								senderName: m.sender_name || '',
								timestamp: m.created_at || new Date().toISOString(),
							})
						})
					this.markRead(this.activeContactId)
				} catch (err) {
					console.error('Failed to load history', err)
				}
			}
		},

		async loadMoreHistory(contactId) {
			if (!contactId) return 0
			const key = String(contactId)
			const conv = this.ensureConversation(key)
			const offset = conv.length
			try {
				const { data } = await fetchHistory(key, offset)
				if (!Array.isArray(data) || data.length === 0) return 0
				const me = this.getCurrentUserId()
				const items = data.map((m) => ({
					id: m.id,
					content: m.content,
					outgoing: String(m.sender_id) === me,
					senderName: m.sender_name || '',
					timestamp: m.created_at || new Date().toISOString(),
				}))
				this.conversations[key] = items.concat(this.conversations[key] || [])
				return items.length
			} catch (err) {
				console.error('Failed to load more history', err)
				return 0
			}
		},

		incrementUnread(id) {
			const contact = this.contacts.find((c) => c.id === String(id))
			if (contact) contact.unread += 1
		},

		markRead(id) {
			const contact = this.contacts.find((c) => c.id === String(id))
			if (contact) contact.unread = 0
		},

		pushError(message) {
			if (message) this.errors.push(message)
		},

		nextError() {
			return this.errors.shift() || null
		},
	},
})
