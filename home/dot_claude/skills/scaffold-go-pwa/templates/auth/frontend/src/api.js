// Thin wrapper over the __TITLE__ REST API. Same-origin; the session rides an
// httpOnly cookie, so every call includes credentials.
const base = '/api/v1'

export async function req(method, path, body) {
  const opts = { method, credentials: 'include', headers: {} }
  if (body !== undefined) {
    opts.headers['Content-Type'] = 'application/json'
    opts.body = JSON.stringify(body)
  }
  const res = await fetch(base + path, opts)
  if (res.status === 204) return null
  const text = await res.text()
  const data = text ? JSON.parse(text) : null
  if (!res.ok) {
    throw new Error((data && data.error) || `${res.status} ${res.statusText}`)
  }
  return data
}

export const api = {
  version: () => req('GET', '/version'),
  instance: () => req('GET', '/instance'),
  register: (b) => req('POST', '/auth/register', b),
  login: (b) => req('POST', '/auth/login', b),
  logout: () => req('POST', '/auth/logout'),
  me: () => req('GET', '/me'),
}
