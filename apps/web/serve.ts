const server = Bun.serve({
  port: 3000,
  async fetch(req) {
    const url = new URL(req.url);
    
    // Serve install script
    if (url.pathname === "/install.sh") {
      return new Response(await Bun.file("./install.sh").text(), {
        headers: { "Content-Type": "text/plain" },
      });
    }

    // Serve static HTML
    return new Response(Bun.file("./index.html"));
  },
});

console.log(`🔥 Crackerbox web → http://localhost:${server.port}`);
