from http.server import HTTPServer, BaseHTTPRequestHandler
import os

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/html')
        self.end_headers()
        port = os.environ.get('PORT', '8000')
        html = f'''
        <h1>Python HTTP Server</h1>
        <p>Port: {port}</p>
        <p>Host: {self.headers.get("Host")}</p>
        <p>Path: {self.path}</p>
        '''
        self.wfile.write(html.encode())

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8000))
    server = HTTPServer(('', port), Handler)
    print(f'Python HTTP server listening on port {port}')
    server.serve_forever()
