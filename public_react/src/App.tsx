import { useState } from 'react'
import RadioPlayer from './RadioPlayer'

function App() {
  const [clicked, setClicked] = useState(false)

  const handleClick = () => { setClicked(!clicked) }

  return (
    <>
      <div className='min-h-screen bg-black w-100 min-w-screen content-center p-20'>
        {clicked ? <RadioPlayer /> : <button onClick={handleClick}>Enter</button>}
      </div>
    </>
  )
}

export default App
