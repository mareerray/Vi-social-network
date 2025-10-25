<template>
  <div class="container mt-4">
    <div v-if="group.group">
      <h2>{{ group.group.name || 'Group' }}</h2>
      <p v-if="group.group.description">{{ group.group.description }}</p>
    </div>
    <div class="mb-3">
      <h5>Members: {{ group.members }}</h5>
    </div>

    <!-- Members list -->
    <div v-if="group.members_list && group.members_list.length > 0" class="card mb-3">
      <div class="card-body">
        <h5>Members</h5>
        <div class="d-flex flex-wrap">
          <div v-for="m in group.members_list" :key="m.id" class="me-2 mb-2 text-center member-tile">
            <div>
              <img v-if="memberAvatar(m.avatar)" :src="memberAvatar(m.avatar)" alt="avatar" class="rounded-circle" style="width:64px;height:64px;object-fit:cover;" />
              <i v-else class="fas fa-user-circle text-primary" style="font-size:64px;"></i>
            </div>
            <div class="small mt-1 member-name" style="word-break:break-word">{{ m.nickname || ('User ' + m.id) }}</div>
          </div>
        </div>
      </div>
    </div>

        <!-- Invite users (visible to members and owner) -->
    <div v-if="group.group && isMember || isOwner" class="card mb-3">
      <div class="card-body">
        <h5>Invite Users to Group</h5>
        <div v-if="followers.length === 0" class="text-muted mb-2">
          No users available to invite
        </div>
        <div class="input-group mb-3">
          <input type="text" class="form-control" v-model="userFilter" placeholder="Search users...">
        </div>
        <div class="list-group">
          <div v-for="f in filteredUsers" :key="f.id" class="list-group-item d-flex justify-content-between align-items-center">
            <div class="d-flex align-items-center">
              <img v-if="memberAvatar(f.avatar)" :src="memberAvatar(f.avatar)" alt="avatar" class="rounded-circle me-2" style="width:32px;height:32px;object-fit:cover;" />
              <i v-else class="fas fa-user-circle text-primary me-2" style="font-size:32px;"></i>
              <div>
                <strong>{{ f.nickname || ('User ' + f.id) }}</strong>
                <div v-if="f.full_name" class="small text-muted">{{ f.full_name }}</div>
              </div>
            </div>
            <button 
              class="btn btn-sm" 
              :class="invitingIds.includes(f.id) ? 'btn-secondary' : 'btn-primary'"
              :disabled="invitingIds.includes(f.id)"
              @click="invite(f.id)">
              {{ invitingIds.includes(f.id) ? 'Inviting...' : 'Invite' }}
            </button>
          </div>
        </div>
      </div>
    </div>
    <!-- Create Post -->
    <div v-if="isMember" class="card mb-3">
      <div class="card-body">
        <h5>Create Post</h5>
        <textarea v-model="content" class="form-control" placeholder="What's on your mind?"></textarea>
        <input type="file" ref="file" class="form-control mt-2" />
        <button class="btn btn-primary mt-2" @click="createPost">Post</button>
      </div>
    </div>
    <div v-else class="alert alert-warning">
      <div>You are not a member of this group.</div>
      <div class="mt-2">
        <button class="btn btn-sm btn-primary" @click="requestJoin" :disabled="hasPendingRequest">{{ hasPendingRequest ? 'Request pending' : 'Request to join' }}</button>
      </div>
    </div>

  <!-- Pending join requests (owner only) -->
  <div v-if="requests && requests.length > 0 && isOwner" class="card mb-3">
      <div class="card-body">
        <h5>Pending Join Requests</h5>
        <div class="list-group">
          <div v-for="r in requests" :key="r.id" class="list-group-item d-flex justify-content-between align-items-center">
            <div class="d-flex align-items-center">
              <img v-if="r.avatar" :src="memberAvatar(r.avatar)" alt="avatar" class="rounded-circle me-2" style="width:32px;height:32px;object-fit:cover;" />
              <div>
                <strong>{{ r.nickname || ('User ' + r.requester_id) }}</strong>
                <div class="small text-muted">Requested at {{ r.created_at }}</div>
              </div>
            </div>
            <div>
              <button :disabled="processingRequestIds.includes(r.id)" class="btn btn-sm btn-success me-2" @click="handleRequest(r.id, 'accept')">
                <span v-if="processingRequestIds.includes(r.id)" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                Accept
              </button>
              <button :disabled="processingRequestIds.includes(r.id)" class="btn btn-sm btn-outline-secondary" @click="handleRequest(r.id, 'decline')">
                <span v-if="processingRequestIds.includes(r.id)" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                Decline
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-for="p in posts" :key="p.id" class="card mb-2">
      <div class="card-body">
        <p>{{ p.content }}</p>
  <img v-if="p.image_url" :src="`http://localhost:8080${p.image_url}`" class="img-fluid" />
        <div class="mt-2">
          <Comment :postId="p.id" :isGroup="true" @comment-added="load" />
        </div>
      </div>
    </div>
    <!-- Events -->
    <div class="card mb-3" v-if="isMember">
      <div class="card-body">
        <h5>Create Event</h5>
        <div class="mb-2">
          <input v-model="newEvent.title" class="form-control mb-2" placeholder="Title" />
          <textarea v-model="newEvent.description" class="form-control mb-2" placeholder="Description"></textarea>
          <input type="datetime-local" v-model="newEvent.event_time" class="form-control mb-2" />
          <button class="btn btn-primary" @click="createEvent">Create Event</button>
        </div>
      </div>
    </div>

    <div v-if="events && events.length > 0" class="mb-3">
      <h5>Events</h5>
      <div v-for="e in events" :key="e.id" class="card mb-2">
        <div class="card-body">
          <h6>{{ e.title }}</h6>
          <p>{{ e.description }}</p>
          <div><strong>When:</strong> {{ e.event_time || 'TBD' }}</div>
          <div class="mt-2">
            <button class="btn btn-sm" :class="{ 'btn-success': e.my_vote==='going' }" @click="vote(e.id, 'going')">Going ({{ e.votes && e.votes.going ? e.votes.going : 0 }})</button>
            <button class="btn btn-sm ms-2" :class="{ 'btn-secondary': e.my_vote==='not_going' }" @click="vote(e.id, 'not_going')">Not going ({{ e.votes && e.votes.not_going ? e.votes.not_going : 0 }})</button>
          </div>
        </div>
      </div>
    </div>
    <!-- Group Chat -->
    <div v-if="isMember" class="card mb-3">
      <div class="card-body d-flex flex-column" style="height: 400px;">
        <h5>Group Chat</h5>
        <div ref="chatBody" class="flex-grow-1 border rounded p-2 mb-2" style="overflow:auto; background:#f8f9fa">
          <div v-if="loadingGroupMessages" class="text-center text-muted">Loading messages...</div>
          <div v-for="m in groupMessages" :key="m.id" class="message-row mb-2" :class="{ outgoing: m.outgoing }">
            <div class="message-bubble">
              <div class="message-meta d-flex align-items-center mb-1">
                <small class="sender-name text-muted">{{ m.senderName }}</small>
              </div>
              <div class="message-content" v-html="m.content"></div>
            </div>
          </div>
        </div>
        <div class="mb-2 text-center">
          <button class="btn btn-sm btn-link" @click="loadMoreMessages">Load more messages</button>
        </div>
        <div class="d-flex align-items-center">
          <div class="me-2 position-relative">
            <button class="btn btn-light" type="button" @click="showEmojiPicker = !showEmojiPicker">😀</button>
            <div v-if="showEmojiPicker" class="position-absolute" style="z-index:1000; background:white; border:1px solid #ddd; padding:6px; display:flex; flex-wrap:wrap; width:200px">
              <button v-for="e in emojiList" :key="e" class="btn btn-sm btn-white m-1" @click="insertEmoji(e)" style="min-width:32px">{{ e }}</button>
            </div>
          </div>
          <input ref="chatInput" v-model="chatInput" @keyup.enter="sendGroupMessage" class="form-control me-2" placeholder="Write a message — you can paste emoji" />
          <button class="btn btn-primary" @click="sendGroupMessage">Send</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getGroup, listGroupPosts, createGroupPost, addGroupComment, inviteToGroup, checkMembership, requestToJoin, respondRequest, listRequests, getRequestStatus, createEvent, voteEvent, listEvents } from '../api/groups'
