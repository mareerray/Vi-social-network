import axios from './index'

export function getNotifications() {
  return axios.get('/api/notifications')
}

export function markNotificationsRead(id=null) {
  return axios.post('/api/notifications/mark-read', id ? { id } : {})
}

export function acceptFollowRequest(senderId) {
  return api.post('/api/follow/accept', { sender_id: senderId })
}

export function declineFollowRequest(senderId) {
  return api.post('/api/follow/decline', { sender_id: senderId })
}