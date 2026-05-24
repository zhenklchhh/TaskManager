import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
// import './index.css'  // Disabled for testing

function TestApp() {
  return (
    <div style={{ padding: 40, background: '#0f172a', minHeight: '100vh', color: 'white', fontFamily: 'sans-serif' }}>
      <h1>React is working!</h1>
      <p>If you see this, the problem is in index.css (Tailwind)</p>
    </div>
  )
}

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <TestApp />
  </StrictMode>,
)
