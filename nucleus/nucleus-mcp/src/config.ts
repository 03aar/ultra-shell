export interface NucleusMCPConfig {
  apiUrl: string;
  apiKey: string;
  logLevel: "debug" | "info" | "warn" | "error";
}

export function loadConfig(): NucleusMCPConfig {
  const logLevel = (process.env.NUCLEUS_MCP_LOG_LEVEL || "info") as NucleusMCPConfig["logLevel"];
  const validLevels = ["debug", "info", "warn", "error"];
  if (!validLevels.includes(logLevel)) {
    throw new Error(
      `Invalid NUCLEUS_MCP_LOG_LEVEL: "${logLevel}". Must be one of: ${validLevels.join(", ")}`
    );
  }

  return {
    apiUrl: process.env.NUCLEUS_API_URL || "http://localhost:8080",
    apiKey: process.env.NUCLEUS_API_KEY || "dev-nucleus-key-local",
    logLevel,
  };
}

const LOG_LEVELS: Record<string, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

export class Logger {
  private level: number;

  constructor(logLevel: string) {
    this.level = LOG_LEVELS[logLevel] ?? 1;
  }

  debug(message: string, ...args: unknown[]): void {
    if (this.level <= 0) {
      process.stderr.write(`[DEBUG] ${message} ${args.map((a) => JSON.stringify(a)).join(" ")}\n`);
    }
  }

  info(message: string, ...args: unknown[]): void {
    if (this.level <= 1) {
      process.stderr.write(`[INFO] ${message} ${args.map((a) => JSON.stringify(a)).join(" ")}\n`);
    }
  }

  warn(message: string, ...args: unknown[]): void {
    if (this.level <= 2) {
      process.stderr.write(`[WARN] ${message} ${args.map((a) => JSON.stringify(a)).join(" ")}\n`);
    }
  }

  error(message: string, ...args: unknown[]): void {
    if (this.level <= 3) {
      process.stderr.write(`[ERROR] ${message} ${args.map((a) => JSON.stringify(a)).join(" ")}\n`);
    }
  }
}
