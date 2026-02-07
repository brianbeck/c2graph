import type { JobStatus } from "../types/graph";

interface StatusBarProps {
  status: JobStatus | null;
  isScanning: boolean;
  error: string | null;
}

export function StatusBar({ status, isScanning, error }: StatusBarProps) {
  if (error) {
    return (
      <div style={{ padding: "8px 16px", background: "#2d1b1b", color: "#ef4444", fontSize: "13px" }}>
        Error: {error}
      </div>
    );
  }

  if (!status && !isScanning) return null;

  const statusColor: Record<string, string> = {
    queued: "#eab308",
    processing: "#3b82f6",
    complete: "#22c55e",
    capped: "#f59e0b",
    failed: "#ef4444",
  };

  const statusText = status?.status || "queued";
  const color = statusColor[statusText] || "#888";

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: "12px",
        padding: "8px 16px",
        background: "#1a1a2e",
        borderBottom: "1px solid #333",
        fontSize: "13px",
        color: "#ccc",
      }}
    >
      <span style={{ color }}>
        {isScanning ? "\u25CF" : "\u25CB"} {statusText.toUpperCase()}
      </span>
      {status && status.wallets_processed > 0 && (
        <span>
          {status.wallets_processed.toLocaleString()} / {status.wallets_cap.toLocaleString()} wallets
        </span>
      )}
      {status?.status === "capped" && (
        <span style={{ color: "#f59e0b" }}>
          Wallet limit reached. Some peripheral wallets may not be fully explored.
        </span>
      )}
      {isScanning && (
        <div
          style={{
            flex: 1,
            height: "4px",
            background: "#333",
            borderRadius: "2px",
            overflow: "hidden",
          }}
        >
          <div
            style={{
              width: status
                ? `${Math.min((status.wallets_processed / Math.max(status.wallets_cap, 1)) * 100, 100)}%`
                : "0%",
              height: "100%",
              background: color,
              transition: "width 0.3s",
            }}
          />
        </div>
      )}
    </div>
  );
}
