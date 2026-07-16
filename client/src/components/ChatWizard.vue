<script setup lang="ts">
import { ref } from 'vue'
import type { Ref } from 'vue'

const emit = defineEmits<{
  'create-session': [chosenFilePath: string]
}>()

const chosenFile: Ref<File | null> = ref(null)
const submitting: Ref<boolean> = ref(false)

function handleFileChange(event: Event): void {
  const input = event.target as HTMLInputElement
  chosenFile.value = input.files?.[0] ?? null
}

function handleSubmit(): void {
  if (!chosenFile.value) return
  emit('create-session', chosenFile.value.name)
}
</script>

<template>
  <div class="wizard">
    <div class="wizard-card">
      <h1>Start a new session</h1>
      <p>Choose a file to begin.</p>
      <form class="wizard-form" @submit.prevent="handleSubmit">
        <div class="wizard-field">
          <input type="file" @change="handleFileChange" />
        </div>
        <button type="submit" :disabled="!chosenFile || submitting">Create session</button>
      </form>
    </div>
  </div>
</template>
