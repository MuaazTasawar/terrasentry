import type { DriftEvent } from "@/lib/apiClient";

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

const KIND_STYLES: Record<string, string> = {
  Deployment: "bg-sky-950 text-sky-400 ring-sky-800",
  StatefulSet: "bg-violet-950 text-violet-400 ring-violet-800",
  ConfigMap: "bg-cyan-950 text-cyan-400 ring-cyan-800",
};

export default function DriftTable({ events }: { events: DriftEvent[] }) {
  if (events.length === 0) {
    return (
      <p className="py-12 text-center text-sm text-zinc-500">
        No drift detected yet — the operator hasn&apos;t flagged any divergence.
      </p>
    );
  }

  return (
    <div className="overflow-x-auto rounded-md ring-1 ring-zinc-800">
      <table className="min-w-full divide-y divide-zinc-800 text-sm">
        <thead className="bg-zinc-900/60">
          <tr className="text-left text-xs uppercase tracking-wide text-zinc-500">
            <th className="px-4 py-3 font-medium">Kind</th>
            <th className="px-4 py-3 font-medium">Resource</th>
            <th className="px-4 py-3 font-medium">Namespace</th>
            <th className="px-4 py-3 font-medium">Diff</th>
            <th className="px-4 py-3 font-medium">Detected</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-zinc-800">
          {events.map((event) => (
            <tr key={event.id} className="hover:bg-zinc-900/40">
              <td className="px-4 py-3">
                <span
                  className={`inline-flex items-center rounded-sm px-2 py-0.5 text-xs font-mono font-medium ring-1 ring-inset ${
                    KIND_STYLES[event.resource_kind] ??
                    "bg-zinc-800 text-zinc-400 ring-zinc-700"
                  }`}
                >
                  {event.resource_kind}
                </span>
              </td>
              <td className="px-4 py-3 font-mono text-zinc-200">{event.resource_name}</td>
              <td className="px-4 py-3 font-mono text-zinc-400">{event.namespace}</td>
              <td className="max-w-lg truncate px-4 py-3 text-zinc-400" title={event.diff}>
                {event.diff}
              </td>
              <td className="whitespace-nowrap px-4 py-3 text-zinc-500">
                {formatDate(event.detected_at)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}