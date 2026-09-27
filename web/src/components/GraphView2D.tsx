import { useEffect, useRef, useState } from "react";
import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCenter,
  type SimulationNodeDatum,
} from "d3-force";
import { fetchGraph, GraphFetchError, type Graph } from "../api/graph";

interface SimNode extends SimulationNodeDatum {
  id: string;
  type: string;
  name: string;
}

interface SimLink {
  source: string | SimNode;
  target: string | SimNode;
}

export function GraphView2D(repoId: string) {
  const [graph, setGraph] = useState<Graph | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [positions, setPositions] = useState<
    Map<string, { x: number; y: number }>
  >(new Map());
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    const controller = new AbortController();

    fetchGraph(repoId, controller.signal)
      .then((g) => {
        if (!mountedRef.current) return;
        setGraph(g);
      })
      .catch((err) => {
        if (!mountedRef.current) return;
        if (err instanceof DOMException && err.name === "AbortError") {
          return; // intentional cancellation, not a real failure
        }
        if (err instanceof GraphFetchError) {
          setError(err.message);
        } else {
          setError("Unexpected error loading graph");
        }
      })
      .finally(() => {
        if (mountedRef.current) setLoading(false);
      });

    return () => {
      mountedRef.current = false;
      controller.abort();
    };
  }, [repoId]);

  useEffect(() => {
    if (!graph) return;

    const nodes: SimNode[] = graph.nodes.map((n) => ({ ...n }));
    const links: SimLink[] = graph.edges.map((e) => ({
      source: e.from,
      target: e.to,
    }));

    const simulation = forceSimulation<SimNode>(nodes)
      .force(
        "link",
        forceLink<SimNode, SimLink>(links)
          .id((d) => d.id)
          .distance(80),
      )
      .force("charge", forceManyBody().strength(-150))
      .force("center", forceCenter(400, 300))
      .on("tick", () => {
        const next = new Map<string, { x: number; y: number }>();
        for (const n of nodes) {
          if (n.x !== undefined && n.y !== undefined) {
            next.set(n.id, { x: n.x, y: n.y });
          }
        }
        setPositions(next);
      });

    return () => {
      simulation.stop();
    };
  }, [graph]);

  if (loading) return <div>Loading graph…</div>;
  if (error) return <div style={{ color: "red" }}>Error: {error}</div>;
  if (!graph) return null;

  return (
    <svg width={800} height={600} style={{ background: "#111" }}>
      {graph.edges.map((e, i) => {
        const from = positions.get(e.from);
        const to = positions.get(e.to);
        if (!from || !to) return null;
        return (
          <line
            key={i}
            x1={from.x}
            y1={from.y}
            x2={to.x}
            y2={to.y}
            stroke="#555"
            strokeWidth={1}
          />
        );
      })}
      {graph.nodes.map((n) => {
        const pos = positions.get(n.id);
        if (!pos) return null;
        return (
          <g key={n.id}>
            <circle
              cx={pos.x}
              cy={pos.y}
              r={n.type === "File" ? 8 : 5}
              fill={n.type === "File" ? "#61dafb" : "#f0a500"}
            />
            <text x={pos.x + 10} y={pos.y + 4} fontSize={10} fill="#ccc">
              {n.name.split("\\").pop()}
            </text>
          </g>
        );
      })}
    </svg>
  );
}
