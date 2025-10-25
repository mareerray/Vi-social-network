import api from './index'

export function listGroups() {
  return api.get('/groups')
}

export function createGroup(data) {
  return api.post('/group/create', data)
}

export function getGroup(id) {
  return api.get('/group?id=' + id)
}

export function inviteToGroup(payload) {
  return api.post('/group/invite', payload)
}

export function respondInvite(payload) {
  return api.post('/group/invite/respond', payload)
}

export function listGroupPosts(group_id) {
  return api.get('/group/posts?group_id=' + group_id)
}

export function checkMembership(group_id) {
  return api.get('/group/membership?group_id=' + group_id)
}

export function requestToJoin(group_id) {
  return api.post('/group/request', { group_id })
}

export function getRequestStatus(group_id) {
  return api.get('/group/request/status?group_id=' + group_id)
}

export function respondRequest(payload) {
  return api.post('/group/request/respond', payload)
}

export function listRequests(group_id) {
  return api.get('/group/requests?group_id=' + group_id)
}

export function createGroupPost(formData) {
  return api.post('/group/post/create', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
}

export function addGroupComment(payload) {
  return api.post('/group/comment', payload)
}

export function listGroupComments(post_id) {
  return api.get(`/group/comments?post_id=${post_id}`)
}


export function createEvent(payload) {
  return api.post('/group/event/create', payload)
}

export function voteEvent(payload) {
  return api.post('/group/event/vote', payload)
}

export function listEvents(group_id) {
  return api.get('/group/events?group_id=' + group_id)
}

export function fetchGroupMessages(group_id, opts = {}) {
  const params = new URLSearchParams()
  params.set('group_id', String(group_id))
  if (opts.before_id) params.set('before_id', String(opts.before_id))
  if (opts.limit) params.set('limit', String(opts.limit))
  return api.get('/group/messages?' + params.toString())
}
