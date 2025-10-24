import { defineStore } from 'pinia'
import { getNotifications, markNotificationsRead } from '../api/notifications'

export const useNotificationStore = defineStore('notifications', {
  state: () => ({ list: [] }),
  actions: {
    async fetch() {
      const res = await getNotifications()
      this.list = res.data
    },
    async markAllRead() {
      await markNotificationsRead()
      this.list = this.list.map(n => ({ ...n, is_read: true }))
    },
    async markRead(id) {
      await markNotificationsRead(id)
      this.list = this.list.map(n => (n.id === id ? { ...n, is_read: true } : n))
    }
  }
})
