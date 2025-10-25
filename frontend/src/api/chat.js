// api/chat.js — thin wrapper that delegates to the Pinia chat store
import { useChatStore } from '@/store/chat'
import api from '@/api'

export const connectWebSocket = () => {
  const store = useChatStore()
  return store.connect()
}

export const disconnectWebSocket = () => {
  const store = useChatStore()
  return store.disconnect()
}

export const sendMessage = (payload) => {
  const store = useChatStore()
  return store.sendMessage(payload)
}

export const getMessages = () => {
  const store = useChatStore()
  return store.messages
}

// Fetch message history between current user and other user (paginated)
export const fetchHistory = (userId, offset = 0) => {
  return api.get('/messages/history', { params: { user_id: userId, offset } })
}