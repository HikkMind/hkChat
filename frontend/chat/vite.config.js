
export default {
  server: {
    host: '0.0.0.0',
    hmr: {
      host: 'localhost',
      // port: 80,
      protocol: 'ws',
    },
    proxy: {
      '/login': 'http://auth:8081',
      '/register': 'http://auth:8081',
      '/verifytoken': 'http://auth:8081',
      // '/chatlist': {
      //   target: 'ws://connection:8080',
      //   ws: true,
      // }
    }
  }
};