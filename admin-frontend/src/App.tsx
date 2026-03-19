import { ConfigProvider, App as AntdApp } from 'antd';
import AppRouter from './router';
import { theme } from './theme';

export default function App() {
  return (
    <ConfigProvider theme={theme}>
      <AntdApp>
        <AppRouter />
      </AntdApp>
    </ConfigProvider>
  );
}

