import type { Scan } from "@/lib/apiClient";
import RiskBadge from "./RiskBadge";

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    });
  } catch {
    return iso;
  }
}

export default function ScanTable({ scans }: { scans: Scan[] }) {
  if (scans.length === 0) {
    return (
      <p className="py-12 text-center text-sm text-zinc-500">
        No scans match the current filters.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto rounded-md ring-1 ring-zinc-800">
      <table className="min-w-full divide-y divide-zinc-800 text-sm">
        <thead className="bg-zinc-900/60">
          <tr className="text-left text-xs uppercase tracking-wide text-zinc-500">
            <th className="px-4 py-3 font-medium">Repo</th>
            <th className="px-4 py-3 font-medium">Risk</th>
            <th className="px-4 py-3 font-medium">Score</th>
            <th className="px-4 py-3 font-medium">Status</th>
            <th className="px-4 py-3 font-medium">Reasoning</th>
            <th className="px-4 py-3 font-medium">Scanned</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-zinc-800">
          {scans.map((scan) => (
            <tr key={scan.id} className="hover:bg-zinc-900/40">
              <td className="px-4 py-3 font-mono text-zinc-200">{scan.repo_name}</td>
              <td className="px-4 py-3">
                <RiskBadge level={scan.risk_level} />
              </td>
              <td className="px-4 py-3 font-mono text-zinc-400">{scan.risk_score}</td>
              <td className="px-4 py-3 text-zinc-400 capitalize">{scan.status}</td>
              <td className="max-w-md truncate px-4 py-3 text-zinc-400" title={scan.reasoning}>
                {scan.reasoning || "—"}
              </td>
              <td className="whitespace-nowrap px-4 py-3 text-zinc-500">
                {formatDate(scan.created_at)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}