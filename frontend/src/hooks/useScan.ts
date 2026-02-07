import { useState, useEffect, useCallback, useRef } from "react";
import { submitScan, getJobStatus } from "../lib/api";
import type { JobStatus } from "../types/graph";

interface UseScanResult {
  scan: (address: string, depth: number) => Promise<void>;
  status: JobStatus | null;
  isScanning: boolean;
  error: string | null;
}

export function useScan(): UseScanResult {
  const [status, setStatus] = useState<JobStatus | null>(null);
  const [isScanning, setIsScanning] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, []);

  const scan = useCallback(async (address: string, depth: number) => {
    setError(null);
    setIsScanning(true);
    setStatus(null);

    // Stop any existing polling
    if (pollRef.current) clearInterval(pollRef.current);

    try {
      const res = await submitScan({ address, depth });

      if (res.status === "ready") {
        setStatus({
          job_id: "",
          root_address: address,
          requested_depth: depth,
          status: "complete",
          wallets_processed: 0,
          wallets_cap: 0,
          created_at: "",
          updated_at: "",
        });
        setIsScanning(false);
        return;
      }

      // Start polling for job status
      if (res.job_id) {
        pollRef.current = setInterval(async () => {
          try {
            const jobStatus = await getJobStatus(res.job_id!);
            setStatus(jobStatus);

            if (
              jobStatus.status === "complete" ||
              jobStatus.status === "capped" ||
              jobStatus.status === "failed"
            ) {
              if (pollRef.current) clearInterval(pollRef.current);
              setIsScanning(false);
            }
          } catch (err) {
            console.error("Poll error:", err);
          }
        }, 2000);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Scan failed");
      setIsScanning(false);
    }
  }, []);

  return { scan, status, isScanning, error };
}
