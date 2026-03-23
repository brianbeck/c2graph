import type { GraphNode } from "../types/graph";

interface NodeDetailProps {
  node: GraphNode | null;
  onClose: () => void;
}

function riskColor(score: number): string {
  if (score < 0.3) return "#22c55e";
  if (score < 0.6) return "#eab308";
  return "#ef4444";
}

export function NodeDetail({ node, onClose }: NodeDetailProps) {
  if (!node) return null;

  const isWallet = node.type === "wallet";
  const riskScore = node.props.risk_score ?? 0;

  return (
    <div
      style={{
        position: "absolute",
        right: 0,
        top: 0,
        bottom: 0,
        width: "420px",
        boxSizing: "border-box",
        background: "#1a1a2e",
        borderLeft: "1px solid #333",
        padding: "20px",
        overflowY: "auto",
        color: "#e0e0e0",
        fontSize: "13px",
        zIndex: 10,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "16px",
        }}
      >
        <h3 style={{ margin: 0, fontSize: "16px" }}>
          {isWallet ? "Wallet" : "Transaction"}
        </h3>
        <button
          onClick={onClose}
          style={{
            background: "none",
            border: "none",
            color: "#888",
            cursor: "pointer",
            fontSize: "18px",
          }}
        >
          x
        </button>
      </div>

      {/* Address / Signature */}
      <div style={{ marginBottom: "12px" }}>
        <label style={{ color: "#888", fontSize: "11px", display: "block" }}>
          {isWallet ? "Address" : "Signature"}
        </label>
        <div
          style={{
            fontFamily: "monospace",
            fontSize: "12px",
            wordBreak: "break-all",
            padding: "6px",
            background: "#0a0a0f",
            borderRadius: "4px",
            cursor: "pointer",
          }}
          onClick={() => navigator.clipboard.writeText(node.id)}
          title="Click to copy"
        >
          {node.id}
        </div>
      </div>

      {/* Risk Score (wallets only) */}
      {isWallet && (
        <div style={{ marginBottom: "12px" }}>
          <label style={{ color: "#888", fontSize: "11px", display: "block" }}>
            Risk Score
          </label>
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <div
              style={{
                flex: 1,
                height: "8px",
                background: "#333",
                borderRadius: "4px",
                overflow: "hidden",
              }}
            >
              <div
                style={{
                  width: `${riskScore * 100}%`,
                  height: "100%",
                  background: riskColor(riskScore),
                  borderRadius: "4px",
                }}
              />
            </div>
            <span
              style={{
                color: riskColor(riskScore),
                fontWeight: "bold",
                minWidth: "36px",
              }}
            >
              {(riskScore * 100).toFixed(0)}
            </span>
          </div>
        </div>
      )}

      {/* Bot Likelihood (wallets only) */}
      {isWallet && (node.props.bot_likelihood ?? 0) > 0 && (
        <div style={{ marginBottom: "12px" }}>
          <label style={{ color: "#888", fontSize: "11px", display: "block" }}>
            Bot Likelihood
          </label>
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <div
              style={{
                flex: 1,
                height: "8px",
                background: "#333",
                borderRadius: "4px",
                overflow: "hidden",
              }}
            >
              <div
                style={{
                  width: `${(node.props.bot_likelihood ?? 0) * 100}%`,
                  height: "100%",
                  background: (node.props.bot_likelihood ?? 0) >= 0.5 ? "#06b6d4" : "#334155",
                  borderRadius: "4px",
                }}
              />
            </div>
            <span
              style={{
                color: (node.props.bot_likelihood ?? 0) >= 0.5 ? "#06b6d4" : "#888",
                fontWeight: "bold",
                minWidth: "36px",
              }}
            >
              {((node.props.bot_likelihood ?? 0) * 100).toFixed(0)}%
            </span>
          </div>
        </div>
      )}

      {/* Tags */}
      {isWallet && node.props.tags && node.props.tags.length > 0 && (
        <div style={{ marginBottom: "12px" }}>
          <label style={{ color: "#888", fontSize: "11px", display: "block" }}>
            Tags
          </label>
          <div style={{ display: "flex", gap: "4px", flexWrap: "wrap" }}>
            {node.props.tags.map((tag) => (
              <span
                key={tag}
                style={{
                  padding: "2px 8px",
                  background: "#333",
                  borderRadius: "12px",
                  fontSize: "11px",
                }}
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Risk Factors */}
      {isWallet &&
        node.props.risk_factors &&
        node.props.risk_factors.length > 0 && (
          <div style={{ marginBottom: "12px" }}>
            <label
              style={{ color: "#888", fontSize: "11px", display: "block" }}
            >
              Risk Factors
            </label>
            <div style={{ display: "flex", gap: "4px", flexWrap: "wrap" }}>
              {node.props.risk_factors.map((factor) => (
                <span
                  key={factor}
                  style={{
                    padding: "2px 8px",
                    background: "#3b1111",
                    color: "#ef4444",
                    borderRadius: "12px",
                    fontSize: "11px",
                  }}
                >
                  {factor}
                </span>
              ))}
            </div>
          </div>
        )}

      {/* Properties */}
      <div>
        <label style={{ color: "#888", fontSize: "11px", display: "block", marginBottom: "6px" }}>
          Properties
        </label>
        <table style={{ width: "100%", borderCollapse: "collapse", tableLayout: "fixed" }}>
          <tbody>
            {Object.entries(node.props)
              .filter(
                ([key]) =>
                  !["type", "tags", "risk_factors", "risk_score", "bot_likelihood"].includes(key)
              )
              .map(([key, value]) => (
                <tr key={key} style={{ borderBottom: "1px solid #222" }}>
                  <td
                    style={{
                      padding: "4px 8px 4px 0",
                      color: "#888",
                      verticalAlign: "top",
                      whiteSpace: "nowrap",
                      width: "120px",
                    }}
                  >
                    {key}
                  </td>
                  <td
                    style={{
                      padding: "4px 0",
                      fontFamily: "monospace",
                      fontSize: "12px",
                      wordBreak: "break-all",
                      overflowWrap: "anywhere",
                    }}
                  >
                    {formatValue(value)}
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function formatValue(val: unknown): string {
  if (val === null || val === undefined) return "-";
  if (typeof val === "number") {
    if (val > 1_000_000) return val.toLocaleString();
    return String(val);
  }
  if (Array.isArray(val)) return val.join(", ");
  return String(val);
}
