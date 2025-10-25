import api from './index'
import axios from './index'

export function getNotifications() {
  return api.get('/notifications')
}

export function markNotificationsRead(id=null) {
  return api.post('/notifications/mark-read', id ? { id } : {})
}

export function acceptFollowRequest(senderId) {
  return api.post('/follow/accept', { sender_id: senderId })
}

export function declineFollowRequest(senderId) {
  return api.post('/follow/decline', { sender_id: senderId })
}