import api from './index'

export function listGroups() {
  return api.get('/api/groups')
}

export function createGroup(data) {
  return api.post('/api/group/create', data)
}

export function getGroup(id) {
  return api.get('/api/group?id=' + id)
}

export function inviteToGroup(payload) {
  return api.post('/api/group/invite', payload)
}

export function respondInvite(payload) {
  return api.post('/api/group/invite/respond', payload)
}

export function listGroupPosts(group_id) {
  return api.get('/api/group/posts?group_id=' + group_id)
}

export function checkMembership(group_id) {
  return api.get('/api/group/membership?group_id=' + group_id)
}

export function requestToJoin(group_id) {
  return api.post('/api/group/request', { group_id })
}

export function getRequestStatus(group_id) {
  return api.get('/api/group/request/status?group_id=' + group_id)
}

export function respondRequest(payload) {
  return api.post('/api/group/request/respond', payload)
}

export function listRequests(group_id) {
  return api.get('/api/group/requests?group_id=' + group_id)
}

export function createGroupPost(formData) {
  return api.post('/api/group/post/create', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
}

export function addGroupComment(payload) {
  return api.post('/api/group/comment', payload)
}

export function listGroupComments(post_id) {
  return api.get(`/api/group/comments?post_id=${post_id}`)
}


export function createEvent(payload) {
  return api.post('/api/group/event/create', payload)
}

export function voteEvent(payload) {
  return api.post('/api/group/event/vote', payload)
}

export function listEvents(group_id) {
  return api.get('/api/group/events?group_id=' + group_id)
}

export function fetchGroupMessages(group_id, opts = {}) {
  const params = new URLSearchParams()
  params.set('group_id', String(group_id))
  if (opts.before_id) params.set('before_id', String(opts.before_id))
  if (opts.limit) params.set('limit', String(opts.limit))
  return api.get('/api/group/messages?' + params.toString())
}
