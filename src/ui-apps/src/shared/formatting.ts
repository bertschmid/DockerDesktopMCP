export function escapeHtml(input: string): string {
  return input
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

export function formatDate(input: string): string {
  const d = new Date(input);
  if (Number.isNaN(d.getTime())) {
    return input;
  }
  return d.toLocaleString();
}

export function parseHumanBytes(value: string): number {
  const byteExpr = /^(\d+(?:\.\d+)?)\s*([KMGTPE]?I?B)$/i;
  const m = byteExpr.exec(value.trim());
  if (!m) {
    return 0;
  }
  const num = Number(m[1]);
  const unit = m[2].toUpperCase().replace("IB", "B");
  const powerMap: Record<string, number> = {
    B: 0,
    KB: 1,
    MB: 2,
    GB: 3,
    TB: 4,
    PB: 5,
    EB: 6
  };
  const pow = powerMap[unit] ?? 0;
  return num * Math.pow(1024, pow);
}

export function prettyNumber(n: number): string {
  return new Intl.NumberFormat().format(n);
}
