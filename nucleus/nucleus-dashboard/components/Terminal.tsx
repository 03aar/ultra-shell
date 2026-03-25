'use client';

import { useEffect, useRef } from 'react';
import type { Terminal as XTerminal } from 'xterm';

interface TerminalProps {
  wsUrl?: string;
}

export default function Terminal({ wsUrl }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<XTerminal | null>(null);

  useEffect(() => {
    let disposed = false;

    const initTerminal = async () => {
      const { Terminal: XTerm } = await import('xterm');
      const { FitAddon } = await import('xterm-addon-fit');
      const { WebLinksAddon } = await import('xterm-addon-web-links');

      if (disposed) return;

      const terminal = new XTerm({
        theme: {
          background: '#0a0a0a',
          foreground: '#e8e8e8',
          cursor: '#00ff88',
          cursorAccent: '#000000',
          selectionBackground: '#00ff8830',
          black: '#000000',
          red: '#ef4444',
          green: '#00ff88',
          yellow: '#fbbf24',
          blue: '#60a5fa',
          magenta: '#a78bfa',
          cyan: '#38bdf8',
          white: '#e8e8e8',
          brightBlack: '#666666',
          brightRed: '#f87171',
          brightGreen: '#4ade80',
          brightYellow: '#fde047',
          brightBlue: '#93c5fd',
          brightMagenta: '#c084fc',
          brightCyan: '#67e8f9',
          brightWhite: '#ffffff',
        },
        fontFamily: 'JetBrains Mono, monospace',
        fontSize: 13,
        lineHeight: 1.4,
        cursorBlink: true,
        cursorStyle: 'bar',
      });

      const fitAddon = new FitAddon();
      terminal.loadAddon(fitAddon);
      terminal.loadAddon(new WebLinksAddon());

      if (terminalRef.current) {
        terminal.open(terminalRef.current);
        fitAddon.fit();
      }

      xtermRef.current = terminal;

      // Connect to WebSocket
      const wsEndpoint =
        wsUrl || `ws://localhost:8080/api/v1/ws/stream?api_key=dev-nucleus-key-local`;
      const ws = new WebSocket(wsEndpoint);

      ws.onopen = () => {
        terminal.writeln('\x1b[1;32m[NUCLEUS]\x1b[0m Connected to shell runtime');
        terminal.writeln('\x1b[1;32m[NUCLEUS]\x1b[0m Watching command executions...');
        terminal.writeln('');
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === 'execution_complete') {
            const exec = msg.data;
            const exitColor = exec.exit_code === 0 ? '32' : '31';
            terminal.writeln(
              `\x1b[90m${new Date(exec.timestamp).toLocaleTimeString()}\x1b[0m ` +
                `\x1b[1;${exitColor}m$\x1b[0m ${exec.command.raw}`
            );

            if (exec.risk_flags && exec.risk_flags.length > 0) {
              for (const flag of exec.risk_flags) {
                const riskColor =
                  flag.level === 'critical' ? '31' : flag.level === 'high' ? '33' : '33';
                terminal.writeln(
                  `  \x1b[${riskColor}m! [${flag.level.toUpperCase()}]\x1b[0m ${flag.message}`
                );
              }
            }

            if (exec.stdout && exec.stdout.trim()) {
              const lines = exec.stdout.trim().split('\n').slice(0, 10);
              for (const line of lines) {
                terminal.writeln(`  ${line}`);
              }
              if (exec.stdout.trim().split('\n').length > 10) {
                terminal.writeln(`  \x1b[90m... (truncated)\x1b[0m`);
              }
            }

            if (exec.rollback_available) {
              terminal.writeln(
                `  \x1b[32m<- Rollback available (${exec.id.slice(0, 8)})\x1b[0m`
              );
            }

            terminal.writeln('');
          } else if (msg.type === 'rollback_complete') {
            terminal.writeln(
              `\x1b[1;33m[ROLLBACK]\x1b[0m Execution rolled back successfully`
            );
            terminal.writeln('');
          }
        } catch {
          // ignore parse errors
        }
      };

      ws.onclose = () => {
        terminal.writeln('\x1b[1;31m[NUCLEUS]\x1b[0m Disconnected from shell runtime');
      };

      // Handle resize
      const resizeObserver = new ResizeObserver(() => {
        try {
          fitAddon.fit();
        } catch {
          // ignore fit errors during disposal
        }
      });
      if (terminalRef.current) {
        resizeObserver.observe(terminalRef.current);
      }

      return () => {
        resizeObserver.disconnect();
        ws.close();
        terminal.dispose();
      };
    };

    let cleanup: (() => void) | undefined;
    initTerminal().then((fn) => {
      cleanup = fn;
    });

    return () => {
      disposed = true;
      if (cleanup) cleanup();
      if (xtermRef.current) {
        xtermRef.current.dispose();
        xtermRef.current = null;
      }
    };
  }, [wsUrl]);

  return (
    <div className="h-full">
      <link
        rel="stylesheet"
        href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css"
      />
      <div ref={terminalRef} className="h-full w-full" />
    </div>
  );
}
