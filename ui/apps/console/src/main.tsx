import {createRoot} from 'react-dom/client'
import {RouterProvider} from '@tanstack/react-router'

import './reset.css'
import '@gravity-ui/uikit/styles/fonts.css'
import '@gravity-ui/uikit/styles/styles.css'

import {router} from './router.tsx'

createRoot(document.getElementById('root')!).render(<RouterProvider router={router} />)
