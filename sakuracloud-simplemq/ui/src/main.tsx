import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { SWRConfig } from 'swr'
import { App } from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>

    <SWRConfig
      value={{
        refreshInterval: 1000,
        fetcher: (resource, init) => fetch(resource, init).then(res => res.json())
      }}
    >
      <App />
    </SWRConfig>
  </StrictMode>,
)
