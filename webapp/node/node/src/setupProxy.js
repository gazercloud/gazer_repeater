const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  app.use(
    '/api/request',
    createProxyMiddleware({
      target: 'https://123-n.gazer.cloud',
        changeOrigin: true,
		secure: false,
        cookieDomainRewrite: "gazer.cloud",
      onProxyReq: (proxyReq) => {
        if (proxyReq.getHeader('origin')) {
          proxyReq.setHeader('origin', 'https://123-n.gazer.cloud')
        }
      }
    })
  );
};