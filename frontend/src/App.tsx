import { useState, useCallback, useEffect } from "react";
import { AuthProvider, useAuth } from "./auth/AuthProvider";
import { ErrorBoundary } from "./components/ErrorBoundary";
import { LoginPage } from "./auth/LoginPage";
import { SearchBar } from "./components/SearchBar";
import { GraphView } from "./components/GraphView";
import { NodeDetail } from "./components/NodeDetail";
import { StatusBar } from "./components/StatusBar";
import { GraphToolbar, defaultFilters } from "./components/GraphToolbar";
import type { GraphFilters } from "./components/GraphToolbar";
import { useScan } from "./hooks/useScan";
import { useGraph } from "./hooks/useGraph";
import type { GraphNode } from "./types/graph";

function C2GraphApp() {
  const { user, loading: authLoading, signOut } = useAuth();
  const { scan, status, isScanning, error: scanError } = useScan();
  const { graph, fetchGraph, error: graphError } = useGraph();
  const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);
  const [currentAddress, setCurrentAddress] = useState<string>("");
  const [currentDepth, setCurrentDepth] = useState<number>(1);
  const [filters, setFilters] = useState<GraphFilters>(defaultFilters);

  const handleScan = useCallback(
    (address: string, depth: number) => {
      setCurrentAddress(address);
      setCurrentDepth(depth);
      setSelectedNode(null);
      scan(address, depth);
    },
    [scan]
  );

  // Fetch graph when scan completes
  useEffect(() => {
    if (
      status &&
      (status.status === "complete" || status.status === "capped") &&
      currentAddress
    ) {
      fetchGraph(currentAddress, currentDepth);
    }
  }, [status, currentAddress, currentDepth, fetchGraph]);

  if (authLoading) {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          minHeight: "100vh",
          background: "#0a0a0f",
          color: "#888",
        }}
      >
        Loading...
      </div>
    );
  }

  // Check if Supabase is configured
  const supabaseConfigured =
    import.meta.env.VITE_SUPABASE_URL &&
    import.meta.env.VITE_SUPABASE_URL !== "https://your-project.supabase.co";

  if (supabaseConfigured && !user) {
    return <LoginPage />;
  }

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        background: "#0a0a0f",
        color: "#e0e0e0",
      }}
    >
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "8px 16px",
          background: "#111",
          borderBottom: "1px solid #333",
        }}
      >
        <h1
          style={{
            margin: 0,
            fontSize: "18px",
            fontWeight: "bold",
            letterSpacing: "0.05em",
          }}
        >
          C2GRAPH
        </h1>
        {user && (
          <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
            <span style={{ color: "#888", fontSize: "13px" }}>
              {user.email}
            </span>
            <button
              onClick={signOut}
              style={{
                padding: "4px 12px",
                fontSize: "12px",
                background: "#333",
                color: "#ccc",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              Sign Out
            </button>
          </div>
        )}
      </div>

      <SearchBar onScan={handleScan} disabled={isScanning} />
      <StatusBar status={status} isScanning={isScanning} error={scanError || graphError} />

      {/* Main content */}
      <div style={{ flex: 1, position: "relative", overflow: "hidden" }}>
        <GraphToolbar
          graph={graph}
          filters={filters}
          onFiltersChange={setFilters}
        />
        <GraphView
          graph={graph}
          filters={filters}
          onNodeClick={setSelectedNode}
        />
        <NodeDetail node={selectedNode} onClose={() => setSelectedNode(null)} />
      </div>
    </div>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <C2GraphApp />
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App;
