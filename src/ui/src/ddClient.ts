// Local shim for @docker/extension-api-client.
// When running inside Docker Desktop, window.ddClient is injected by the host.
// This shim exposes the same subset of the API used by App.tsx.

interface DDService {
  get(path: string): Promise<unknown>;
  post?(path: string, body?: unknown): Promise<unknown>;
}

interface DDVm {
  service?: DDService;
}

interface DDExtension {
  vm?: DDVm;
}

// Result returned by docker.cli.exec — matches the Docker Desktop Extension SDK.
export interface ExecResult {
  stdout: string;
  stderr: string;
}

interface DockerCli {
  exec(cmd: string, args: string[]): Promise<ExecResult>;
}

interface DDDocker {
  cli: DockerCli;
}

interface DDClient {
  extension: DDExtension;
  /** Available only inside Docker Desktop; undefined in development/fallback. */
  docker?: DDDocker;
}

declare global {
  interface Window {
    ddClient?: DDClient;
  }
}


export function createDockerDesktopClient(): DDClient {
  const hostWindow = globalThis as typeof globalThis & { ddClient?: DDClient };
  if (hostWindow.ddClient) {
    return hostWindow.ddClient;
  }
  // Fallback stub for development / outside Docker Desktop
  return {
    extension: {
      vm: {
        service: {
          get: (path: string) =>
            fetch(`http://127.0.0.1:3282${path}`).then((r) => r.json()),
          post: (path: string, body?: unknown) =>
            fetch(`http://127.0.0.1:3282${path}`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: body ? JSON.stringify(body) : undefined,
            }).then((r) => r.json()),
        },
      },
    },
  };
}
