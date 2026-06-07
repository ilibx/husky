<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const url = computed(() => String(route.query.url || ''))
</script>

<template>
  <iframe v-if="url" :src="url" class="external-frame" />
</template>

<style scoped>
.external-frame {
  width: 100%;
  height: calc(100vh - 96px);
  border: 0;
  background: #fff;
}
</style>
