<template>
	<nav class="navbar navbar-expand-lg navbar-dark bg-primary shadow-sm">
		<div class="container-fluid">
			<router-link class="navbar-brand fw-bold text-white" to="/">
				<i class="fas fa-users me-2"></i>Social Network
			</router-link>

			<button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
				<span class="navbar-toggler-icon"></span>
			</button>

			<div class="collapse navbar-collapse" id="navbarNav">
				<ul class="navbar-nav me-auto" v-if="user">
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/">
							<i class="fas fa-home me-1"></i>Home
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/profile">
							<i class="fas fa-user me-1"></i>Profile
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/people">
							<i class="fas fa-user-friends me-1"></i>People
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/chat">
							<i class="fas fa-comments me-1"></i>Chat
						</router-link>
					</li>
					<li class="nav-item">
						<router-link class="nav-link text-white" to="/groups">
							<i class="fas fa-users-cog me-1"></i>Groups
						</router-link>
					</li>
				</ul>

				<div class="d-flex align-items-center" v-if="user">
					<!-- Notifications -->
					<div class="dropdown me-3" :class="{ show: open }">
						<button class="btn btn-outline-light position-relative" @click="toggleOpen" :aria-expanded="open">
							<i class="fas fa-bell"></i>
							<span v-if="unreadCount" class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">
								{{ unreadCount }}
							</span>
						</button>
						<div class="dropdown-menu dropdown-menu-end p-0" :class="{ show: open }" style="width: 320px;">
							<div class="dropdown-header bg-light">
								<strong>Notifications</strong>
							</div>
							<div v-if="!notifications || notifications.length === 0" class="dropdown-item-text text-muted text-center py-3">
								No notifications
							</div>
							<div v-else>
								<div v-for="n in notifications" :key="n.id" 
									 class="dropdown-item border-bottom" 
									 :class="{ 'bg-light': !n.is_read }"
									 @click.prevent="openNotification(n)">
									<div class="d-flex justify-content-between align-items-start">
										<span class="fw-semibold text-primary">{{ n.type }}</span>
										<small class="text-muted">{{ formatTime(n.created_at) }}</small>
									</div>
									<div class="small text-muted mt-1">{{ (parseData(n) && parseData(n).preview) ? parseData(n).preview : '' }}</div>
									<!-- Follow Request Notification -->
									<div v-if="n.type === 'follow_request'" class="mt-2">
										<div class="d-flex align-items-center">
											<i class="fas fa-user-plus text-primary me-2"></i>
											<span>
											<strong>{{ parseData(n).actor_nickname || 'Someone' }}</strong>
											wants to follow you.
											</span>
										</div>

										<div class="mt-2">
											<button
											class="btn btn-sm btn-success me-2"
											@click.stop.prevent="handleFollowRequest(parseData(n), 'accept')"
											>
											Accept
											</button>
											<button
											class="btn btn-sm btn-outline-secondary"
											@click.stop.prevent="handleFollowRequest(parseData(n), 'decline')"
											>
											Decline
											</button>
										</div>
									</div>

									<!-- Follow Request Accepted -->
									<div v-else-if="n.type === 'follow_request_accepted'" class="mt-2 text-success">
									<i class="fas fa-check-circle me-2"></i>
									<strong>{{ parseData(n).actor_nickname || 'Someone' }}</strong>
									accepted your follow request.
									</div>

									<!-- Follow Request Declined -->
									<div v-else-if="n.type === 'follow_request_declined'" class="mt-2 text-danger">
									<i class="fas fa-times-circle me-2"></i>
									<strong>{{ parseData(n).actor_nickname || 'Someone' }}</strong>
									declined your follow request.
									</div>

									<div v-if="n.type === 'group_invite'" class="mt-2">
										<button class="btn btn-sm btn-success me-2" @click.stop.prevent="(() => { const d = parseData(n); if (d) respondToInvite(d.invite_id, 'accept', d.group_id) })()">Accept</button>
										<button class="btn btn-sm btn-outline-secondary" @click.stop.prevent="(() => { const d = parseData(n); if (d) respondToInvite(d.invite_id, 'decline', d.group_id) })()">Decline</button>
									</div>
									<button v-if="!n.is_read" 
										@click.stop.prevent="markRead(n.id)" 
										class="btn btn-sm btn-outline-primary mt-2">
										Mark read
									</button>
								</div>
								<div class="dropdown-footer p-2 border-top">
									<button class="btn btn-sm btn-primary w-100" @click="markAll">
										Mark all read
									</button>
								</div>
							</div>
						</div>
					</div>

					<!-- User Profile -->
					<div class="dropdown">
						<button class="btn btn-outline-light dropdown-toggle d-flex align-items-center" 
							@click="profileOpen = !profileOpen" aria-expanded="false">
					   <img v-if="user && user.avatar" :src="user.avatar" alt="avatar" 
								 class="rounded-circle me-2" style="width: 24px; height: 24px;" />
							<i v-else class="fas fa-user-circle me-2"></i>
							{{ user.nickname || user.user_id || "You"}}
						</button>
						<ul class="dropdown-menu dropdown-menu-end" v-show="profileOpen">
							<li><router-link class="dropdown-item" to="/profile">
								<i class="fas fa-user me-2"></i>My Profile
							</router-link></li>
							<li><router-link class="dropdown-item" to="/profile/edit">
								<i class="fas fa-edit me-2"></i>Edit Profile
							</router-link></li>
							<li><hr class="dropdown-divider"></li>
							<li><button class="dropdown-item text-danger" @click="onLogout">
								<i class="fas fa-sign-out-alt me-2"></i>Logout
							</button></li>
						</ul>
					</div>

					<!-- Logout quick button visible on all pages -->
					<button class="btn btn-outline-light ms-2" @click="onLogout" title="Logout">
						<i class="fas fa-sign-out-alt"></i>
					</button>
				</div>

				<div class="d-flex" v-else>
					<router-link class="btn btn-outline-light me-2" to="/login">
						<i class="fas fa-sign-in-alt me-1"></i>Login
					</router-link>
					<router-link class="btn btn-warning" to="/register">
						<i class="fas fa-user-plus me-1"></i>Register
					</router-link>
				</div>
			</div>
		</div>
	</nav>
