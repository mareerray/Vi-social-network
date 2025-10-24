// api/auth.js
import axios from 'axios';

// It's good practice to create an axios instance
const apiClient = axios.create({
  baseURL: 'http://localhost:8080', // backend runs on :8080
  withCredentials: true, // send cookies with requests
});

export const register = (userData) => {
  return apiClient.post('/register', userData);
};

// login accepts either email or nickname as the identifier
export const login = async (identifier, password) => {
  const response = await apiClient.post('/login', { identifier, password });
  return response.data; // { user_id: "..." }
};

export const logout = () => {
  return apiClient.post('/logout');
};

export const getMe = async () => {
  // backend exposes /api/check-session which validates cookie and returns user info
  const response = await apiClient.get('/api/check-session');
  return response.data; // expect { user_id: "..." }
};