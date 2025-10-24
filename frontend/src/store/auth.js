// store/auth.js
import { defineStore } from 'pinia';
import * as api from '@/api/auth';
import { useChatStore } from '@/store/chat'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null, // Store user data here after login
  }),
  actions: {
    async login(identifier, password) {
      const userData = await api.login(identifier, password);
      // normalize response { user_id: '...' }
      this.user = { user_id: userData.user_id, nickname: userData.nickname || null };
      // auto-connect chat
      const chat = useChatStore()
      chat.connect()
      return this.user
    },
    async register(userData) {
      const res = await api.register(userData)
      // registration endpoint returns { user_id } and sets cookie
      if (res && res.data && res.data.user_id) {
        this.user = { user_id: res.data.user_id }
        const chat = useChatStore()
        chat.connect()
      }
      return res
    },
    async logout() {
      await api.logout();
      const chat = useChatStore()
      chat.disconnect()
      chat.$reset()
      this.user = null;
    },
    // Action to check if user is already logged in (via cookie)
    async fetchUser() {
      try {
        const userData = await api.getMe();
        this.user = userData;
      } catch (error) {
        this.user = null;
      }
    }
  },
});