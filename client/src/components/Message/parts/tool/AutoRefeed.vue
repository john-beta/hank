<script setup lang="ts">
import type { ToolPart } from '../../../../types'
import { computed } from 'vue'
import { CodeBlock } from '../../../ai-elements/code-block'
import { Tool, ToolContent, ToolHeader } from '../../../ai-elements/tool'

const props = defineProps<{
  part: ToolPart
}>()

// Params are always { "script": "<python source>" } — show the script as Python.
const script = computed(() => {
  if (!props.part.args) return ''
  try {
    const parsed = JSON.parse(props.part.args)
    return typeof parsed?.script === 'string' ? parsed.script : props.part.args
  } catch {
    return props.part.args
  }
})

// Result is either { success: true, output } (output is JSON) or
// { success: false, error } (plain text).
const result = computed(() => {
  if (props.part.result == null) return null
  try {
    return JSON.parse(props.part.result) as { success: boolean; output?: string; error?: string }
  } catch {
    return null
  }
})

const isError = computed(() => result.value?.success === false)

// Successful output is JSON — pretty-print when possible.
const outputCode = computed(() => {
  const out = result.value?.output ?? ''
  try {
    return JSON.stringify(JSON.parse(out), null, 2)
  } catch {
    return out
  }
})

// Map our model onto the AI Elements state vocabulary, reflecting success/error.
const state = computed<'input-available' | 'output-available' | 'output-error'>(() => {
  if (props.part.state !== 'result') return 'input-available'
  return isError.value ? 'output-error' : 'output-available'
})
</script>

<template>
  <Tool :default-open="false">
    <ToolHeader :type="`tool-${part.name ?? 'unknown'}`" :state="state" :title="part.name" />
    <ToolContent>
      <div class="space-y-2 p-4">
        <h4 class="font-medium text-muted-foreground text-xs uppercase tracking-wide">
          Parameters
        </h4>
        <CodeBlock :code="script" language="python" />
      </div>
      <div v-if="part.state === 'result'" class="space-y-2 p-4">
        <h4 class="font-medium text-muted-foreground text-xs uppercase tracking-wide">
          {{ isError ? 'Error' : 'Result' }}
        </h4>
        <CodeBlock v-if="!isError" :code="outputCode" language="json" />
        <pre
          v-else
          class="tool-part-result rounded-md border border-red-300 px-4 py-3 text-red-800 whitespace-pre-wrap"
        >
  {{ result?.error }}
</pre
        >
      </div>
    </ToolContent>
  </Tool>
</template>
