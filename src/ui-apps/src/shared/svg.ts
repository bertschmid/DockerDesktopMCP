export interface PieSlice {
  label: string;
  value: number;
  color: string;
}

function polarToCartesian(cx: number, cy: number, r: number, angle: number): [number, number] {
  const rad = ((angle - 90) * Math.PI) / 180;
  return [cx + r * Math.cos(rad), cy + r * Math.sin(rad)];
}

export function donutPath(cx: number, cy: number, outerR: number, innerR: number, start: number, end: number): string {
  const [sx, sy] = polarToCartesian(cx, cy, outerR, end);
  const [ex, ey] = polarToCartesian(cx, cy, outerR, start);
  const [isx, isy] = polarToCartesian(cx, cy, innerR, end);
  const [iex, iey] = polarToCartesian(cx, cy, innerR, start);
  const largeArc = end - start <= 180 ? 0 : 1;

  return [
    `M ${sx} ${sy}`,
    `A ${outerR} ${outerR} 0 ${largeArc} 0 ${ex} ${ey}`,
    `L ${iex} ${iey}`,
    `A ${innerR} ${innerR} 0 ${largeArc} 1 ${isx} ${isy}`,
    "Z"
  ].join(" ");
}

export function renderDonut(slices: PieSlice[], size = 240): string {
  const total = slices.reduce((acc, s) => acc + s.value, 0);
  if (total <= 0) {
    return "<p class=\"muted\">No data available.</p>";
  }

  const cx = size / 2;
  const cy = size / 2;
  const outerR = size * 0.42;
  const innerR = size * 0.23;

  let angle = 0;
  const paths = slices
    .map((slice) => {
      const portion = (slice.value / total) * 360;
      const start = angle;
      const end = angle + portion;
      angle = end;
      const d = donutPath(cx, cy, outerR, innerR, start, end);
      return `<path d="${d}" fill="${slice.color}"></path>`;
    })
    .join("");

  return `<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}">${paths}</svg>`;
}
