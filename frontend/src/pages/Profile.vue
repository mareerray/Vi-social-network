<template>
  <div class="container py-4">
    <div v-if="profile" class="profile-container">
      <!-- Private Profile View -->
      <div v-if="!profile.is_accessible" class="private-profile-view text-center py-5">
        <div class="card">
          <div class="card-body">
            <div class="profile-avatar-container mb-3">
              <div class="profile-avatar">
                <img v-if="profile.avatar" :src="profile.avatar" alt="avatar" class="avatar-img" />
                <i v-else class="fas fa-user-circle fa-5x text-primary"></i>
              </div>
            </div>
            <h2 class="profile-name mb-1">{{ profile.nickname || 'User' }}</h2>
            <span class="badge bg-warning mb-4">
              <i class="fas fa-lock"></i>
              {{ profile.profile_type }} Profile
            </span>
            <div class="alert alert-warning" role="alert">
              <h4 class="alert-heading"><i class="fas fa-eye-slash"></i> This Account is Private</h4>
              <p>Follow this account to see their content.</p>
            </div>
            <button 
              @click="toggleFollow" 
              :class="following ? 'btn btn-outline-danger' : 'btn btn-primary'"
              class="btn-lg px-4 mt-3"
            >
              <i :class="following ? 'fas fa-user-minus' : 'fas fa-user-plus'"></i>
              {{ following ? 'Unfollow' : (pending ? 'Request Sent' : 'Follow') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Full Profile View -->
      <div v-else>
        <!-- Profile Header -->
        <div class="row">
          <div class="col-12">
            <div class="card profile-header-card">
              <div class="profile-cover"></div>
              <div class="card-body position-relative">
                <!-- Profile Avatar -->
                <div class="profile-avatar-container">
                  <div class="profile-avatar">
                    <img v-if="profile.avatar" :src="profile.avatar" alt="avatar" class="avatar-img" />
                    <i v-else class="fas fa-user-circle fa-5x text-primary"></i>
                  </div>
                  <!-- Avatar upload removed: avatars are provided via URL only -->
                  <div v-if="isMine" class="edit-avatar-btn" style="display:none"></div>
                </div>

                <!-- Profile Info -->
                <div class="profile-info mt-4">
                  <div class="row align-items-start">
                    <div class="col-md-8">
                      <h2 class="profile-name mb-1">{{ profile.nickname || 'Anonymous User' }}</h2>
                      <h5 class="text-muted mb-2">{{ profile.first_name }} {{ profile.last_name }}</h5>
                      <p class="profile-about mb-3" v-if="profile.about">{{ profile.about }}</p>
                      
                      <!-- Profile Stats -->
                      <div class="profile-stats d-flex gap-4 mb-3">
                        <div class="stat-item">
                          <strong class="d-block text-primary">{{ followers.length }}</strong>
                          <small class="text-muted">Followers</small>
                        </div>
                        <div class="stat-item">
                          <strong class="d-block text-primary">{{ followingList.length }}</strong>
                          <small class="text-muted">Following</small>
                        </div>
                        <div class="stat-item">
                          <strong class="d-block text-primary">{{ postCount }}</strong>
                          <small class="text-muted">Posts</small>
                        </div>
                      </div>

                      <!-- Contact Info -->
                      <div v-if="profile.email && isMine" class="contact-info mb-3">
                        <small class="text-muted">
                          <i class="fas fa-envelope me-2"></i>
                          {{ profile.email }}
                        </small>
                      </div>
                    </div>

                    <div class="col-md-4 text-md-end">
                      <!-- Privacy Settings (Own Profile) -->
                      <div v-if="isMine" class="privacy-controls mb-3">
                        <label class="form-label fw-semibold">
                          <i class="fas fa-shield-alt text-primary me-2"></i>
                          Profile Privacy
                        </label>
                        <select v-model="profile.profile_type" @change="changePrivacy" class="form-select">
                          <option value="public">
                            Public
                          </option>
                          <option value="private">
                            Private
                          </option>
                        </select>
                      </div>

                      <!-- Follow Controls (Other Profiles) -->
                      <div v-else class="follow-controls">
                        <button 
                          @click="toggleFollow" 
                          :class="following ? 'btn btn-outline-danger' : 'btn btn-primary'"
                          class="btn-lg px-4"
                          :disabled="pending"
                        >
                          <i :class="following ? 'fas fa-user-minus' : 'fas fa-user-plus'"></i>
                          {{ following ? 'Unfollow' : (pending ? 'Request Sent' : 'Follow') }}
                        </button>
                      </div>

                      <!-- Privacy Badge -->
                      <div class="privacy-badge">
                        <span class="badge" :class="profile.profile_type === 'public' ? 'bg-success' : 'bg-warning'">
                          <i :class="profile.profile_type === 'public' ? 'fas fa-globe' : 'fas fa-lock'"></i>
                          {{ profile.profile_type }} Profile
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Profile Content -->
        <div class="row mt-4">
          <!-- Left Sidebar -->
          <div class="col-lg-4">
            <div class="sidebar-sticky">
              <!-- About Card -->
              <div class="card mb-4">
                <div class="card-header bg-gradient">
                  <h5 class="mb-0 text-dark"><i class="fas fa-user-circle me-2"></i>About</h5>
                </div>
                <div class="card-body">
                  <ul class="list-unstyled">
                    <li v-if="profile.first_name" class="mb-2">
                      <i class="fas fa-user me-2 text-muted"></i>
                      <strong>Name:</strong> {{ profile.first_name }} {{ profile.last_name }}
                    </li>
                    <li v-if="profile.date_of_birth" class="mb-2">
                      <i class="fas fa-calendar-alt me-2 text-muted"></i>
                      <strong>Born:</strong> {{ profile.date_of_birth }}
                    </li>
                    <li v-if="profile.created_at" class="mb-2">
                      <i class="fas fa-clock me-2 text-muted"></i>
                      <strong>Joined:</strong> {{ profile.created_at }}
                    </li>
                  </ul>
                </div>
              </div>

              <!-- Follow Requests Card removed per UX request -->

              <!-- Followers Card -->
              <div class="card mb-4">
                <div class="card-header bg-gradient">
                  <h5 class="mb-0 text-dark">
                    <i class="fas fa-users me-2"></i>
                    Followers ({{ followers.length }})
                  </h5>
                </div>
                <div class="card-body">
                  <div v-if="followers.length === 0" class="text-center py-3">
                    <p class="text-muted mb-0">No followers yet</p>
                  </div>
                  <div v-else class="list-group list-group-flush">
                    <div v-for="f in followers" :key="f.id" class="list-group-item d-flex align-items-center">
                      <i class="fas fa-user-circle fa-2x text-primary me-3"></i>
                      <span>{{ f.nickname }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Following Card -->
              <div class="card">
                <div class="card-header bg-gradient">
                  <h5 class="mb-0 text-dark">
                    <i class="fas fa-heart me-2"></i>
                    Following ({{ followingList.length }})
                  </h5>
                </div>
                <div class="card-body">
                  <div v-if="followingList.length === 0" class="text-center py-3">
                    <p class="text-muted mb-0">Not following anyone yet</p>
                  </div>
                  <div v-else class="list-group list-group-flush">
                    <div v-for="f in followingList" :key="f.id" class="list-group-item d-flex align-items-center">
                      <i class="fas fa-user-circle fa-2x text-success me-3"></i>
                      <span>{{ f.nickname }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Right Content: Posts -->
          <div class="col-lg-8">
            <!-- Create Post (if my profile) -->
            <div v-if="isMine" class="mb-4">
              <CreatePost @post-created="load" />
            </div>
            
            <!-- User's Posts -->
            <div class="posts-feed">
              <h4 class="mb-3">Posts</h4>
              <div v-if="posts.length === 0" class="card text-center py-5">
                <div class="card-body">
                  <i class="fas fa-newspaper fa-3x text-muted mb-3"></i>
                  <h5 class="text-muted">No Posts Yet</h5>
                  <p class="text-muted">This user hasn't posted anything.</p>
                </div>
              </div>
              <div v-for="p in posts" :key="p.id" class="mb-4">
                <PostCard :post="p" @comment-added="load" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else class="text-center py-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
      <p class="text-muted mt-2">Loading profile...</p>
    </div>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import * as api from '@/api/users'
import { listPosts } from '@/api/post'
import { useAuthStore } from '@/store/auth'
import PostCard from '@/components/PostCard.vue'
import CreatePost from '@/components/CreatePost.vue'

export default {
  components: { PostCard, CreatePost },
  setup() {
    const auth = useAuthStore()
    const route = useRoute()
    
    const profile = ref(null)
    const followers = ref([])
    const followingList = ref([])
  const posts = ref([])
  const followRequests = ref([])
    const following = ref(false)
    const pending = ref(false) // To track pending follow requests

    const userId = computed(() => route.params.id || auth.user?.user_id)
    const isMine = computed(() => auth.user && profile.value && String(auth.user.user_id) === String(profile.value.id))
    const postCount = computed(() => posts.value.length)

    const msg = ref('')

    // avatar will be provided by the user as a URL only

    const load = async () => {
      if (!userId.value) return
      
      try {
        console.log('Loading profile for user ID:', userId.value)
        const p = await api.getProfile(userId.value)
        // defensive: ensure we always have an object
        profile.value = (p && typeof p === 'object') ? p : { is_accessible: false }

        // If profile is not accessible, we only need to check follow status
        if (!profile.value.is_accessible) {
          if (auth.user && !isMine.value) {
            const status = await api.getFollowStatus(profile.value.id)
            following.value = !!status.following
            pending.value = !!status.request_pending
          }
          return
        }

        // If accessible, load all data
        const f = await api.getFollowers(userId.value)
        followers.value = Array.isArray(f) ? f : []
        const fl = await api.getFollowing(userId.value)
        followingList.value = Array.isArray(fl) ? fl : []
        const pList = await listPosts(userId.value)
        posts.value = Array.isArray(pList) ? pList : []
        if (isMine.value) {
          await refreshRequests()
        } else {
          followRequests.value = []
        }

        // Check if the logged-in user is following this profile
        if (auth.user) {
          if (isMine.value) {
            following.value = false
            pending.value = false
          } else {
            const status = await api.getFollowStatus(profile.value.id)
            following.value = !!status.following
            pending.value = !!status.request_pending
          }
        }
      } catch (error) {
        console.error("Failed to load profile data:", error)
        profile.value = { is_accessible: false, nickname: 'Error', avatar: null, profile_type: 'private' } // Show private view on error
        followers.value = []
        followingList.value = []
        posts.value = []
        followRequests.value = []
      }
    }

    const changePrivacy = async () => {
      if (!isMine.value) return
      try {
        await api.setPrivacy(profile.value.profile_type)
      } catch (error) {
        console.error("Failed to change privacy:", error)
      }
    }

    const toggleFollow = async () => {
      if (!profile.value || isMine.value) return
      try {
        if (following.value) {
          // Unfollow
          await api.unfollow(profile.value.id)
          following.value = false
          pending.value = false
        } else {
          // Follow
          await api.follow(profile.value.id)

          // Immediately show "pending" if target is private
          if (profile.value.profile_type === 'private') {
            pending.value = true
            following.value = false
          } else {
            // For public profiles, automatically follow
            following.value = true
            pending.value = false
          }
        }

        // Refresh counts safely
        const f = await api.getFollowers(userId.value)
        followers.value = Array.isArray(f) ? f : []
        const fl = await api.getFollowing(userId.value)
        followingList.value = Array.isArray(fl) ? fl : []
      } catch (error) {
        console.error('Follow/unfollow error:', error)
      }
    }



    const refreshRequests = async () => {
      if (isMine.value) {
        try {
          followRequests.value = await api.listFollowRequests()
        } catch (error) {
          followRequests.value = []
        }
      }
    }

    const handleAccept = async (senderId) => {
      try {
        await api.acceptFollowRequest(senderId)
        await refreshRequests()
        followers.value = await api.getFollowers(userId.value)
      } catch (error) {
        console.error('Failed to accept follow request', error)
      }
    }

    const handleDecline = async (senderId) => {
      try {
        await api.declineFollowRequest(senderId)
        await refreshRequests()
      } catch (error) {
        console.error('Failed to decline follow request', error)
      }
    }

  watch(userId, load, { immediate: true })

  // Note: avatar upload UI was intentionally removed — avatars are URL-only now.
  return { profile, followers, followingList, followRequests, posts, following, pending, postCount, isMine, load, changePrivacy, toggleFollow, handleAccept, handleDecline, msg }
  }
}
</script>

<style scoped>
.profile-container {
  max-width: 1000px;
  margin: 0 auto;
}

.profile-header-card {
  border: none;
  border-radius: 1rem;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.profile-cover {
  height: 200px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
}

.avatar-img {
  width: 128px;
  height: 128px;
  border-radius: 50%;
  border: 4px solid white;
  object-fit: cover;
}

.profile-avatar-container {
  position: relative;
  margin-top: -75px;
  display: inline-block;
}

.edit-avatar-btn {
  position: absolute;
  bottom: 5px;
  right: 5px;
}

.private-profile-view .profile-avatar-container {
  margin-top: 0;
}
</style>
