import type { DagNode, DagEdge } from '@/types'

interface LayoutNode {
  node: DagNode
  x: number
  y: number
}

interface LayoutOptions {
  nodeWidth: number
  nodeHeight: number
  horizontalGap: number
  verticalGap: number
}

interface LayoutResult {
  nodes: LayoutNode[]
  edges: DagEdge[]
  width: number
  height: number
}

/**
 * 计算 DAG 布局
 * 使用层次布局算法
 */
export function computeDagLayout(
  nodes: DagNode[],
  edges: DagEdge[],
  options: LayoutOptions
): LayoutResult {
  const { nodeWidth, nodeHeight, horizontalGap, verticalGap } = options

  if (nodes.length === 0) {
    return { nodes: [], edges: [], width: 0, height: 0 }
  }

  // 构建邻接表
  const adjacency = new Map<string, string[]>()
  const inDegree = new Map<string, number>()

  nodes.forEach((node) => {
    adjacency.set(node.id, [])
    inDegree.set(node.id, 0)
  })

  edges.forEach((edge) => {
    const neighbors = adjacency.get(edge.source)
    if (neighbors) {
      neighbors.push(edge.target)
    }
    inDegree.set(edge.target, (inDegree.get(edge.target) || 0) + 1)
  })

  // 计算层次（拓扑排序 + BFS）
  const levels = new Map<string, number>()
  const queue: string[] = []

  // 找出所有入度为0的节点作为起点
  nodes.forEach((node) => {
    if (inDegree.get(node.id) === 0) {
      queue.push(node.id)
      levels.set(node.id, 0)
    }
  })

  // BFS 计算层次
  while (queue.length > 0) {
    const current = queue.shift()!
    const currentLevel = levels.get(current)!

    const neighbors = adjacency.get(current) || []
    for (const neighbor of neighbors) {
      const newLevel = currentLevel + 1
      if (!levels.has(neighbor) || levels.get(neighbor)! < newLevel) {
        levels.set(neighbor, newLevel)
      }
      queue.push(neighbor)
    }
  }

  // 处理未连接的节点（应该在图外）
  nodes.forEach((node) => {
    if (!levels.has(node.id)) {
      levels.set(node.id, 0)
    }
  })

  // 按层次分组节点
  const levelGroups = new Map<number, string[]>()
  levels.forEach((level, nodeId) => {
    if (!levelGroups.has(level)) {
      levelGroups.set(level, [])
    }
    levelGroups.get(level)!.push(nodeId)
  })

  // 计算节点位置
  const layoutNodes: LayoutNode[] = []
  const maxLevel = Math.max(...Array.from(levels.values()), 0)
  const totalHeight = (maxLevel + 1) * (nodeHeight + verticalGap)

  levelGroups.forEach((nodeIds, level) => {
    const levelWidth = nodeIds.length * (nodeWidth + horizontalGap) - horizontalGap
    const startX = (levelWidth > 0 ? (800 - levelWidth) / 2 : 0)

    nodeIds.forEach((nodeId, index) => {
      const node = nodes.find((n) => n.id === nodeId)!
      layoutNodes.push({
        node,
        x: startX + index * (nodeWidth + horizontalGap),
        y: level * (nodeHeight + verticalGap),
      })
    })
  })

  // 计算总宽度
  let maxX = 0
  layoutNodes.forEach((item) => {
    maxX = Math.max(maxX, item.x + nodeWidth)
  })

  return {
    nodes: layoutNodes,
    edges,
    width: maxX + horizontalGap,
    height: totalHeight + verticalGap,
  }
}

/**
 * 查找从起点到终点的所有路径
 */
export function findAllPaths(
  nodes: DagNode[],
  edges: DagEdge[],
  startId: string,
  endId: string
): string[][] {
  const adjacency = new Map<string, string[]>()

  nodes.forEach((node) => {
    adjacency.set(node.id, [])
  })

  edges.forEach((edge) => {
    const neighbors = adjacency.get(edge.source)
    if (neighbors) {
      neighbors.push(edge.target)
    }
  })

  const paths: string[][] = []
  const currentPath: string[] = []

  const dfs = (current: string) => {
    currentPath.push(current)

    if (current === endId) {
      paths.push([...currentPath])
    } else {
      const neighbors = adjacency.get(current) || []
      for (const neighbor of neighbors) {
        dfs(neighbor)
      }
    }

    currentPath.pop()
  }

  dfs(startId)
  return paths
}

/**
 * 检查是否存在循环依赖
 */
export function hasCycle(nodes: DagNode[], edges: DagEdge[]): boolean {
  const visited = new Set<string>()
  const recStack = new Set<string>()

  const adjacency = new Map<string, string[]>()
  nodes.forEach((node) => {
    adjacency.set(node.id, [])
  })
  edges.forEach((edge) => {
    adjacency.get(edge.source)?.push(edge.target)
  })

  const dfs = (nodeId: string): boolean => {
    visited.add(nodeId)
    recStack.add(nodeId)

    const neighbors = adjacency.get(nodeId) || []
    for (const neighbor of neighbors) {
      if (!visited.has(neighbor)) {
        if (dfs(neighbor)) return true
      } else if (recStack.has(neighbor)) {
        return true
      }
    }

    recStack.delete(nodeId)
    return false
  }

  for (const node of nodes) {
    if (!visited.has(node.id)) {
      if (dfs(node.id)) return true
    }
  }

  return false
}
