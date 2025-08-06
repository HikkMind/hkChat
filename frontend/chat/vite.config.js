
export default {
  server: {
    host: '0.0.0.0',
    proxy: {
      // '/api': 'http://server:8080',
      '/login': 'http://auth:8081',
      '/register': 'http://auth:8081',
      // '/chatlist': 'http://connection:8080',
      '/chatlist': {
        target: 'ws://connection:8080',
        ws: true,
      }
    }
  }
};