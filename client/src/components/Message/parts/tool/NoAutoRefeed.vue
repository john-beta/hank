<script setup lang="ts">
import type { ToolPart } from '../../../../types'
import type { VNode } from 'vue'
import { computed, h } from 'vue'
import { FileTree, FileTreeFile, FileTreeFolder } from '../../../ai-elements/file-tree'
import HITL from './HITL.vue'

const props = defineProps<{
  part: ToolPart
}>()

interface FileNode {
  type: 'file'
  name: string
  path: string
}
interface FolderNode {
  type: 'folder'
  name: string
  path: string
  children: TreeNode[]
}
type TreeNode = FileNode | FolderNode

// Args are always { "proposed_workspace_entries": [{ "path": "..." }] }.
const paths = computed<string[]>(() => {
  if (!props.part.args) return []
  try {
    const parsed = JSON.parse(props.part.args) as {
      proposed_workspace_entries?: { path: string }[]
    }
    return (parsed.proposed_workspace_entries ?? []).map((e) => e.path)
  } catch {
    return []
  }
})

// Build a nested folder/file tree from the flat paths, de-duplicating shared
// folder prefixes (e.g. two entries under "cmd/" share one folder node).
const tree = computed<TreeNode[]>(() => {
  const root: FolderNode = { type: 'folder', name: '', path: '', children: [] }
  for (const fullPath of paths.value) {
    const segments = fullPath.split('/').filter(Boolean)
    let cursor = root
    segments.forEach((segment, i) => {
      const path = segments.slice(0, i + 1).join('/')
      if (i === segments.length - 1) {
        cursor.children.push({ type: 'file', name: segment, path })
        return
      }
      let next = cursor.children.find(
        (c): c is FolderNode => c.type === 'folder' && c.name === segment,
      )
      if (!next) {
        next = { type: 'folder', name: segment, path, children: [] }
        cursor.children.push(next)
      }
      cursor = next
    })
  }
  return root.children
})

// Expand every folder by default so proposed entries are visible immediately.
const allFolderPaths = computed(() => {
  const set = new Set<string>()
  const walk = (nodes: TreeNode[]) => {
    for (const node of nodes) {
      if (node.type === 'folder') {
        set.add(node.path)
        walk(node.children)
      }
    }
  }
  walk(tree.value)
  return set
})

function FileTreeNode(nodeProps: { node: TreeNode }): VNode {
  if (nodeProps.node.type === 'file') {
    return h(FileTreeFile, { path: nodeProps.node.path, name: nodeProps.node.name })
  }
  return h(FileTreeFolder, { path: nodeProps.node.path, name: nodeProps.node.name }, () =>
    nodeProps.node.type === 'folder'
      ? nodeProps.node.children.map((child) => h(FileTreeNode, { node: child, key: child.path }))
      : [],
  )
}
</script>

<template>
  <div class="tool-part">
    <FileTree v-if="tree.length" :default-expanded="allFolderPaths">
      <FileTreeNode v-for="node in tree" :key="node.path" :node="node" />
    </FileTree>
    <HITL :part="part" />
  </div>
</template>
