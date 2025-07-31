
export default {
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': 'http://server:8080',
      '/login': 'http://auth:8081',
      '/register': 'http://auth:8081',
      '/chatlist': 'http://server:8080',
      '/chat': {
        target: 'ws://server:8080',
        ws: true,
      }
    }
  }
};