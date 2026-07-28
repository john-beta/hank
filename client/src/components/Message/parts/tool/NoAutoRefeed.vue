<script setup lang="ts">
import type { ToolPart } from '../../../../types'
import type { VNode } from 'vue'
import { ArrowRightIcon } from '@lucide/vue'
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

interface WorkspaceEntry {
  current_path: string
  proposed_path: string
}

// Args are always { "proposed_workspace_entries": [{ "current_path": "...", "proposed_path": "..." }] }.
const entries = computed<WorkspaceEntry[]>(() => {
  if (!props.part.args) return []
  try {
    const parsed = JSON.parse(props.part.args) as {
      proposed_workspace_entries?: WorkspaceEntry[]
    }
    return parsed.proposed_workspace_entries ?? []
  } catch {
    return []
  }
})

// Build a nested folder/file tree from the flat paths, de-duplicating shared
// folder prefixes (e.g. two entries under "cmd/" share one folder node).
function buildTree(paths: string[]): TreeNode[] {
  const root: FolderNode = { type: 'folder', name: '', path: '', children: [] }
  for (const fullPath of paths) {
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
}

// Expand every folder by default so proposed entries are visible immediately.
function collectFolderPaths(nodes: TreeNode[]): Set<string> {
  const set = new Set<string>()
  const walk = (items: TreeNode[]) => {
    for (const node of items) {
      if (node.type === 'folder') {
        set.add(node.path)
        walk(node.children)
      }
    }
  }
  walk(nodes)
  return set
}

const leftTree = computed(() => buildTree(entries.value.map((e) => e.current_path)))
const rightTree = computed(() => buildTree(entries.value.map((e) => e.proposed_path)))
const leftExpanded = computed(() => collectFolderPaths(leftTree.value))
const rightExpanded = computed(() => collectFolderPaths(rightTree.value))

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
    <div v-if="entries.length" class="flex items-center gap-3">
      <FileTree :default-expanded="leftExpanded" class="flex-1">
        <FileTreeNode v-for="node in leftTree" :key="node.path" :node="node" />
      </FileTree>
      <ArrowRightIcon class="size-4 shrink-0 text-muted-foreground" />
      <FileTree :default-expanded="rightExpanded" class="flex-1">
        <FileTreeNode v-for="node in rightTree" :key="node.path" :node="node" />
      </FileTree>
    </div>
    <HITL :part="part" />
  </div>
</template>
