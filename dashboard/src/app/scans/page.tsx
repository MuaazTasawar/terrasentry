"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { ApiError, fetchScans, getToken, type Scan } from "@/lib/apiClient";
import ScanTable from "@/components/ScanTable";

const RISK_FILTERS = ["all", "low", "medium", "high"] as const;
const STATUS_FILTERS = ["all", "scored", "approved", "rejected"] as const;

export default function ScansPage() {
  const router = useRouter();
  const [scans, setScans] = useState<Scan[]>([]);
  const [riskFilter, setRiskFilter] = useState<(typeof RISK_FILTERS)[number]>("all");
  const [statusFilter, setStatusFilter] = useState<(typeof STATUS_FILTERS)[number]>("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchScans({
        risk_level: riskFilter === "all" ? undefined : riskFilter,
        status: statusFilter === "all" ? undefined : statusFilter,
      });
      setScans(data);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        router.push("/login");
        return;
      }
      setError("Failed to load scan history.");
    } finally {
      setLoading(false);
    }
  }, [riskFilter, statusFilter, router]);

  useEffect(() => {
    if (!getToken()) {
      router.push("/login");
      return;
    }
    // Fetching data on mount is synchronizing state with an external system
    // (the API) — exactly what effects are for. See
    // https://react.dev/learn/you-might-not-need-an-effect#fetching-data
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load, router]);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="font-mono text-lg font-semibold text-zinc-100">Scan History</h1>
          <p className="mt-1 text-sm text-zinc-500">
            Every Terraform plan scanned, with risk score and outcome.
          </p>
        </div>
      </div>

      <div className="flex flex-wrap gap-4">
        <FilterGroup label="Risk" value={riskFilter} options={RISK_FILTERS} onChange={setRiskFilter} />
        <FilterGroup label="Status" value={statusFilter} options={STATUS_FILTERS} onChange={setStatusFilter} />
      </div>

      {error ? (
        <p className="text-sm text-rose-400">{error}</p>
      ) : loading ? (
        <p className="text-sm text-zinc-500">Loading…</p>
      ) : (
        <ScanTable scans={scans} />
      )}
    </div>
  );
}

function FilterGroup<T extends string>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: readonly T[];
  onChange: (value: T) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-xs font-medium uppercase tracking-wide text-zinc-500">{label}</span>
      <div className="flex gap-1 rounded-sm bg-zinc-900 p-1 ring-1 ring-zinc-800">
        {options.map((option) => (
          <button
            key={option}
            onClick={() => onChange(option)}
            className={`rounded-sm px-2.5 py-1 text-xs font-mono capitalize transition-colors ${
              value === option
                ? "bg-zinc-700 text-zinc-100"
                : "text-zinc-500 hover:text-zinc-300"
            }`}
          >
            {option}
          </button>
        ))}
      </div>
    </div>
  );
}