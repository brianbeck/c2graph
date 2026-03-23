import { useState, type CSSProperties } from "react";
import type { GraphResponse } from "../types/graph";

export interface GraphFilters {
  riskMin: number;
  riskMax: number;
  showWallets: boolean;
  showTransactions: boolean;
  showInitiated: boolean;
  showSolTransfers: boolean;
  showTokenTransfers: boolean;
}

export const defaultFilters: GraphFilters = {
  riskMin: 0,
  riskMax: 100,
  showWallets: true,
  showTransactions: true,
  showInitiated: true,
  showSolTransfers: true,
  showTokenTransfers: true,
};

interface GraphToolbarProps {
  graph: GraphResponse | null;
  filters: GraphFilters;
  onFiltersChange: (filters: GraphFilters) => void;
}

export function GraphToolbar({
  graph,
  filters,
  onFiltersChange,
}: GraphToolbarProps) {
  const [expanded, setExpanded] = useState(false);

  if (!graph || graph.nodes.length === 0) return null;

  const walletCount = graph.nodes.filter((n) => n.type === "wallet").length;
  const txCount = graph.nodes.filter((n) => n.type === "transaction").length;

  const set = (partial: Partial<GraphFilters>) =>
    onFiltersChange({ ...filters, ...partial });

  return (
    <div
      style={{
        position: "absolute",
        top: "12px",
        left: "12px",
        zIndex: 10,
        display: "flex",
        flexDirection: "column",
        gap: "4px",
      }}
    >
      {/* Compact bar */}
      <div
        style={{
          display: "flex",
          gap: "6px",
          alignItems: "center",
          padding: "6px 10px",
          background: "rgba(26,26,46,0.92)",
          borderRadius: "6px",
          border: "1px solid #333",
          fontSize: "12px",
          color: "#aaa",
        }}
      >
        <span>
          {walletCount} wallets, {txCount} txns, {graph.links.length} edges
        </span>
        <Separator />
        <button onClick={() => setExpanded(!expanded)} style={pillBtn}>
          {expanded ? "Hide Filters" : "Filters"}
        </button>
        <button onClick={() => exportJSON(graph)} style={pillBtn}>
          JSON
        </button>
        <button onClick={() => exportCSV(graph)} style={pillBtn}>
          CSV
        </button>
      </div>

      {/* Filter panel */}
      {expanded && (
        <div
          style={{
            padding: "12px",
            background: "rgba(26,26,46,0.95)",
            borderRadius: "6px",
            border: "1px solid #333",
            fontSize: "12px",
            color: "#ccc",
            display: "flex",
            flexDirection: "column",
            gap: "10px",
            width: "340px",
          }}
        >
          {/* Risk score range */}
          <div>
            <label style={labelStyle}>
              Risk Score: {filters.riskMin} &ndash; {filters.riskMax}
            </label>
            <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
              <span style={{ color: "#888", fontSize: "11px" }}>0</span>
              <input
                type="range"
                min={0}
                max={100}
                value={filters.riskMin}
                onChange={(e) => {
                  const v = Number(e.target.value);
                  set({ riskMin: Math.min(v, filters.riskMax) });
                }}
                style={sliderStyle}
              />
              <input
                type="range"
                min={0}
                max={100}
                value={filters.riskMax}
                onChange={(e) => {
                  const v = Number(e.target.value);
                  set({ riskMax: Math.max(v, filters.riskMin) });
                }}
                style={sliderStyle}
              />
              <span style={{ color: "#888", fontSize: "11px" }}>100</span>
            </div>
          </div>

          {/* Node type toggles */}
          <div>
            <label style={labelStyle}>Node Types</label>
            <div style={{ display: "flex", gap: "8px" }}>
              <Toggle
                label={`Wallets (${walletCount})`}
                checked={filters.showWallets}
                onChange={(v) => set({ showWallets: v })}
              />
              <Toggle
                label={`Transactions (${txCount})`}
                checked={filters.showTransactions}
                onChange={(v) => set({ showTransactions: v })}
              />
            </div>
          </div>

          {/* Link type toggles */}
          <div>
            <label style={labelStyle}>Edge Types</label>
            <div style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}>
              <Toggle
                label="Initiated"
                color="#4f46e5"
                checked={filters.showInitiated}
                onChange={(v) => set({ showInitiated: v })}
              />
              <Toggle
                label="SOL"
                color="#22c55e"
                checked={filters.showSolTransfers}
                onChange={(v) => set({ showSolTransfers: v })}
              />
              <Toggle
                label="Token"
                color="#eab308"
                checked={filters.showTokenTransfers}
                onChange={(v) => set({ showTokenTransfers: v })}
              />
            </div>
          </div>

          {/* Reset */}
          <button
            onClick={() => onFiltersChange(defaultFilters)}
            style={{
              ...pillBtn,
              alignSelf: "flex-start",
              color: "#888",
            }}
          >
            Reset Filters
          </button>
        </div>
      )}
    </div>
  );
}

