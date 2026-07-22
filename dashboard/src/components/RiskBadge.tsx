const RISK_STYLES: Record<string, string> = {
  low: "bg-emerald-950 text-emerald-400 ring-emerald-800",
  medium: "bg-amber-950 text-amber-400 ring-amber-800",
  high: "bg-rose-950 text-rose-400 ring-rose-800",
};

export default function RiskBadge({ level }: { level: string }) {
  const style = RISK_STYLES[level] ?? "bg-zinc-800 text-zinc-400 ring-zinc-700";

  return (
    <span
      className={`inline-flex items-center rounded-sm px-2 py-0.5 text-xs font-mono font-medium uppercase tracking-wide ring-1 ring-inset ${style}`}
    >
      {level}
    </span>
  );
}