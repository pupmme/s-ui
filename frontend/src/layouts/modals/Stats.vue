<template>
  <v-dialog transition="dialog-bottom-transition" width="500">
    <v-card class="rounded-lg">
      <v-card-title>{{ $t('stats.title') || 'Stats' }} - {{ tag }}</v-card-title>
      <v-divider></v-divider>
      <v-card-text>
        <div class="text-center py-4">
          <v-icon color="grey">mdi-chart-bar</v-icon>
          <div class="text-grey mt-2">Resource: {{ resource }}</div>
          <div class="text-grey mt-1">Tag: {{ tag }}</div>
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
export default {
  props: ['visible', 'resource', 'tag'],
  emits: ['close'],
}
</script>
