'use client';

import './globals.css';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const navItems = [
  { href: '/', label: 'Dashboard', icon: '⬡' },
  { href: '/graph', label: 'Execution Graph', icon: '◈' },
  { href: '/sessions', label: 'Sessions', icon: '◉' },
  { href: '/context', label: 'Environment', icon: '◎' },
  { href: '/agent', label: 'Agent', icon: '◆' },
  { href: '/api-explorer', label: 'API Explorer', icon: '⬢' },
];

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  return (
    <html lang="en">
      <body className="min-h-screen bg-nucleus-bg text-nucleus-text">
        <div className="flex h-screen">
          {/* Sidebar */}
          <nav className="w-56 bg-nucleus-surface border-r border-nucleus-border flex flex-col">
            <div className="p-4 border-b border-nucleus-border">
              <h1 className="font-heading text-xl font-bold text-nucleus-accent tracking-wider">
                NUCLEUS
              </h1>
              <p className="text-xs text-nucleus-muted mt-1 font-body">AI-Native Shell Runtime</p>
            </div>
            <div className="flex-1 py-2">
              {navItems.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                    pathname === item.href
                      ? 'bg-nucleus-panel text-nucleus-accent border-r-2 border-nucleus-accent'
                      : 'text-nucleus-muted hover:text-nucleus-text hover:bg-nucleus-panel/50'
                  }`}
                >
                  <span className="text-base">{item.icon}</span>
                  <span className="font-body">{item.label}</span>
                </Link>
              ))}
            </div>
            <div className="p-4 border-t border-nucleus-border">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-nucleus-accent animate-pulse" />
                <span className="text-xs text-nucleus-muted font-mono">v1.0.0</span>
              </div>
            </div>
          </nav>

          {/* Main Content */}
          <main className="flex-1 overflow-auto">{children}</main>
        </div>
      </body>
    </html>
  );
}
