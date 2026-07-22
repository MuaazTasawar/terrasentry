"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { ApiError, fetchAuditTrail, getToken, type AuditEntry } from "@/lib/apiClient";
import AuditTable from "@/components/AuditTable";

export default function AuditPage() {
  const router = useRouter();
  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setEntries(await fetchAuditTrail());
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        router.push("/login");
        return;
      }
      setError("Failed to load approval audit trail.");
    } finally {
      setLoading(false);
    }
  }, [router]);

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
      <div>
        <h1 className="font-mono text-lg font-semibold text-zinc-100">Approval Audit Trail</h1>
        <p className="mt-1 text-sm text-zinc-500">
          Who approved or rejected each scan, and when.
        </p>
      </div>

      {error ? (
        <p className="text-sm text-rose-400">{error}</p>
      ) : loading ? (
        <p className="text-sm text-zinc-500">Loading…</p>
      ) : (
        <AuditTable entries={entries} />
      )}
    </div>
  );
}