</template>

<script>
import { defineComponent, ref, computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/store/auth'
import { useNotificationStore } from '@/store/notification'
import { respondInvite } from '@/api/groups'
import api from '@/api'


export default defineComponent({
	setup() {
		const auth = useAuthStore()
		const notif = useNotificationStore()

		const { user } = storeToRefs(auth)
		const { list: notifications } = storeToRefs(notif)

		// debug: log the actual user value (storeToRefs returns a ref)
		console.log("user is:", user.value)
		
		const unreadCount = computed(() => (notifications.value || []).filter(n => !n.is_read).length)

		const open = ref(false)
		const toggleOpen = async () => {
			open.value = !open.value
			if (open.value) await notif.fetch()
		}

		const profileOpen = ref(false)

		const markAll = async () => {
			await notif.markAllRead()
		}

		const markRead = async (id) => {
			await notif.markRead(id)
		}

		const respondToInvite = async (inviteId, action, groupId) => {
			try {
				await respondInvite({ invite_id: inviteId, action })
				await notif.fetch()
				// let other parts of the app know an invite was handled so they can refresh
				try { window.dispatchEvent(new CustomEvent('group-invite-responded', { detail: { invite_id: inviteId, action, group_id: groupId } })) } catch(e) {}
			} catch (e) {
				console.error('Failed to respond to invite', e)
			}
		}

		const handleFollowRequest = async (data, action) => {
			try {
				console.log('📦 Raw notification data:', data)
				// Extract sender_id from notification data
				const senderId = data.follower_id
				console.log('🔑 Extracted senderId:', senderId)
				
				if (!senderId) {
				console.error('Missing follower_id:', data)
				return
				}

				const endpoint = action === 'accept' 
				? '/follow/accept' 
				: '/follow/decline'
				
				console.log('📤 Sending to backend:', { sender_id: senderId })

				// Send { sender_id: ... } to backend
				await api.post(endpoint, { sender_id: senderId })
				
				// Refresh notifications
				await notif.fetch()
				
				console.log(`✅ Follow request ${action}ed`)
			} catch (e) {
				console.error('❌ Failed:', e)
				console.error('Error response:', e.response?.data)
			}
		}


		const onLogout = async () => {
			await auth.logout()
			location.href = '/'
		}

		const parseData = (n) => {
			if (!n || !n.data) return null
			try { return JSON.parse(n.data) } catch (e) { return null }
		}

		const formatTime = (ts) => {
			if (!ts) return ''
			const d = new Date(ts)
			if (isNaN(d.getTime())) return ts
			return d.toLocaleString()
		}

		const openNotification = (n) => {
			const d = parseData(n)
			let url = '/'
			if (d && d.url) url = d.url
			markRead(n.id).then(() => {
				toggleOpen()
				window.location.href = url
			})
		}

		return { user, notifications, unreadCount, open, toggleOpen, markAll, markRead, onLogout, profileOpen, respondToInvite, parseData, formatTime, openNotification, handleFollowRequest }
	}
})
</script>

<style scoped>
.navbar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
}

.navbar-brand {
  font-size: 1.5rem;
  font-weight: 700;
}

.nav-link {
  font-weight: 500;
  transition: color 0.3s ease;
}

.nav-link:hover {
  color: #ffc107 !important;
}

.dropdown-menu {
  border: none;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  border-radius: 0.5rem;
}

.dropdown-item:hover {
  background-color: #f8f9fa;
}

.btn-outline-light:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.badge {
  font-size: 0.6rem;
}

/* Notifications: ensure long text wraps and dropdown scrolls instead of overlapping */
.dropdown-menu {
	max-height: 60vh;
	overflow-y: auto;
}
.dropdown-item {
	white-space: normal !important;
	word-break: break-word;
	overflow-wrap: anywhere;
}
.dropdown-item .fw-semibold {
	display: block;
	margin-bottom: 0.25rem;
}
.dropdown-item .small {
	display: block;
}
.dropdown-footer {
	background: #fff;
}
</style>

