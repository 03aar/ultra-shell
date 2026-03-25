'use client';

import { useCallback, useEffect, useState } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  Node,
  Edge as FlowEdge,
  useNodesState,
  useEdgesState,
  MarkerType,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { ExecutionNode, Edge, getContextGraph } from '@/lib/api';

const categoryColors: Record<string, string> = {
  filesystem_mutation: '#fbbf24',
  network: '#60a5fa',
  git: '#4ade80',
  build: '#a78bfa',
  docker: '#38bdf8',
  process_spawn: '#fb923c',
  env_change: '#c084fc',
  read_only: '#666666',
  package_manager: '#f472b6',
  unknown: '#666666',
};

interface ExecutionGraphProps {
  sessionId?: string;
  onNodeClick?: (node: ExecutionNode) => void;
}

export default function ExecutionGraph({ sessionId, onNodeClick }: ExecutionGraphProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [loading, setLoading] = useState(true);

  const fetchGraph = useCallback(async () => {
    try {
      const graph = await getContextGraph(sessionId);
      const flowNodes: Node[] = graph.nodes.map((node, idx) => ({
        id: node.id,
        position: {
          x: (idx % 6) * 250,
          y: Math.floor(idx / 6) * 120,
        },
        data: {
          label: (
            <div className="text-left">
              <div className="font-mono text-xs font-bold truncate max-w-[180px]">
                {node.command.binary} {node.command.args.slice(0, 2).join(' ')}
              </div>
              <div className="text-[10px] opacity-60 mt-0.5">
                {new Date(node.timestamp).toLocaleTimeString()} | {node.duration_ms}ms
              </div>
              {node.exit_code !== 0 && (
                <div className="text-[10px] text-red-400 mt-0.5">exit: {node.exit_code}</div>
              )}
            </div>
          ),
          execution: node,
        },
        style: {
          background: '#111111',
          border: `2px solid ${node.exit_code !== 0 ? '#ef4444' : categoryColors[node.command.category] || '#666666'}`,
          borderRadius: '8px',
          padding: '8px 12px',
          color: '#e8e8e8',
          fontSize: '12px',
          width: 200,
        },
      }));

      const flowEdges: FlowEdge[] = graph.edges.map((edge, idx) => ({
        id: `e-${idx}`,
        source: edge.from,
        target: edge.to,
        label: edge.dependency_type.replace('_', ' '),
        style: { stroke: '#333333' },
        labelStyle: { fill: '#666666', fontSize: 10 },
        markerEnd: { type: MarkerType.ArrowClosed, color: '#333333' },
        animated: edge.dependency_type === 'file_read_write',
      }));

      setNodes(flowNodes);
      setEdges(flowEdges);
    } catch (e) {
      console.error('Failed to fetch graph:', e);
    } finally {
      setLoading(false);
    }
  }, [sessionId]);

  useEffect(() => {
    fetchGraph();
    const interval = setInterval(fetchGraph, 10000);
    return () => clearInterval(interval);
  }, [fetchGraph]);

  const handleNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    if (node.data.execution && onNodeClick) {
      onNodeClick(node.data.execution);
    }
  }, [onNodeClick]);

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center text-nucleus-muted">
        Loading execution graph...
      </div>
    );
  }

  return (
    <div className="h-full w-full">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        fitView
        className="bg-nucleus-bg"
      >
        <Background color="#1a1a1a" gap={20} />
        <Controls
          style={{ background: '#111111', border: '1px solid #1a1a1a', borderRadius: '8px' }}
        />
        <MiniMap
          style={{ background: '#0a0a0a', border: '1px solid #1a1a1a' }}
          nodeColor={(n) => {
            const exec = n.data?.execution as ExecutionNode;
            if (!exec) return '#666666';
            if (exec.exit_code !== 0) return '#ef4444';
            return categoryColors[exec.command.category] || '#666666';
          }}
        />
      </ReactFlow>
    </div>
  );
}
