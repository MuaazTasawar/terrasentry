import type { AuditEntry } from "@/lib/apiClient";

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

export default function AuditTable({ entries }: { entries: AuditEntry[] }) {
  if (entries.length === 0) {
    return (
      <p className="py-12 text-center text-sm text-zinc-500">
        No approval decisions recorded yet.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto rounded-md ring-1 ring-zinc-800">
      <table className="min-w-full divide-y divide-zinc-800 text-sm">
        <thead className="bg-zinc-900/60">
          <tr className="text-left text-xs uppercase tracking-wide text-zinc-500">
            <th className="px-4 py-3 font-medium">Repo</th>
            <th className="px-4 py-3 font-medium">Decision</th>
            <th className="px-4 py-3 font-medium">Decided By</th>
            <th className="px-4 py-3 font-medium">Decided At</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-zinc-800">
          {entries.map((entry) => (
            <tr key={entry.id} className="hover:bg-zinc-900/40">
              <td className="px-4 py-3 font-mono text-zinc-200">{entry.repo_name}</td>
              <td className="px-4 py-3">
                <span
                  className={`inline-flex items-center rounded-sm px-2 py-0.5 text-xs font-mono font-medium uppercase tracking-wide ring-1 ring-inset ${
                    entry.decision === "approved"
                      ? "bg-emerald-950 text-emerald-400 ring-emerald-800"
                      : "bg-rose-950 text-rose-400 ring-rose-800"
                  }`}
                >
                  {entry.decision}
                </span>
              </td>
              <td className="px-4 py-3 text-zinc-400">{entry.decided_by}</td>
              <td className="whitespace-nowrap px-4 py-3 text-zinc-500">
                {formatDate(entry.decided_at)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}