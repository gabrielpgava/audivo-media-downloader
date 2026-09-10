import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('unified downloader screen', () => {
  it('renders the task-first idle experience without provider navigation', () => {
    const markup = renderToStaticMarkup(<App />)

    expect(markup).toContain('URL do YouTube ou Apple Music')
    expect(markup).toContain('Analisar')
    expect(markup).toContain('Abrir configurações')
    expect(markup).not.toContain('YouTube Downloader')
    expect(markup).not.toContain('Apple Music Downloader')
  })
})
