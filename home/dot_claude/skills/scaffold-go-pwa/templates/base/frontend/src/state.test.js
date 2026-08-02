import { describe, it, expect, beforeEach } from 'vitest'
import { state, setError, setNotice, clearBanner } from './state.js'

beforeEach(() => {
  clearBanner()
})

describe('banner', () => {
  it('starts hidden', () => {
    expect(state.banner.text).toBe('')
  })

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
