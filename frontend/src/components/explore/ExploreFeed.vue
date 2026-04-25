<script setup lang="ts">
import { ref } from 'vue'
import ExploreCard from './ExploreCard.vue'
import type { ExploreItem } from '../../services/api'

const props = defineProps<{
  items: ExploreItem[]
  hasMore: boolean
}>()

const emit = defineEmits<{
  loadMore: []
  like: [id: number]
  sameStyle: [item: ExploreItem]
}>()

const currentIndex = ref(0)

function onChange(e: { detail: { current: number } }) {
  currentIndex.value = e.detail.current

  // Trigger pagination when approaching end
  if (props.items.length > 0 && e.detail.current >= props.items.length - 3) {
    if (props.hasMore) {
      emit('loadMore')
    }
  }
}

function handleLike(id: number) {
  emit('like', id)
}

function handleSameStyle(item: ExploreItem) {
  emit('sameStyle', item)
}
</script>

<template>
  <swiper
    class="feed"
    vertical
    :current="currentIndex"
    @change="onChange"
  >
    <swiper-item
      v-for="item in items"
      :key="item.id"
    >
      <ExploreCard
        :item="item"
        @like="handleLike"
        @same-style="handleSameStyle"
      />
    </swiper-item>
  </swiper>
</template>

<style scoped>
.feed {
  position: absolute;
  inset: 0;
}
</style>
