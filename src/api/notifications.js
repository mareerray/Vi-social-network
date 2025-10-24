import axios from './index'

export function getNotifications() {
  return axios.get('/api/notifications')
}

export function markNotificationsRead(id=null) {
  return axios.post('/api/notifications/mark-read', id ? { id } : {})
}
