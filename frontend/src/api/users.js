import axios from 'axios';
import api from '.';

/* const apiClient = axios.create({
  baseURL: 'http://localhost:8080',
  withCredentials: true,
});
 */
export const getProfile = async (id) => {
  const url = id ? `/profile/${id}` : '/profile';
  const res = await api.get(url);
  return res.data;
}

export const updateProfile = async (payload) => {
  const res = await api.post('/profile/update', payload);
  return res.data;
}

export const uploadAvatar = async (file) => {
  const fd = new FormData();
  fd.append('file', file);
  fd.append('type', 'avatar');
  const res = await api.post('/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } });
  return res.data; // expects { url: '/uploads/avatars/...' }
}

export const getFollowers = async (id) => {
  const url = id ? `/profile/followers?id=${id}` : '/profile/followers';
  const res = await api.get(url);
  return res.data;
}

export const getFollowing = async (id) => {
  const url = id ? `/profile/following?id=${id}` : '/profile/following';
  const res = await api.get(url);
  return res.data;
}

export const listFollowRequests = async () => {
  const res = await api.get('/follow/requests');
  return res.data;
}

export const getFollowStatus = async (targetId) => {
  const res = await api.get(`/follow/status?target_id=${targetId}`);
  return res.data;
}

export const acceptFollowRequest = async (senderId) => {
  const res = await api.post('/follow/accept', { sender_id: senderId });
  return res.data;
}

export const declineFollowRequest = async (senderId) => {
  const res = await api.post('/follow/decline', { sender_id: senderId });
  return res.data;
}

export const setPrivacy = async (profile_type) => {
  const res = await api.post('/profile/privacy', { profile_type });
  return res.data;
}

export const follow = async (targetId) => {
  const res = await api.post('/follow', { target_id: targetId });
  return res.data;
}

export const unfollow = async (targetId) => {
  const res = await api.post('/unfollow', { target_id: targetId });
  return res.data;
}

export const listUsers = async () => {
  try {
    const res = await api.get('/users');
    console.log('listUsers API response:', res);
    if (!res.data) {
      console.error('listUsers: No data in response');
      return { users: [] };
    }
    return res.data;
  } catch (error) {
    console.error('listUsers API error:', error);
    throw error;
  }
}
