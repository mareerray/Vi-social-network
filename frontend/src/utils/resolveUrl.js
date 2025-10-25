export function resolveAsset(path) {
  if (!path) return ''
  // if it's a short token like 'avatar' or 'none', treat as empty
  if (typeof path === 'string' && !path.includes('/') && path.length <= 10) return ''
  // already absolute
  if (/^https?:\/\//i.test(path)) return path
  // ensure leading slash
  const normalized = path.startsWith('/') ? path : `/${path}`
  // point directly to backend in dev
  return normalized;
}
