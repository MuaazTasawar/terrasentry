import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Produces a self-contained `.next/standalone` build with only the
  // production dependencies it actually needs — keeps the Docker image
  // lean instead of shipping the full node_modules tree.
  output: "standalone",
};

export default nextConfig;