---
applyTo: "**/*.tsx"
---

# TypeScript / React Best Practices — DockerDesktopMCP UI

## Environment

- **TypeScript**: ^6.0.2 (strict mode enabled, `noEmit: true`)
- **React**: ^19.2.5 with `react-dom` ^19.2.5
- **Build tool**: Vite ^8.0.8 with `@vitejs/plugin-react` ^6.0.1
- **UI library**: MUI (Material UI) ^9.0.0 with `@emotion/react` ^11.14.0 and `@emotion/styled` ^11.14.1
- **Target**: ES2020, module bundler resolution, `jsx: react-jsx`
- **Entry point**: `src/ui/src/index.tsx`; app root: `src/ui/src/App.tsx`

## TypeScript

- Enable and respect the `strict: true` compiler option — no `@ts-ignore` or `any` casts without a documented reason.
- Declare explicit return types on all exported functions and React components.
- Prefer `interface` for object shapes, `type` for unions/aliases and utility types.
- Use `const` by default; use `let` only when reassignment is required; never use `var`.
- Acronyms in identifiers follow PascalCase: `MCP`, `URL`, `HTTP`, `ID`.
- Avoid non-null assertion (`!`) except where the DOM guarantees non-null (e.g. `getElementById('root')!`).

## React

- Use **function components** exclusively — no class components except the existing `ErrorBoundary` in `index.tsx`.
- Name every component with PascalCase and export it as a named export (e.g. `export function App()`).
- Destructure props in the function signature and annotate the prop type inline or with a named `interface`.
- Mark prop interfaces for read-only components with `Readonly<>` (e.g. `Readonly<{ status: ServerStatus }>`).
- Use React hooks (`useState`, `useEffect`, `useCallback`, `useMemo`) for all state and side-effect management.
- Declare stable callbacks with `useCallback` and memoised values with `useMemo` to avoid unnecessary re-renders.
- Always supply `key` props when rendering lists; use a stable, unique value — not the array index.

## MUI (Material UI v9)

- Import components from `@mui/material` and icons from `@mui/icons-material` using named imports — never default imports.
- Use `Stack` for one-dimensional layouts and `Grid` (v2 API, `size` prop) for two-dimensional grids.
- Apply inline styles via the `sx` prop; do not use plain `style={{}}` attributes for MUI components.
- Use MUI theme colours (`color="text.secondary"`, `bgcolor="action.hover"`, etc.) instead of hard-coded hex values.
- Wrap icon buttons in `<Tooltip>` and always provide a `title` for accessibility.
- Use `variant="outlined"` for `Paper` cards and `size="small"` for secondary actions.

## File & Module Organisation

- One component per file; the file name matches the component name in PascalCase (e.g. `App.tsx`).
- Utility/helper modules use camelCase (e.g. `ddClient.ts`).
- Group imports: React → third-party libraries → local modules, each separated by a blank line.
- Do not import from `node_modules` paths other than the declared dependencies in `package.json`.

## State & Async

- Model server/fetch states with discriminated union types (e.g. `type ServerStatus = 'checking' | 'running' | 'offline'`).
- Wrap all `fetch` / API calls in `try/catch` and return `undefined` on failure rather than throwing.
- Use the nullish coalescing assignment operator (`??=`) to fall back between data sources.
- Cancel or guard async operations tied to component lifecycle with `useCallback` deps or `AbortController`.

## Vite & Build

- Run `npm run build` from `src/ui/` to compile; the output goes to `src/ui/dist/`.
- **The build must produce zero warnings.** Resolve all TypeScript and Vite warnings before committing.
- The `base: './'` option in `vite.config.ts` is required for Docker Desktop Extensions — do not change it.
- Do not import assets with absolute paths; use relative paths so the `base` rewrite works correctly.

## Styling

- Use MUI's `sx` prop for component-level styling — avoid separate CSS files or inline `style` objects on MUI components.
- Use `fontFamily: 'monospace'` for code/endpoint text; use the theme's typography scale for all other text.
- Respect dark/light mode — use theme-aware colour tokens; do not hard-code colours for foreground or background.

## Dependencies

- Declared runtime dependencies (do not add without justification):
  - `@emotion/react ^11.14.0`
  - `@emotion/styled ^11.14.1`
  - `@mui/icons-material ^9.0.0`
  - `@mui/material ^9.0.0`
  - `react ^19.2.5`
  - `react-dom ^19.2.5`
- Declared dev dependencies:
  - `@types/react ^19.2.14`
  - `@types/react-dom ^19.2.3`
  - `@vitejs/plugin-react ^6.0.1`
  - `typescript ^6.0.2`
  - `vite ^8.0.8`
- **Never commit `node_modules/`** — only list packages in `package.json`.

## Accessibility

- Every interactive element must have a human-readable label or `title` prop.
- Use `<span>` wrappers around disabled MUI components so that `Tooltip` can attach (as shown in `App.tsx`).
- Prefer semantic HTML elements (`<strong>`, `<code>`, `<pre>`) where appropriate.

## Security

- Do not log or expose secrets or credentials in the UI.
- Sanitise any user-supplied content before rendering — avoid `dangerouslySetInnerHTML`.
- Limit `fetch` calls to the known local endpoint (`127.0.0.1:3282`); do not forward arbitrary URLs.
