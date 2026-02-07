import { supabase } from "./supabase";
import type {
  GraphResponse,
  ScanRequest,
  ScanResponse,
  JobStatus,
} from "../types/graph";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

async function getAuthHeaders(): Promise<Record<string, string>> {
  const {
    data: { session },
  } = await supabase.auth.getSession();
  if (session?.access_token) {
    return {
      Authorization: `Bearer ${session.access_token}`,
      "Content-Type": "application/json",
    };
  }
  // Dev mode: no auth
  return { "Content-Type": "application/json" };
}

export async function submitScan(req: ScanRequest): Promise<ScanResponse> {
  const headers = await getAuthHeaders();
  const res = await fetch(`${API_URL}/api/scan`, {
    method: "POST",
    headers,
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function getJobStatus(jobId: string): Promise<JobStatus> {
  const headers = await getAuthHeaders();
  const res = await fetch(`${API_URL}/api/status/${jobId}`, { headers });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return res.json();
}

export async function getGraph(
  address: string,
  depth: number
): Promise<GraphResponse> {
  const headers = await getAuthHeaders();
  const res = await fetch(
    `${API_URL}/api/graph/${address}?depth=${depth}`,
    { headers }
  );
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return res.json();
}
