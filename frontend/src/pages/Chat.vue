<template>
	<div class="container py-4 chat-wrapper">
		<div class="d-flex flex-column flex-md-row align-items-md-center justify-content-md-between mb-3 gap-2">
			<div>
				<h2 class="mb-1"><i class="fas fa-comments me-2"></i>Messenger</h2>
				<small class="text-muted">Stay in touch with your connections in real time.</small>
			</div>
			<div class="d-flex gap-2">
				<button class="btn btn-outline-primary" type="button" @click="connect" :disabled="connected">
					<i class="fas fa-plug me-1"></i>Connect
				</button>
				<button class="btn btn-outline-secondary" type="button" @click="disconnect" :disabled="!connected">
					<i class="fas fa-power-off me-1"></i>Disconnect
				</button>
				<button class="btn btn-outline-info" type="button" @click="refreshContacts" :disabled="!connected">
					<i class="fas fa-sync me-1"></i>Refresh
				</button>
			</div>
		</div>

		<div class="card shadow-sm border-0 chat-card">
			<div class="row g-0">
				<div class="col-lg-4 col-xl-3 border-end contacts-panel">
					<div class="px-3 py-3 border-bottom">
						<div class="input-group input-group-sm">
							<span class="input-group-text"><i class="fas fa-search"></i></span>
							<input v-model="search" type="text" class="form-control" placeholder="Search conversations" />
						</div>
					</div>
					<div class="contacts-list">
						<button
							v-for="contact in filteredContacts"
							:key="contact.id"
							type="button"
							class="contact-item btn btn-light text-start w-100"
							:class="{ active: contact.id === activeContactId }"
							@click="selectContact(contact.id)"
						>
							<div class="d-flex align-items-center">
								<div class="avatar me-3">
									<img v-if="contact.avatar" :src="contact.avatar" alt="avatar" />
									<div v-else class="placeholder">
										<i class="fas fa-user"></i>
									</div>
									<span class="status" :class="contact.isOnline ? 'online' : 'offline'"></span>
								</div>
								<div class="flex-grow-1">
									<div class="fw-semibold">{{ contact.displayName }}</div>
									<small class="text-muted">@{{ contact.nickname }}</small>
								</div>
								<div v-if="contact.unread" class="badge rounded-pill bg-primary ms-auto">{{ contact.unread }}</div>
							</div>
						</button>
						<div v-if="!filteredContacts.length" class="text-center text-muted py-4 small">
							No contacts yet. Start following people to chat with them.
						</div>
					</div>
				</div>

				<div class="col-lg-8 col-xl-9 position-relative">
					<div v-if="activeContact" class="conversation h-100 d-flex flex-column">
						<div class="conversation-header border-bottom px-4 py-3 d-flex align-items-center justify-content-between">
							<div class="d-flex align-items-center gap-3">
									<div class="avatar">
									<img v-if="activeContact.avatar" :src="activeContact.avatar" alt="avatar" />
									<div v-else class="placeholder">
										<i class="fas fa-user"></i>
									</div>
									<span class="status" :class="activeContact.isOnline ? 'online' : 'offline'"></span>
								</div>
								<div>
									<h5 class="mb-0">{{ activeContact.displayName }}</h5>
									<small class="text-muted">@{{ activeContact.nickname }}</small>
								</div>
							</div>
							<span class="badge" :class="connected ? 'bg-success' : 'bg-secondary'">
								{{ connected ? 'Connected' : 'Offline' }}
							</span>
						</div>

						<div class="conversation-body flex-grow-1" ref="messagePane">
							<div
								v-for="message in activeConversation"
								:key="message.id"
								class="message-row"
								:class="{ outgoing: message.outgoing }"
							>
								<div class="message-bubble">
									<p class="mb-1">{{ message.content }}</p>
									<small class="text-muted">{{ formatTimestamp(message.timestamp) }}</small>
								</div>
							</div>
						</div>

						<div class="conversation-footer border-top px-4 py-3">
							<div v-if="errorBanner" class="alert alert-danger py-2 px-3 small mb-3">
								<i class="fas fa-exclamation-circle me-2"></i>{{ errorBanner }}
							</div>
							<div class="input-group">
								<textarea
									ref="composer"
									v-model="draft"
									class="form-control"
									rows="2"
									placeholder="Write a message..."
									@keydown.enter.exact.prevent="send"
								></textarea>
								<button class="btn btn-primary" type="button" @click="send" :disabled="!draft.trim()">
									<i class="fas fa-paper-plane me-1"></i>Send
								</button>
							</div>
						</div>
					</div>

					<div v-else class="conversation-placeholder d-flex align-items-center justify-content-center text-center text-muted">
						<div>
							<i class="fas fa-comments fa-3x mb-3"></i>
							<p class="mb-0">Select someone from the list to start chatting.</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import { ref, computed, watch, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { useChatStore } from '@/store/chat'

export default {
	setup() {
		const chat = useChatStore()
		const search = ref('')
		const draft = ref('')
		const errorBanner = ref('')
		const messagePane = ref(null)
		let loadingMore = false

		const connected = computed(() => chat.connected)
		const activeConversation = computed(() => chat.activeConversation)
		const activeContact = computed(() => chat.activeContact)
		const activeContactId = computed(() => chat.activeContactId)
		const contacts = computed(() => chat.contacts)

		const filteredContacts = computed(() => {
			const term = search.value.trim().toLowerCase()
			if (!term) return contacts.value
			return contacts.value.filter((c) => {
				return (
					(c.displayName || '').toLowerCase().includes(term) ||
					(c.nickname || '').toLowerCase().includes(term)
				)
			})
		})

		const connect = () => chat.connect()
		const disconnect = () => chat.disconnect()
		const refreshContacts = () => chat.requestUserList()
		const selectContact = (id) => chat.setActiveContact(id)

		const send = () => {
			if (!draft.value.trim()) return
			const ok = chat.sendMessage({ type: 'message', content: draft.value, receiver_id: activeContactId.value })
			if (ok) {
				draft.value = ''
			} else {
				showNextError()
			}
		}

		const formatTimestamp = (ts) => {
			if (!ts) return ''
			const date = new Date(ts)
			if (Number.isNaN(date.getTime())) return ts
			return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
		}

		const showNextError = () => {
			const message = chat.nextError()
			if (!message) return
			errorBanner.value = message
			setTimeout(() => {
				if (errorBanner.value === message) {
					errorBanner.value = ''
				}
			}, 4000)
		}

		watch(
			() => chat.errors.length,
			(len) => {
				if (len > 0) {
					showNextError()
				}
			}
		)

		watch(
			activeConversation,
			async () => {
				await nextTick()
				if (messagePane.value) {
					messagePane.value.scrollTop = messagePane.value.scrollHeight
				}
			},
			{ deep: true }
		)

		const onScroll = async (e) => {
			const el = e.target
			if (!el) return
			if (el.scrollTop <= 40 && !loadingMore) {
				loadingMore = true
				const prevHeight = el.scrollHeight
				const added = await chat.loadMoreHistory(activeContactId.value)
				await nextTick()
				if (added > 0) {
					// preserve scroll position so content doesn't jump
					const newHeight = el.scrollHeight
					el.scrollTop = newHeight - prevHeight + el.scrollTop
				}
				loadingMore = false
			}
		}

		onMounted(() => {
			if (!chat.connected) {
				chat.connect()
			}
		})

		// attach scroll listener
		onMounted(() => {
			if (messagePane.value) {
				messagePane.value.addEventListener('scroll', onScroll)
			}
		})

		onBeforeUnmount(() => {
			if (chat.connected) {
				chat.disconnect()
			}
		})

		return {
			
			search,
			draft,
			errorBanner,
			messagePane,
			connected,
			contacts,
			activeConversation,
			activeContact,
			activeContactId,
			filteredContacts,
			connect,
			disconnect,
			refreshContacts,
			selectContact,
			send,
			formatTimestamp,
			onScroll,
		}
	},
}
</script>

<style scoped>
.chat-wrapper {
	max-width: 1100px;
}

.chat-card {
	height: 640px;
}

.contacts-panel {
	background: #f8f9ff;
}

.contacts-list {
	max-height: 520px;
	overflow-y: auto;
	padding: 12px;
	padding-top: 0;
}

.contact-item {
	margin-top: 12px;
	border: none;
	background: #fff;
	box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
	transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.contact-item.active,
.contact-item:hover {
	transform: translateY(-2px);
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.contact-item .avatar {
	position: relative;
	width: 48px;
	height: 48px;
}

.contact-item .avatar img,
.conversation-header .avatar img {
	width: 48px;
	height: 48px;
	object-fit: cover;
	border-radius: 50%;
}

.contact-item .avatar .placeholder,
.conversation-header .avatar .placeholder {
	width: 48px;
	height: 48px;
	border-radius: 50%;
	background: rgba(118, 75, 162, 0.15);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #764ba2;
}

.status {
	position: absolute;
	bottom: 2px;
	right: 2px;
	width: 12px;
	height: 12px;
	border-radius: 50%;
	border: 2px solid #fff;
}

.status.online {
	background: #28a745;
}

.status.offline {
	background: #adb5bd;
}

.conversation {
	background: #fff;
}

.conversation-body {
	padding: 24px;
	background: linear-gradient(180deg, #f7f8ff 0%, #ffffff 100%);
	overflow-y: auto;
	height: calc(640px - 140px); /* account for header/footer sizes */
}

.message-row {
	display: flex;
	margin-bottom: 16px;
}

.message-row.outgoing {
	justify-content: flex-end;
}

.message-bubble {
	max-width: 70%;
	background: #fff;
	border-radius: 18px;
	padding: 12px 16px;
	box-shadow: 0 6px 16px rgba(102, 126, 234, 0.15);
}

.message-row.outgoing .message-bubble {
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: #fff;
	box-shadow: 0 6px 20px rgba(76, 110, 245, 0.3);
}

.conversation-footer textarea {
	resize: none;
}

.conversation-placeholder {
	min-height: 520px;
}

@media (max-width: 991.98px) {
	.contacts-panel {
		border-bottom: 1px solid rgba(0, 0, 0, 0.05);
	}
	.contacts-list {
		max-height: 220px;
	}
	.conversation-header {
		flex-wrap: wrap;
		gap: 12px;
	}
}
</style>

