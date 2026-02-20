const port = parseInt(Deno.env.get("PORT") || "8000");

const handler = (req: Request): Response => {
  const url = new URL(req.url);
  const html = `
    <h1>Deno HTTP Server</h1>
    <p>Port: ${port}</p>
    <p>Host: ${req.headers.get("host")}</p>
    <p>Path: ${url.pathname}</p>
  `;
  return new Response(html, {
    headers: { "content-type": "text/html" },
  });
};

console.log(`Deno HTTP server listening on port ${port}`);
Deno.serve({ port }, handler);
