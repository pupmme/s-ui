<template>
  <v-dialog transition="dialog-bottom-transition" width="500">
    <v-card class="rounded-lg">
      <v-card-title>{{ $t('admin.changes') }}</v-card-title>
      <v-divider></v-divider>
      <v-card-text>
        <v-list v-if="changes.length > 0">
          <v-list-item v-for="(change, i) in changes" :key="i">
            <v-list-item-title>{{ change }}</v-list-item-title>
          </v-list-item>
        </v-list>
        <div v-else class="text-center py-4">
          <v-icon color="grey">mdi-history</v-icon>
          <div class="text-grey mt-2">No changes recorded</div>
        </div>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" variant="outlined" @click="$emit('close')">{{ $t('actions.close') }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { ref } from 'vue'

export default {
  props: ['visible', 'actor', 'admins'],
  emits: ['close'],
  setup(props: any) {
    const changes = ref<string[]>([])
    return { changes }
  }
}
</script>
