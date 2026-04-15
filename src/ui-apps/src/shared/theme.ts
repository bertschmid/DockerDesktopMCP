import {
  applyDocumentTheme,
  applyHostFonts,
  applyHostStyleVariables,
  App,
  PostMessageTransport
} from "@modelcontextprotocol/ext-apps";

export function createAndConnectApp(name: string): App {
  const app = new App({ name, version: "1.0.0" });

  app.onhostcontextchanged = (ctx) => {
    if (ctx.theme) {
      applyDocumentTheme(ctx.theme);
    }
    if (ctx.styles?.variables) {
      applyHostStyleVariables(ctx.styles.variables);
    }
    if (ctx.styles?.css?.fonts) {
      applyHostFonts(ctx.styles.css.fonts);
    }
    if (ctx.safeAreaInsets) {
      const { top, right, bottom, left } = ctx.safeAreaInsets;
      document.body.style.padding = `${top}px ${right}px ${bottom}px ${left}px`;
    }
  };

  app.onteardown = async () => ({});

  return app;
}

export async function connectApp(app: App): Promise<void> {
  await app.connect(new PostMessageTransport());
}

export function applyBaseStyles(): void {
  const css = `
    :root { color-scheme: light dark; }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: var(--font-sans, ui-sans-serif, system-ui);
      background: var(--color-background-primary, #111827);
      color: var(--color-text-primary, #f9fafb);
    }
    #root { padding: 12px; }
    h1, h2, h3 { margin: 0 0 10px; }
    .muted { color: var(--color-text-secondary, #9ca3af); }
    .chip {
      border-radius: 9999px;
      padding: 2px 10px;
      font-size: 12px;
      display: inline-block;
      border: 1px solid var(--color-border-default, #374151);
      margin-right: 6px;
      margin-bottom: 6px;
    }
    .chip-running { background: rgba(16, 185, 129, 0.15); }
    .chip-paused { background: rgba(245, 158, 11, 0.15); }
    .chip-stopped, .chip-exited { background: rgba(107, 114, 128, 0.2); }
    .card {
      border: 1px solid var(--color-border-default, #374151);
      border-radius: 10px;
      padding: 10px;
      margin-bottom: 12px;
      background: var(--color-background-secondary, #1f2937);
    }
  `;

  const style = document.createElement("style");
  style.textContent = css;
  document.head.appendChild(style);
}
