const http = require('http');
const port = process.env.PORT || 3000;

http.createServer((req, res) => {
  res.writeHead(200, {'Content-Type': 'text/plain'});
  res.end(`Hello from Node on port ${port}!\nHost: ${req.headers.host}\n`);
}).listen(port, () => {
  console.log(`Server running on port ${port}`);
});
