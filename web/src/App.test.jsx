// Минимальный тест без роутера
import { useState } from 'react'

export default function TestApp() {
  const [count, setCount] = useState(0)
  
  return (
    <div style={{ padding: 40, background: '#0f172a', minHeight: '100vh', color: 'white' }}>
      <h1>React Works! 🎉</h1>
      <p>If you see this, React is rendering correctly.</p>
      <button 
        onClick={() => setCount(c => c + 1)}
        style={{ padding: '10px 20px', fontSize: 16, cursor: 'pointer' }}
      >
        Count: {count}
      </button>
    </div>
  )
}
