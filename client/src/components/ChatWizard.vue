<script setup lang="ts">
import { ref } from 'vue'
import type { Ref } from 'vue'

const emit = defineEmits<{
  'create-session': [chosenRootDir: string]
}>()

const chosenRootDir: Ref<string | null> = ref(null)
const submitting: Ref<boolean> = ref(false)

function handleSubmit(): void {
  if (!chosenRootDir.value) return
  emit('create-session', chosenRootDir.value)
}
</script>

<template>
  <div class="wizard">
    <div class="wizard-card">
      <h1>Start a new session</h1>
      <p>Choose a file to begin.</p>
      <form class="wizard-form" @submit.prevent="handleSubmit">
        <div class="wizard-field">
          <input type="text" v-model="chosenRootDir" />
        </div>
        <button type="submit" :disabled="!chosenRootDir || submitting">Create session</button>
      </form>
    </div>
  </div>
</template>
