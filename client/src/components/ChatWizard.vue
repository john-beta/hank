<script setup lang="ts">
import { ref } from 'vue'
import type { Ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useChatSession } from '../composables/useChatSession'

const chatSession = useChatSession()

const chosenRootDir: Ref<string> = ref('')
const submitting: Ref<boolean> = ref(false)
const error: Ref<string | null> = ref(null)

async function handleSubmit(): Promise<void> {
  if (!chosenRootDir.value || submitting.value) return
  submitting.value = true
  error.value = null
  try {
    await chatSession.create(chosenRootDir.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create session'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="flex h-full flex-1 items-center justify-center p-4">
    <div class="w-full max-w-sm rounded-lg border bg-card p-6 shadow-xs">
      <h1 class="text-lg font-semibold text-foreground">Start a new session</h1>
      <p class="mt-1 text-sm text-muted-foreground">Choose a file to begin.</p>
      <form class="mt-4 flex flex-col gap-3" @submit.prevent="handleSubmit">
        <Input v-model="chosenRootDir" type="text" placeholder="Path to file" />
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <Button type="submit" :disabled="!chosenRootDir || submitting">Create session</Button>
      </form>
    </div>
  </div>
</template>