/* ---------- sub-components ---------- */

function Toggle({
  label,
  checked,
  onChange,
  color,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
  color?: string;
}) {
  return (
    <label
      style={{
        display: "flex",
        alignItems: "center",
        gap: "4px",
        cursor: "pointer",
        fontSize: "11px",
        color: checked ? "#ccc" : "#666",
      }}
    >
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        style={{ accentColor: color || "#4f46e5" }}
      />
      {color && (
        <span
          style={{
            width: "8px",
            height: "8px",
            borderRadius: "50%",
            background: color,
            display: "inline-block",
          }}
        />
      )}
      {label}
    </label>
  );
}

function Separator() {
  return (
    <span
      style={{
        width: "1px",
        height: "14px",
        background: "#444",
        display: "inline-block",
      }}
    />
  );
}

/* ---------- styles ---------- */

const pillBtn: CSSProperties = {
  padding: "3px 10px",
  fontSize: "11px",
  background: "#2a2a3e",
  color: "#ccc",
  border: "1px solid #444",
  borderRadius: "4px",
  cursor: "pointer",
};

const labelStyle: CSSProperties = {
  color: "#888",
  fontSize: "11px",
  display: "block",
  marginBottom: "4px",
};

const sliderStyle: CSSProperties = {
  flex: 1,
  height: "4px",
  accentColor: "#4f46e5",
};

/* ---------- export helpers ---------- */

function exportJSON(graph: GraphResponse) {
  const blob = new Blob([JSON.stringify(graph, null, 2)], {
    type: "application/json",
  });
  download(blob, "c2graph.json");
}

function exportCSV(graph: GraphResponse) {
  // Nodes CSV
  const nodeHeaders = ["id", "type", "risk_score", "tx_count", "lamports", "owner", "tags"];
  const nodeRows = graph.nodes.map((n) =>
    [
      csvEscape(n.id),
      n.type,
      n.props.risk_score ?? "",
      n.props.tx_count ?? "",
      n.props.lamports ?? "",
      csvEscape(String(n.props.owner ?? "")),
      csvEscape((n.props.tags ?? []).join("; ")),
    ].join(",")
  );

  // Links CSV
  const linkHeaders = ["source", "target", "type", "amount_lamports", "ui_amount", "mint"];
  const linkRows = graph.links.map((l) =>
    [
      csvEscape(l.source),
      csvEscape(l.target),
      l.type,
      l.props?.amount_lamports ?? "",
      l.props?.ui_amount ?? "",
      csvEscape(String(l.props?.mint ?? "")),
    ].join(",")
  );

  const csv =
    "# NODES\n" +
    nodeHeaders.join(",") +
    "\n" +
    nodeRows.join("\n") +
    "\n\n# EDGES\n" +
    linkHeaders.join(",") +
    "\n" +
    linkRows.join("\n");

  const blob = new Blob([csv], { type: "text/csv" });
  download(blob, "c2graph.csv");
}

function csvEscape(val: string): string {
  if (val.includes(",") || val.includes('"') || val.includes("\n")) {
    return `"${val.replace(/"/g, '""')}"`;
  }
  return val;
}

function download(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

/* ---------- filtering logic ---------- */

export function applyFilters(
  graph: GraphResponse,
  filters: GraphFilters
): GraphResponse {
  // Filter nodes
  const visibleNodeIds = new Set<string>();
  const filteredNodes = graph.nodes.filter((n) => {
    if (n.type === "wallet") {
      if (!filters.showWallets) return false;
      const score = (n.props.risk_score ?? 0) * 100;
      if (score < filters.riskMin || score > filters.riskMax) return false;
    }
    if (n.type === "transaction" && !filters.showTransactions) return false;
    visibleNodeIds.add(n.id);
    return true;
  });

  // Filter links: both endpoints must be visible and link type must be enabled
  const filteredLinks = graph.links.filter((l) => {
    if (!visibleNodeIds.has(l.source) || !visibleNodeIds.has(l.target))
      return false;
    if (l.type === "INITIATED" && !filters.showInitiated) return false;
    if (l.type === "TRANSFERRED_SOL" && !filters.showSolTransfers) return false;
    if (l.type === "TRANSFERRED_TOKEN" && !filters.showTokenTransfers)
      return false;
    return true;
  });

  // Remove orphan transaction nodes (no remaining links)
  const linkedNodeIds = new Set<string>();
  filteredLinks.forEach((l) => {
    linkedNodeIds.add(l.source);
    linkedNodeIds.add(l.target);
  });

  const finalNodes = filteredNodes.filter(
    (n) => n.type === "wallet" || linkedNodeIds.has(n.id)
  );

  return { nodes: finalNodes, links: filteredLinks };
}
