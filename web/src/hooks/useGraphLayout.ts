import { useEffect, useRef, useState } from "react";
import { forceSimulation, forceLink, forceManyBody, forceCenter, type SimulationNodeDatum } from "d3-force";
import { fetchGraph, GraphFetchError, type Graph, type GraphNode } from "../api/graph";

export interface LayoutNode extends SimulationNodeDatum {
  id: string;
  type: string;
  name: string;
  path?: string;
  z: number;
}

interface SimLink {
  source: string | LayoutNode;
  target: string | LayoutNode;
}

export interface GraphLayoutResult {
  graph: Graph | null;
  positions: Map<string, { x: number; y: number; z: number }>;
  loading: boolean;
  error: string | null;
}

function zFor(type: string): number {
  // Simple depth separation by node type — files sit forward, modules further back.
  return type === "File" ? 0 : -100;
}

export function useGraphLayout(repoId: string): GraphLayoutResult {
  const [graph, setGraph] = useState<Graph | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [positions, setPositions] = useState<Map<string, { x: number; y: number; z: number }>>(new Map());
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
        if (err instanceof DOMException && err.name === "AbortError") return;
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

    const nodes: LayoutNode[] = graph.nodes.map((n: GraphNode) => ({
      ...n,
      z: zFor(n.type),
    }));
    const links: SimLink[] = graph.edges.map((e) => ({ source: e.from, target: e.to }));

    const simulation = forceSimulation<LayoutNode>(nodes)
      .force("link", forceLink<LayoutNode, SimLink>(links).id((d) => d.id).distance(60))
      .force("charge", forceManyBody().strength(-300))
      .force("center", forceCenter(0, 0))
      .on("tick", () => {
        const next = new Map<string, { x: number; y: number; z: number }>();
        for (const n of nodes) {
          if (n.x !== undefined && n.y !== undefined) {
            next.set(n.id, { x: n.x, y: n.y, z: n.z });
          }
        }
        setPositions(next);
      });

    return () => {
      simulation.stop();
    };
  }, [graph]);

  return { graph, positions, loading, error };
}