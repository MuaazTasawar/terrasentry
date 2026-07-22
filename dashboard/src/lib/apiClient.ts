/**
 * Thin fetch wrapper around the existing Go API — the dashboard talks to
 * the same backend the mobile app does, never a second/duplicate backend.
 */

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") ?? "http://localhost:8080";

const TOKEN_STORAGE_KEY = "terrasentry_dashboard_token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_STORAGE_KEY);
}

export function setToken(token: string): void {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(TOKEN_STORAGE_KEY, token);
}

export function clearToken(): void {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(TOKEN_STORAGE_KEY);
}

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });

  if (!res.ok) {
    let detail = res.statusText;
    try {
      const body = await res.json();
      detail = body.error ?? detail;
    } catch {
      // response wasn't JSON — fall back to statusText
    }
    throw new ApiError(detail, res.status);
  }

  return res.json() as Promise<T>;
}

// --- Types (mirrors the Go API's response shapes) ---

export type Scan = {
  id: string;
  repo_name: string;
  plan_summary: string;
  risk_score: number;
  risk_level: "low" | "medium" | "high";
  reasoning: string;
  status: string;
  created_at: string;
};

export type DriftEvent = {
  id: string;
  resource_kind: string;
  resource_name: string;
  namespace: string;
  diff: string;
  detected_at: string;
};

export type AuditEntry = {
  id: string;
  plan_scan_id: string;
  repo_name: string;
  decision: "approved" | "rejected";
  decided_by: string;
  decided_at: string;
};

export type ScanFilters = {
  status?: string;
  risk_level?: string;
};

// --- API calls ---

export async function login(email: string, password: string): Promise<string> {
  const data = await request<{ token: string }>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
  return data.token;
}

export async function fetchScans(filters: ScanFilters = {}): Promise<Scan[]> {
  const params = new URLSearchParams();
  if (filters.status) params.set("status", filters.status);
  if (filters.risk_level) params.set("risk_level", filters.risk_level);
  const query = params.toString();
  const data = await request<{ scans: Scan[] }>(`/api/v1/scans${query ? `?${query}` : ""}`);
  return data.scans;
}

export async function fetchDriftEvents(): Promise<DriftEvent[]> {
  const data = await request<{ events: DriftEvent[] }>("/api/v1/drift-events");
  return data.events;
}

export async function fetchAuditTrail(): Promise<AuditEntry[]> {
  const data = await request<{ audit: AuditEntry[] }>("/api/v1/audit");
  return data.audit;
}