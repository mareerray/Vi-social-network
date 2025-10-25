import api from './index';

export const createPost = async (postData) => {
  const res = await api.post('/posts/create', postData);
  return res.data;
}

export const listPosts = async (user_id) => {
  const url = user_id ? `/posts?user_id=${user_id}` : '/posts';
  const res = await api.get(url);
  return res.data;
}

export const addComment = async (post_id, content, image_url) => {
  const res = await api.post('/posts/comment', { post_id, content, image_url });
  return res.data;
}
