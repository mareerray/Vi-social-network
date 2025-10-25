// api/auth.js
import axios from 'axios';
import api from '.';

export const register = (userData) => {
  return api.post('/register', userData);
};

// login accepts either email or nickname as the identifier
export const login = async (identifier, password) => {
  const response = await api.post('/login', { identifier, password });
  return response.data; // { user_id: "..." }
};

export const logout = () => {
  return api.post('/logout');
};

export const getMe = async () => {
  // backend exposes /api/check-session which validates cookie and returns user info
  const response = await api.get('/check-session');
  return response.data; // expect { user_id: "..." }
};