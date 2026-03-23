import { useRef, useCallback, useEffect, useMemo } from "react";
import ForceGraph2D from "react-force-graph-2d";
import type { GraphNode, GraphLink, GraphResponse } from "../types/graph";
import type { GraphFilters } from "./GraphToolbar";
import { applyFilters } from "./GraphToolbar";

interface GraphViewProps {
  graph: GraphResponse | null;
  filters: GraphFilters;
  onNodeClick: (node: GraphNode) => void;
}

function riskColor(score: number): string {
  // Green (safe) -> Yellow (medium) -> Red (high risk)
  if (score < 0.3) return `rgb(${Math.round(score * 3.3 * 255)}, 200, 80)`;
  if (score < 0.6) return `rgb(255, ${Math.round((1 - (score - 0.3) * 3.3) * 200)}, 50)`;
  return `rgb(239, ${Math.round((1 - (score - 0.6) * 2.5) * 68)}, 68)`;
}

export function GraphView({ graph, filters, onNodeClick }: GraphViewProps) {
  const fgRef = useRef<any>(null);

  // Zoom to fit when graph data changes
  useEffect(() => {
    if (fgRef.current && graph && graph.nodes.length > 0) {
      setTimeout(() => {
        fgRef.current?.zoomToFit(400, 50);
      }, 500);
    }
  }, [graph]);

  const handleNodeClick = useCallback(
    (node: any) => {
      if (!graph) return;
      const graphNode = graph.nodes.find((n) => n.id === node.id);
      if (graphNode) onNodeClick(graphNode);
    },
    [graph, onNodeClick]
  );

  // All hooks must be called before any early return
  const emptyGraph = { nodes: [] as GraphNode[], links: [] as GraphLink[] };
  const filtered = useMemo(
    () => (graph && graph.nodes.length > 0 ? applyFilters(graph, filters) : emptyGraph),
    [graph, filters]
  );

  const graphData = useMemo(
    () => ({
      nodes: filtered.nodes.map((n) => ({
        id: n.id,
        type: n.type,
        riskScore: n.props.risk_score ?? 0,
        botLikelihood: n.props.bot_likelihood ?? 0,
        txCount: n.props.tx_count ?? 1,
        label:
          n.type === "wallet"
            ? `${n.id.slice(0, 4)}...${n.id.slice(-4)}`
            : `tx:${n.id.slice(0, 6)}`,
      })),
      links: filtered.links.map((l) => ({
        source: l.source,
        target: l.target,
        type: l.type,
        label: formatLinkLabel(l),
      })),
    }),
    [filtered]
  );

  if (!graph || graph.nodes.length === 0) {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          flex: 1,
          color: "#555",
          fontSize: "16px",
        }}
      >
        Enter a Solana wallet address and click Scan to visualize the transaction graph.
      </div>
    );
  }

  return (
    <ForceGraph2D
      ref={fgRef}
      graphData={graphData}
      backgroundColor="#0a0a0f"
      nodeLabel="label"
      nodeVal={(node: any) =>
        node.type === "wallet" ? Math.max(2, Math.log2(node.txCount + 1)) : 1
      }
      nodeColor={(node: any) =>
        node.type === "wallet" ? riskColor(node.riskScore) : "#555"
      }
      nodeCanvasObject={(node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
        const size = node.type === "wallet" ? 6 : 3;
        const color = node.type === "wallet" ? riskColor(node.riskScore) : "#555";
        const isBot = node.botLikelihood >= 0.5;

        if (node.type === "wallet") {
          // Bot indicator: dashed cyan ring
          if (isBot) {
            ctx.beginPath();
            ctx.arc(node.x, node.y, size + 3, 0, 2 * Math.PI);
            ctx.setLineDash([3, 2]);
            ctx.strokeStyle = "#06b6d4";
            ctx.lineWidth = 1.5;
            ctx.stroke();
            ctx.setLineDash([]);
          }

          // Circle for wallets
          ctx.beginPath();
          ctx.arc(node.x, node.y, size, 0, 2 * Math.PI);
          ctx.fillStyle = color;
          ctx.fill();
          ctx.strokeStyle = "#fff";
          ctx.lineWidth = 0.5;
          ctx.stroke();
        } else {
          // Small diamond for transactions
          ctx.beginPath();
          ctx.moveTo(node.x, node.y - size);
          ctx.lineTo(node.x + size, node.y);
          ctx.lineTo(node.x, node.y + size);
          ctx.lineTo(node.x - size, node.y);
          ctx.closePath();
          ctx.fillStyle = color;
          ctx.fill();
        }

        // Draw label at reasonable zoom
        if (globalScale > 1.5 && node.type === "wallet") {
          const label = isBot ? `[BOT] ${node.label}` : node.label;
          const fontSize = 10 / globalScale;
          ctx.font = `${fontSize}px monospace`;
          ctx.textAlign = "center";
          ctx.textBaseline = "top";
          ctx.fillStyle = isBot ? "#06b6d4" : "#aaa";
          ctx.fillText(label, node.x, node.y + size + 2);
        }
      }}
      linkColor={(link: any) => {
        switch (link.type) {
          case "INITIATED":
            return "#4f46e5";
          case "TRANSFERRED_SOL":
            return "#22c55e";
          case "TRANSFERRED_TOKEN":
            return "#eab308";
          default:
            return "#333";
        }
      }}
      linkDirectionalArrowLength={4}
      linkDirectionalArrowRelPos={0.8}
      linkWidth={0.5}
      onNodeClick={handleNodeClick}
      cooldownTicks={100}
      warmupTicks={50}
    />
  );
}

function formatLinkLabel(link: {
  type: string;
  props?: Record<string, unknown>;
}): string {
  if (!link.props) return link.type;
  if (link.type === "TRANSFERRED_SOL" && link.props.amount_lamports) {
    const sol = (link.props.amount_lamports as number) / 1_000_000_000;
    return `${Math.abs(sol).toFixed(4)} SOL`;
  }
  if (link.type === "TRANSFERRED_TOKEN" && link.props.ui_amount) {
    return `${link.props.ui_amount} ${link.props.mint ? (link.props.mint as string).slice(0, 4) + "..." : ""}`;
  }
  return link.type;
}
