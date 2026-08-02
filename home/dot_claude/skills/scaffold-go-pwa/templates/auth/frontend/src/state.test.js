import { describe, it, expect, beforeEach } from 'vitest'
import { state, setError, setNotice, clearBanner, initials } from './state.js'

beforeEach(() => {
  clearBanner()
})

describe('banner', () => {
  it('distinguishes an error from a notice so the styling can differ', () => {
    setError('boom')
    expect(state.banner).toEqual({ kind: 'error', text: 'boom' })
    setNotice('saved')
    expect(state.banner).toEqual({ kind: 'notice', text: 'saved' })
  })

  it('clears', () => {
    setError('boom')
    clearBanner()
    expect(state.banner).toEqual({ kind: '', text: '' })
  })
})

describe('initials', () => {
  it('takes at most two initials', () => {
    expect(initials('Ada Lovelace')).toBe('AL')
    expect(initials('Ada Byron King Lovelace')).toBe('AB')
  })

  it('handles a single name and extra spacing', () => {
    expect(initials('  ada  ')).toBe('A')
  })

  it('falls back when there is no name', () => {
    expect(initials('')).toBe('?')
    expect(initials(null)).toBe('?')
  })
})
