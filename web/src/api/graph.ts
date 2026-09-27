export interface GraphNode {
  id: string;
  type: string;
  path?: string;
  name: string;
}

export interface GraphEdge {
  from: string;
  to: string;
  type: string;
}

export interface Graph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export class GraphFetchError extends Error {
  status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = "GraphFetchError";
    this.status = status;
  }
}

const API_URL = import.meta.env.VITE_API_URL;
const API_KEY = import.meta.env.VITE_API_KEY;

export async function fetchGraph(repoId: string, signal?: AbortSignal): Promise<Graph> {
  if (!repoId) {
    throw new GraphFetchError("repoId is required");
  }
  if (!API_URL || !API_KEY) {
    throw new GraphFetchError(
      "API_URL or API_KEY not configured — check your .env file",
    );
  }

  let response: Response;
  try {
    response = await fetch(`${API_URL}/v1/graph/${encodeURIComponent(repoId)}`, {
      headers: { "X-API-Key": API_KEY },
      signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      throw err; // let the caller's abort-handling logic see this, don't relabel it
    }
    throw new GraphFetchError(
      `Could not reach the API — is it running at ${API_URL}?`,
    );
  }

  if (response.status === 401) {
    throw new GraphFetchError("Invalid or missing API key", 401);
  }
  if (response.status === 429) {
    throw new GraphFetchError("Rate limit exceeded — try again shortly", 429);
  }
  if (!response.ok) {
    throw new GraphFetchError(
      `Unexpected response from API (status ${response.status})`,
      response.status,
    );
  }

  let data: unknown;
  try {
    data = await response.json();
  } catch {
    throw new GraphFetchError("API returned malformed JSON");
  }

  // Minimal shape validation — don't trust the response blindly.
  if (
    typeof data !== "object" ||
    data === null ||
    !("nodes" in data) ||
    !("edges" in data)
  ) {
    throw new GraphFetchError("API response missing expected graph shape");
  }

  const raw = data as { nodes: unknown; edges: unknown };

  // Normalize: never let a null nodes/edges field (e.g. an empty Go
  // slice serialized as JSON null) leak into a component as null.
  return {
    nodes: Array.isArray(raw.nodes) ? (raw.nodes as GraphNode[]) : [],
    edges: Array.isArray(raw.edges) ? (raw.edges as GraphEdge[]) : [],
  };
}
