import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080',
  withCredentials: true,
});

export const getProfile = async (id) => {
  const url = id ? `/api/profile/${id}` : '/api/profile';
  const res = await apiClient.get(url);
  return res.data;
}

export const updateProfile = async (payload) => {
  const res = await apiClient.post('/api/profile/update', payload);
  return res.data;
}

export const uploadAvatar = async (file) => {
  const fd = new FormData();
  fd.append('file', file);
  fd.append('type', 'avatar');
  const res = await apiClient.post('/api/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } });
  return res.data; // expects { url: '/uploads/avatars/...' }
}

export const getFollowers = async (id) => {
  const url = id ? `/api/profile/followers?id=${id}` : '/api/profile/followers';
  const res = await apiClient.get(url);
  return res.data;
}

export const getFollowing = async (id) => {
  const url = id ? `/api/profile/following?id=${id}` : '/api/profile/following';
  const res = await apiClient.get(url);
  return res.data;
}

export const listFollowRequests = async () => {
  const res = await apiClient.get('/api/follow/requests');
  return res.data;
}

export const getFollowStatus = async (targetId) => {
  const res = await apiClient.get(`/api/follow/status?target_id=${targetId}`);
  return res.data;
}

export const acceptFollowRequest = async (senderId) => {
  const res = await apiClient.post('/api/follow/accept', { sender_id: senderId });
  return res.data;
}

export const declineFollowRequest = async (senderId) => {
  const res = await apiClient.post('/api/follow/decline', { sender_id: senderId });
  return res.data;
}

export const setPrivacy = async (profile_type) => {
  const res = await apiClient.post('/api/profile/privacy', { profile_type });
  return res.data;
}

export const follow = async (targetId) => {
  const res = await apiClient.post('/api/follow', { target_id: targetId });
  return res.data;
}

export const unfollow = async (targetId) => {
  const res = await apiClient.post('/api/unfollow', { target_id: targetId });
  return res.data;
}

export const listUsers = async () => {
  const res = await apiClient.get('/api/users');
  return res.data;
}
