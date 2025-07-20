
export default {
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': 'http://server:8080',
      '/login': 'http://server:8080',
      '/register': 'http://server:8080',
      '/messager': {
        target: 'ws://server:8080',
        ws: true,
      }
    }
  }
};