import Comment from '@/components/Comment.vue'
import { resolveAsset } from '@/utils/resolveUrl'
import { listUsers } from '@/api/users'
import { useAuthStore } from '@/store/auth'
import { useChatStore } from '@/store/chat'

export default {
  components: { Comment },
  computed: {
    filteredUsers() {
      if (!this.userFilter) return this.followers
      const filter = this.userFilter.toLowerCase()
      return this.followers.filter(user => 
        (user.nickname || '').toLowerCase().includes(filter) ||
        (user.full_name || '').toLowerCase().includes(filter)
      )
    }
  },
  data() { return { group: {}, posts: [], content: '', commentText: {}, followers: [], requests: [], invitingIds: [], hasPendingRequest: false, isMember: false, isOwner: false, events: [], userFilter: '', newEvent: { title: '', description: '', event_time: '' }, // chat
      chatInput: '',
      groupMessages: [],
      loadingGroupMessages: false,
      showEmojiPicker: false,
      emojiList: ['😄','❤️','👍','😂','😢','🎉','👏','🔥','😎','🙌'],
      processingRequestIds: []
    } },
  created() { this.load() },
  mounted() {
    this._onInviteResponded = (e) => {
      try {
        const gid = e && e.detail && e.detail.group_id
        if (gid && String(gid) === String(this.$route.params.id)) {
          // reload group state
          this.load()
        }
      } catch (err) {}
    }
    window.addEventListener('group-invite-responded', this._onInviteResponded)
  },
  beforeUnmount() {
    if (this._onInviteResponded) window.removeEventListener('group-invite-responded', this._onInviteResponded)
    if (this._unwatch) this._unwatch()
  },
  methods: {
    async load() {
      const id = this.$route.params.id
      const g = await getGroup(id)
      this.group = g.data
      console.log('Group data:', this.group) // Log full group data to see structure
      
      // determine ownership
      const auth = useAuthStore()
      this.isOwner = auth.user && this.group && this.group.group && Number(auth.user.user_id) === Number(this.group.group.owner_id)
      
      // check membership early so we know whether to load followers
      await this.checkMembershipAPI(id)
      
      // Get list of current members
      const membersList = this.group && this.group.members_list ? this.group.members_list :
                         this.group && this.group.group && this.group.group.members ? this.group.group.members :
                         []
      console.log('Current group members:', membersList)

      // if current user is a member or owner, load all users that can be invited
      try {
        if (auth.user && (this.isMember || this.isOwner)) {
          const res = await listUsers()
          console.log('Full API response:', res)
          
          let memberIds = []
          try {
            // Prefer top-level members_list (added server-side). Fall back to nested group.members_list.
            const ml = (this.group && Array.isArray(this.group.members_list)) ? this.group.members_list :
                       (this.group && this.group.group && Array.isArray(this.group.group.members_list) ? this.group.group.members_list : [])
            memberIds = ml.map(m => Number((m && (m.id || m)) || m))
            // also exclude the owner explicitly
            if (this.group && this.group.group && this.group.group.owner_id) {
              memberIds.push(Number(this.group.group.owner_id))
            }
            // dedupe
            memberIds = Array.from(new Set(memberIds.filter(id => !Number.isNaN(id))))
            console.log('Group members from group data:', memberIds)
          } catch (err) {
            console.warn('Could not get members from group data:', err)
          }
          
          // Ensure we have users data in the correct format
          const currentUserId = Number(auth.user.user_id)
          const users = Array.isArray(res) ? res : 
                       Array.isArray(res.data) ? res.data :
                       []
          console.log('All users:', users)
          
          // Filter out current user and existing members
          this.followers = users.filter(user => {
            if (!user || !user.id) return false
            const userId = Number(user.id)
            const isAvailable = userId !== currentUserId && !memberIds.includes(userId)
            console.log(`User ${userId}: isAvailable=${isAvailable} (not current user && not member)`)
            return isAvailable
          })
          
          console.log('Filtered users (available to invite):', this.followers)
        }
      } catch (e) {
        console.error('Failed to load users:', e)
        this.followers = []
      }

      const postsRes = await listGroupPosts(id)
      this.posts = postsRes.data
      // init chat if member
      if (this.isMember) {
        this.initGroupChat()
      }
      // check if current user has a pending request (non-owner only)
      try {
        if (!this.isOwner) {
          const s = await getRequestStatus(id)
          this.hasPendingRequest = s && s.data && s.data.has_pending
        } else {
          this.hasPendingRequest = false
        }
      } catch (e) {
        this.hasPendingRequest = false
      }
      // if owner, load pending requests
      try {
        if (this.isOwner) {
          const res = await listRequests(id)
          this.requests = Array.isArray(res.data) ? res.data : []
        }
      } catch (e) {
        this.requests = []
      }
      // load events for group (members only)
      try {
        if (this.isMember) {
          const ev = await this.loadEvents()
          this.events = Array.isArray(ev) ? ev : []
        } else {
          this.events = []
        }
      } catch (e) {
        this.events = []
      }
    },
    async checkMembershipAPI(id) {
      try {
        const res = await checkMembership(id)
        this.isMember = res.data && res.data.is_member
      } catch (e) {
        this.isMember = false
      }
    },

    async createPost() {
      if (!this.isMember) {
        alert('You must be a member of this group to post. Accept an invite first.')
        return
      }
      const fd = new FormData()
      fd.append('group_id', this.$route.params.id)
      fd.append('content', this.content)
      const f = this.$refs.file && this.$refs.file.files && this.$refs.file.files[0]
      if (f) fd.append('image', f)
      try {
        await createGroupPost(fd)
        this.content = ''
        await this.load()
      } catch (err) {
        console.error('Failed to create group post', err)
        alert('Failed to create post: ' + (err?.response?.data?.error || err?.message || 'unknown'))
      }
    },
    async loadEvents() {
      const id = this.$route.params.id
      const res = await listEvents(id)
      return res.data
    },
    async createEvent() {
      if (!this.isMember) { alert('Only members can create events'); return }
      try {
        await createEvent({ group_id: Number(this.$route.params.id), title: this.newEvent.title, description: this.newEvent.description, event_time: this.newEvent.event_time })
        this.newEvent = { title: '', description: '', event_time: '' }
        await this.load()
        alert('Event created')
      } catch (e) { console.error(e); alert('Failed to create event') }
    },
    async vote(eventId, vote) {
      try {
        await voteEvent({ event_id: eventId, vote })
        await this.load()
      } catch (e) { console.error(e); alert('Failed to vote') }
    },
    async invite(userId) {
      if (this.invitingIds.includes(userId)) return
      this.invitingIds.push(userId)
      try {
        const res = await inviteToGroup({ group_id: Number(this.$route.params.id), invitee_id: userId })
        // if server says already_pending, inform user
        if (res && res.data && res.data.status === 'already_pending') {
          alert('Invite already pending for this user')
        } else {
          alert('Invited')
          // remove invited user from followers list to avoid re-inviting
          this.followers = this.followers.filter(f => f.id !== userId)
        }
      } catch (e) {
        console.error('Invite failed', e)
        alert('Failed to invite')
      } finally {
        this.invitingIds = this.invitingIds.filter(id => id !== userId)
      }
    },
    memberAvatar(avatar) {
      // resolveAsset returns full URL or empty string
      return resolveAsset(avatar) || ''
    },
    async requestJoin() {
      try {
        await requestToJoin(Number(this.$route.params.id))
        alert('Join request sent')
      } catch (e) {
        console.error(e)
        alert('Failed to send request')
      }
    },
    async handleRequest(requestId, action) {
      // optimistic UI: mark processing and remove from list locally
      if (this.processingRequestIds.includes(requestId)) return
      this.processingRequestIds.push(requestId)
      const originalRequests = [...this.requests]
      // remove immediately so owner sees it disappear
      this.requests = this.requests.filter(r => r.id !== requestId)
      try {
        const res = await respondRequest({ request_id: requestId, action })
        // if accepted, increment local member count (group.members)
        if (action === 'accept') {
          try {
            if (this.group && typeof this.group.members === 'number') {
              this.group.members = Number(this.group.members) + 1
            }
          } catch (e) {}
        }
        alert(res && res.data && res.data.status ? (res.data.status) : 'Done')
      } catch (e) {
        // rollback on error
        console.error('Respond request failed', e)
        this.requests = originalRequests
        alert('Failed to process request')
      } finally {
        this.processingRequestIds = this.processingRequestIds.filter(id => id !== requestId)
      }
    },
    addComment(postId) {
      addGroupComment({ post_id: postId, content: this.commentText[postId] }).then(()=> { this.commentText[postId] = ''; this.load() })
    }
    ,
    // --- group chat methods ---
    async initGroupChat() {
      if (!this.isMember) return
      // use pinia chat store (reactive)
      this.chatStore = useChatStore()
      if (!this.chatStore.connected) this.chatStore.connect()
      this.loadingGroupMessages = true
      try {
        await this.chatStore.fetchGroupHistory(this.$route.params.id)
        this.groupMessages = this.chatStore.groupConversations[String(this.$route.params.id)] || []
      } catch (e) {
        console.error('Failed to load group messages', e)
        this.groupMessages = []
      } finally {
        this.loadingGroupMessages = false
      }

      // reactive subscription: rely on Pinia reactivity
      this._unwatch = this.$watch(() => this.chatStore.groupConversations[String(this.$route.params.id)], (nv) => {
        this.groupMessages = nv || []
        this.$nextTick(() => this.scrollChatBottom())
      }, { immediate: true })
    },
    async sendGroupMessage() {
      if (!this.chatInput || !this.chatInput.trim()) return
      if (!this.chatStore) {
        this.chatStore = useChatStore()
        if (!this.chatStore.connected) this.chatStore.connect()
      }
      const sent = this.chatStore.sendGroupMessage({ group_id: Number(this.$route.params.id), content: this.chatInput })
      if (sent) {
        this.chatInput = ''
      }
    },
    async loadMoreMessages() {
      if (!this.chatStore) this.chatStore = useChatStore()
      const added = await this.chatStore.loadMoreGroupHistory(this.$route.params.id, 50)
      if (added > 0) {
        // keep scroll position roughly in place: naive approach
        this.$nextTick(() => {})
      }
    },
    insertEmoji(emoji) {
      const el = this.$refs.chatInput
      if (el && el.selectionStart !== undefined) {
        const start = el.selectionStart
        const end = el.selectionEnd
        const val = this.chatInput || ''
        this.chatInput = val.slice(0, start) + emoji + val.slice(end)
        this.$nextTick(() => {
          el.selectionStart = el.selectionEnd = start + emoji.length
          el.focus()
        })
      } else {
        this.chatInput = (this.chatInput || '') + emoji
      }
    },
    scrollChatBottom() {
      const el = this.$refs.chatBody
      if (el) el.scrollTop = el.scrollHeight
    }
  }
}
</script>

<style scoped>
.message-row {
  display: flex;
  width: 100%;
}
.message-row .message-bubble {
  max-width: 75%;
  padding: 10px 12px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 2px 6px rgba(0,0,0,0.04);
}
.message-row.outgoing {
  justify-content: flex-end;
}
.message-row.outgoing .message-bubble {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}
.message-meta .sender-name {
  font-weight: 600;
}
.message-content {
  word-break: break-word;
}
.card .flex-grow-1 > .message-row:first-child {
  margin-top: 6px;
}
</style>
