import { describe, it, expect, vi, afterEach } from 'vitest'
import { req } from './api.js'

function mockFetch(status, body, ok) {
  global.fetch = vi.fn().mockResolvedValue({
    ok,
    status,
    statusText: 'Status',
    text: () => Promise.resolve(body),
  })
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('req', () => {
  it('parses a JSON body', async () => {
    mockFetch(200, '{"version":"dev"}', true)
    await expect(req('GET', '/version')).resolves.toEqual({ version: 'dev' })
  })

  it('returns null for 204 without reading a body', async () => {
    mockFetch(204, '', true)
    await expect(req('POST', '/auth/logout')).resolves.toBeNull()
  })

  it('raises the server error message, not the status line', async () => {
    mockFetch(400, '{"error":"bad input"}', false)
    await expect(req('POST', '/things', {})).rejects.toThrow('bad input')
  })

  it('falls back to the status line when the error carries no message', async () => {
    mockFetch(500, '', false)
    await expect(req('GET', '/things')).rejects.toThrow('500 Status')
  })

  it('sends credentials so the session cookie rides along', async () => {
    mockFetch(200, '{}', true)
    await req('GET', '/version')
    expect(global.fetch).toHaveBeenCalledWith('/api/v1/version', expect.objectContaining({ credentials: 'include' }))
  })
})
