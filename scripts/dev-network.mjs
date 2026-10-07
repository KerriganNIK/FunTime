import { networkInterfaces } from 'node:os';

export function getLanOrigins(port = 5173) {
  return [...new Set(Object.entries(networkInterfaces())
    .filter(([name]) => !/loopback|tun|vpn|wsl|vethernet|virtual|tailscale|docker/i.test(name))
    .flatMap(([, interfaces]) => (interfaces ?? []).filter(item => item.family === 'IPv4' && !item.internal).map(item => item.address))
    .filter(address => /^(192\.168\.|10\.|172\.(1[6-9]|2\d|3[01])\.)/.test(address)))]
    .sort((a, b) => Number(!a.startsWith('192.168.')) - Number(!b.startsWith('192.168.')) || a.localeCompare(b))
    .map(address => `http://${address}:${port}`);
}
