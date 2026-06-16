import {StrictMode} from 'react'
import {createRoot} from 'react-dom/client'
import {configure, ThemeProvider} from '@gravity-ui/uikit';

import '@gravity-ui/uikit/styles/fonts.css'
import '@gravity-ui/uikit/styles/styles.css'

import App from './App.tsx'

configure({
    lang: 'ru',
});

createRoot(document.getElementById('root')!).render(
  <StrictMode>
      <ThemeProvider theme="light">
          <App />
      </ThemeProvider>
  </StrictMode>,
)
