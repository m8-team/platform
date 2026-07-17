import {createRoot} from 'react-dom/client'
import {QueryClient, QueryClientProvider} from '@tanstack/react-query'
import {RouterProvider} from '@tanstack/react-router'

import '../apps/console/src/reset.css'
import '@gravity-ui/uikit/styles/fonts.css'
import '@gravity-ui/uikit/styles/styles.css'
import '../apps/console/src/modules/commerce-intelligence/styles/commerce-intelligence.css'

import {router} from '../apps/console/src/router.tsx'

const queryClient = new QueryClient()

createRoot(document.getElementById('root')!).render(
  <QueryClientProvider client={queryClient}>
    <RouterProvider router={router} />
  </QueryClientProvider>,
)
