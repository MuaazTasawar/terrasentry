"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { ApiError, fetchDriftEvents, getToken, type DriftEvent } from "@/lib/apiClient";
import DriftTable from "@/components/DriftTable";

export default function DriftPage() {
  const router = useRouter();
  const [events, setEvents] = useState<DriftEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setEvents(await fetchDriftEvents());
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        router.push("/login");
        return;
      }
      setError("Failed to load drift events.");
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
        <h1 className="font-mono text-lg font-semibold text-zinc-100">Drift Events</h1>
        <p className="mt-1 text-sm text-zinc-500">
          Live cluster state diverging from last-known-applied, across Deployments,
          StatefulSets, and ConfigMaps.
        </p>
      </div>

      {error ? (
        <p className="text-sm text-rose-400">{error}</p>
      ) : loading ? (
        <p className="text-sm text-zinc-500">Loading…</p>
      ) : (
        <DriftTable events={events} />
      )}
    </div>
  );
}