import { escapeHtml } from "./formatting";

export interface Column<T> {
  header: string;
  render: (row: T) => string;
}

export function renderTable<T>(target: HTMLElement, rows: T[], columns: Column<T>[]): void {
  const headers = columns.map((c) => `<th>${escapeHtml(c.header)}</th>`).join("");
  const body = rows
    .map((row) => {
      const cells = columns
        .map((c) => `<td>${c.render(row)}</td>`)
        .join("");
      return `<tr>${cells}</tr>`;
    })
    .join("");

  target.innerHTML = `
    <table>
      <thead><tr>${headers}</tr></thead>
      <tbody>${body}</tbody>
    </table>
  `;
}

export function applyTableStyles(): void {
  const style = document.createElement("style");
  style.textContent = `
    table {
      width: 100%;
      border-collapse: collapse;
      border: 1px solid var(--color-border-default, #374151);
      border-radius: 8px;
      overflow: hidden;
      font-size: 12px;
    }
    th, td {
      border-bottom: 1px solid var(--color-border-default, #374151);
      padding: 8px;
      text-align: left;
      vertical-align: top;
    }
    th {
      position: sticky;
      top: 0;
      background: var(--color-background-secondary, #1f2937);
      font-weight: 600;
    }
    tr:last-child td {
      border-bottom: 0;
    }
    code {
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 11px;
      background: rgba(127, 127, 127, 0.15);
      padding: 1px 4px;
      border-radius: 4px;
    }
    .actions button {
      margin-right: 6px;
      margin-bottom: 4px;
    }
  `;
  document.head.appendChild(style);
}
