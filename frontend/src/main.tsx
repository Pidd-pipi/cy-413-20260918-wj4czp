import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import { AppRouter } from './router';
import { GlobalErrorBoundary } from './components/common/GlobalErrorBoundary';
import { themeStore } from './stores/themeStore';
import { THEMES } from './constants/themes';
import './styles/app.css';

document.documentElement.dataset.theme = themeStore.get();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <GlobalErrorBoundary>
      <ConfigProvider theme={{ token: { colorPrimary: THEMES[themeStore.get()].color, borderRadius: 14 } }}>
        <BrowserRouter>
          <AppRouter />
        </BrowserRouter>
      </ConfigProvider>
    </GlobalErrorBoundary>
  </React.StrictMode>,
);
