import { useState } from "react";
import { Canvas } from "@react-three/fiber";
import { OrbitControls, Line, Text, Billboard } from "@react-three/drei";
import { useGraphLayout } from "../hooks/useGraphLayout";
import type { GraphNode } from "../api/graph";

interface NodeMeshProps {
  node: GraphNode;
  position: [number, number, number];
  onSelect: (node: GraphNode) => void;
  isSelected: boolean;
}

function NodeMesh({ node, position, onSelect, isSelected }: NodeMeshProps) {
  const isFile = node.type === "File";
  const color = isSelected ? "#ffffff" : isFile ? "#61dafb" : "#f0a500";

  return (
    <group position={position}>
      <mesh
        onClick={(e) => {
          e.stopPropagation();
          onSelect(node);
        }}
      >
        {isFile ? (
          <boxGeometry args={[6, 6, 6]} />
        ) : (
          <sphereGeometry args={[3, 16, 16]} />
        )}
        <meshStandardMaterial color={color} />
      </mesh>
      <Billboard position={[0, 8, 0]}>
        <Text fontSize={3} color="#eee" anchorX="center" anchorY="middle">
          {node.name.split("\\").pop()}
        </Text>
      </Billboard>
    </group>
  );
}
export function GraphView3D() {
const repoId = "test-real-repo";
  const { graph, positions, loading, error } = useGraphLayout(repoId);
  const [selected, setSelected] = useState<GraphNode | null>(null);

  if (loading) return <div>Loading graph…</div>;
  if (error) return <div style={{ color: "red" }}>Error: {error}</div>;
  if (!graph) return null;

  return (
    <div style={{ display: "flex", height: "600px" }}>
      <div style={{ flex: 1 }}>
        <Canvas camera={{ position: [0, 0, 400], fov: 60 }}>
          <ambientLight intensity={0.6} />
          <pointLight position={[100, 100, 100]} intensity={0.8} />
          <OrbitControls enableDamping />

          {graph.edges.map((e, i) => {
            const from = positions.get(e.from);
            const to = positions.get(e.to);
            if (!from || !to) return null;
            return (
              <Line
                key={i}
                points={[
                  [from.x, from.y, from.z],
                  [to.x, to.y, to.z],
                ]}
                color="#555"
                lineWidth={1}
              />
            );
          })}

          {graph.nodes.map((n) => {
            const pos = positions.get(n.id);
            if (!pos) return null;
            return (
              <NodeMesh
                key={n.id}
                node={n}
                position={[pos.x, pos.y, pos.z]}
                onSelect={setSelected}
                isSelected={selected?.id === n.id}
              />
            );
          })}
        </Canvas>
      </div>

      <div
        style={{
          width: "260px",
          padding: "16px",
          background: "#1a1a1a",
          color: "#ccc",
          overflowY: "auto",
        }}
      >
        <h3>Details</h3>
        {selected ? (
          <>
            <p><strong>Name:</strong> {selected.name.split("\\").pop()}</p>
            <p><strong>Type:</strong> {selected.type}</p>
            {selected.path && (
              <p style={{ wordBreak: "break-all", fontSize: "12px" }}>
                <strong>Path:</strong> {selected.path}
              </p>
            )}
          </>
        ) : (
          <p>Click a node to see details</p>
        )}
      </div>
    </div>
  );
}