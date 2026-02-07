import { useState, useCallback } from "react";
import { getGraph } from "../lib/api";
import type { GraphResponse } from "../types/graph";

interface UseGraphResult {
  graph: GraphResponse | null;
  loading: boolean;
  error: string | null;
  fetchGraph: (address: string, depth: number) => Promise<void>;
}

export function useGraph(): UseGraphResult {
  const [graph, setGraph] = useState<GraphResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchGraph = useCallback(async (address: string, depth: number) => {
    setLoading(true);
    setError(null);
    try {
      const data = await getGraph(address, depth);
      setGraph(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch graph");
    } finally {
      setLoading(false);
    }
  }, []);

  return { graph, loading, error, fetchGraph };
}
