# 2. Derive the app icons from a single emoji

Date: __DATE__

## Status

Accepted

## Context

A PWA needs a favicon, an apple-touch icon, and 192/512 plus maskable install
icons. Commissioning artwork for an internal tool is out of proportion, but
shipping no icon costs installability: the browser falls back to a screenshot on
the home screen, and the maskable variant is simply absent.

## Decision

One emoji is the app identity: __EMOJI__.

- The browser tab uses an inline SVG `data:` URI in `index.html` that draws the
  emoji as text. No file, always in sync.
- The install icons are rasterised once from that same emoji into
  `frontend/public/icons/` and committed, because a manifest cannot point at an
  SVG for maskable purposes and iOS ignores SVG for the home-screen icon.

## Consequences

- Changing the identity is changing one character plus a re-run of the generator;
  nothing else references the artwork.
- The PNGs are committed binaries. They are small (a few KB each) and change only
  when the emoji does.
- The rasteriser is a macOS-only AppKit script, since Apple Color Emoji is the
  only emoji font on a stock developer machine here. CI never runs it: the icons
  are build inputs, not build outputs.
- The tab favicon renders with the *viewer's* emoji font, so it differs slightly
  between platforms. The install icons do not: they are pixels.
