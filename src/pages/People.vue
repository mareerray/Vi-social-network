<template>
  <div class="container py-4 discover">
    <div class="d-flex flex-column flex-md-row align-items-md-center justify-content-md-between mb-4">
      <div>
        <h2 class="mb-1"><i class="fas fa-user-friends me-2"></i>Discover People</h2>
        <p class="text-muted mb-0">Find new friends and connect instantly.</p>
      </div>
      <div class="mt-3 mt-md-0 d-flex gap-2">
        <button class="btn btn-outline-secondary" type="button" @click="load" :disabled="loading">
          <i class="fas fa-sync-alt me-1" :class="{ 'fa-spin': loading }"></i>
          Refresh
        </button>
      </div>
    </div>

    <div class="row gy-3 mb-4">
      <div class="col-md-6 col-lg-4">
        <div class="input-group">
          <span class="input-group-text"><i class="fas fa-search"></i></span>
          <input
            v-model="search"
            type="text"
            class="form-control"
            placeholder="Search people by name or nickname"
          />
        </div>
      </div>
      <div class="col-md-6 col-lg-4" v-if="error">
        <div class="alert alert-danger py-2 px-3 mb-0">
          <i class="fas fa-exclamation-triangle me-2"></i>{{ error }}
        </div>
      </div>
    </div>

    <div v-if="loading" class="text-center text-muted py-5">
      <div class="spinner-border text-primary mb-3" role="status"></div>
      <p class="mb-0">Loading users...</p>
    </div>

    <div v-else-if="filteredUsers.length === 0" class="text-center text-muted py-5">
      <i class="fas fa-user-slash fa-3x mb-3"></i>
      <p class="mb-0">No users found. Try a different search.</p>
    </div>

    <div v-else class="row g-4">
      <div v-for="user in filteredUsers" :key="user.id" class="col-sm-6 col-lg-4">
        <div class="card h-100 shadow-sm border-0 user-card">
          <div class="card-body text-center d-flex flex-column">
            <div class="avatar-wrapper mx-auto mb-3">
              <img v-if="user.avatar_url" :src="user.avatar_url" class="rounded-circle" alt="avatar" />
              <div v-else class="avatar-placeholder rounded-circle d-flex align-items-center justify-content-center">
                <i class="fas fa-user fa-2x text-primary"></i>
              </div>
            </div>

            <h5 class="mb-1">{{ user.display_name }}</h5>
            <p class="text-muted mb-2">@{{ user.nickname || ('user' + user.id) }}</p>
            <span class="badge" :class="user.profile_type === 'public' ? 'bg-success' : 'bg-warning text-dark'">
              <i :class="user.profile_type === 'public' ? 'fas fa-globe' : 'fas fa-lock'"></i>
              {{ user.profile_type }} profile
            </span>

            <div class="mt-3 small text-muted" v-if="user.is_self">
              This is you.
            </div>

            <div class="d-flex flex-column gap-2 mt-auto pt-3">
              <router-link :to="`/profile/${user.id}`" class="btn btn-outline-primary">
                <i class="fas fa-id-card me-1"></i>View Profile
              </router-link>
              <button
                v-if="!user.is_self"
                :class="['btn', followButtonClass(user)]"
                type="button"
                :disabled="user.busy || user.request_pending"
                @click="handleFollow(user)"
              >
                <span v-if="user.busy" class="spinner-border spinner-border-sm me-2" role="status"></span>
                <i :class="followIcon(user)" class="me-1"></i>
                {{ followLabel(user) }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import * as userApi from '@/api/users'

export default {
  setup() {
    const users = ref([])
    const loading = ref(false)
    const search = ref('')
    const error = ref('')

  const resolveAvatar = (path) => path

    const normalizeUser = (user) => ({
      ...user,
      display_name: user.display_name || user.nickname || `User ${user.id}`,
      avatar_url: resolveAvatar(user.avatar),
      busy: false,
    })

    const load = async () => {
      loading.value = true
      error.value = ''
      try {
        const data = await userApi.listUsers()
        users.value = Array.isArray(data) ? data.map(normalizeUser) : []
      } catch (err) {
        console.error('Failed to load users', err)
        error.value = 'Unable to load users right now.'
        users.value = []
      } finally {
        loading.value = false
      }
    }

    const followLabel = (user) => {
      if (user.is_following) return 'Following'
      if (user.request_pending) return 'Request Sent'
      return user.profile_type === 'public' ? 'Follow' : 'Request Follow'
    }

    const followIcon = (user) => {
      if (user.is_following) return 'fas fa-user-check'
      if (user.request_pending) return 'fas fa-hourglass-half'
      return 'fas fa-user-plus'
    }

    const followButtonClass = (user) => {
      if (user.is_following) return 'btn-outline-danger'
      if (user.request_pending) return 'btn-outline-secondary'
      return 'btn-primary'
    }

    const handleFollow = async (user) => {
      if (user.is_self || user.busy) return
      user.busy = true
      try {
        if (user.is_following) {
          await userApi.unfollow(user.id)
          user.is_following = false
          user.request_pending = false
        } else {
          await userApi.follow(user.id)
          if (user.profile_type === 'public') {
            user.is_following = true
          } else {
            user.request_pending = true
          }
        }
      } catch (err) {
        console.error('Follow/unfollow failed', err)
        error.value = 'Action failed. Please try again.'
        setTimeout(() => { if (error.value === 'Action failed. Please try again.') error.value = '' }, 4000)
      } finally {
        user.busy = false
      }
    }

    const filteredUsers = computed(() => {
      const term = search.value.trim().toLowerCase()
      if (!term) return users.value
      return users.value.filter((user) => {
        const name = (user.display_name || '').toLowerCase()
        const nickname = (user.nickname || '').toLowerCase()
        return name.includes(term) || nickname.includes(term)
      })
    })

    onMounted(load)

    return {
      users,
      loading,
      search,
      error,
      filteredUsers,
      load,
      handleFollow,
      followLabel,
      followIcon,
      followButtonClass,
    }
  },
}
</script>

<style scoped>
.discover {
  max-width: 1100px;
}

.user-card {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.user-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.08);
}

.avatar-wrapper {
  width: 96px;
  height: 96px;
  position: relative;
}

.avatar-wrapper img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
  border: 3px solid rgba(102, 126, 234, 0.3);
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background: rgba(118, 75, 162, 0.1);
}

.badge {
  font-weight: 500;
  text-transform: capitalize;
}

.btn-outline-danger {
  --bs-btn-color: #dc3545;
  --bs-btn-border-color: #dc3545;
  --bs-btn-hover-bg: #dc3545;
  --bs-btn-hover-border-color: #dc3545;
  --bs-btn-hover-color: #fff;
}
</style>
