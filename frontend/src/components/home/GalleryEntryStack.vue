<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { HomeEntryCard } from '../../pages/home/entry-cards'
import {
  clampStackIndex,
  getVisibleHomeEntries,
  resolveStackSwipe,
} from '../../pages/home/entry-stack'

const props = defineProps<{
  entries: HomeEntryCard[]
}>()

const emit = defineEmits<{
  (e: 'open', entry: HomeEntryCard): void
}>()

const activeIndex = ref(0)
const touchStartY = ref<number | null>(null)

watch(
  () => props.entries.length,
  (length) => {
    activeIndex.value = clampStackIndex(length, activeIndex.value)
  },
  { immediate: true },
)

const visibleCards = computed(() =>
  getVisibleHomeEntries(props.entries, activeIndex.value),
)

function readTouchY(event: any): number | null {
  return event?.changedTouches?.[0]?.pageY
    ?? event?.touches?.[0]?.pageY
    ?? null
}

function handleTouchStart(event: any) {
  touchStartY.value = readTouchY(event)
}

function handleTouchEnd(event: any) {
  const endY = readTouchY(event)
  if (touchStartY.value == null || endY == null) {
    touchStartY.value = null
    return
  }

  const { nextIndex } = resolveStackSwipe({
    count: props.entries.length,
    activeIndex: activeIndex.value,
    deltaY: endY - touchStartY.value,
  })

  activeIndex.value = nextIndex
  touchStartY.value = null
}

function handleTouchCancel() {
  touchStartY.value = null
}

function handleCardTap(index: number, entry: HomeEntryCard) {
  if (index === activeIndex.value) {
    emit('open', entry)
    return
  }

  activeIndex.value = clampStackIndex(props.entries.length, index)
}
</script>

<template>
  <view
    class="entry-stack"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    @touchcancel="handleTouchCancel"
  >
    <view
      v-for="card in visibleCards"
      :key="card.entry.key"
      class="entry-stack__card"
      :class="[
        `entry-stack__card--${card.slot}`,
        `entry-stack__card--${card.entry.accent}`,
      ]"
      @tap="handleCardTap(card.index, card.entry)"
    >
      <view class="entry-stack__surface">
        <template v-if="card.slot === 'active'">
          <view class="entry-stack__active-top">
            <view>
              <text class="entry-stack__eyebrow">{{ card.entry.eyebrow }}</text>
              <text class="entry-stack__title">{{ card.entry.title }}</text>
            </view>
            <view class="entry-stack__seal" />
          </view>

          <text class="entry-stack__description">{{ card.entry.description }}</text>

          <view class="entry-stack__footer">
            <view class="entry-stack__tags">
              <text
                v-for="tag in card.entry.tags.slice(0, 3)"
                :key="tag"
                class="entry-stack__tag"
              >
                {{ tag }}
              </text>
            </view>
            <text class="entry-stack__signature">Curated Edition</text>
          </view>
        </template>

        <template v-else>
          <view class="entry-stack__peek">
            <view>
              <text class="entry-stack__peek-eyebrow">{{ card.entry.eyebrow }}</text>
              <text class="entry-stack__peek-title">{{ card.entry.title }}</text>
            </view>
            <view class="entry-stack__peek-mark" />
          </view>
        </template>
      </view>
    </view>
  </view>
</template>

<style scoped>
.entry-stack {
  position: relative;
  height: 620rpx;
}

.entry-stack__card {
  position: absolute;
  left: 0;
  right: 0;
  border-radius: 36rpx;
  overflow: hidden;
  transition: top 180ms ease, transform 180ms ease, box-shadow 180ms ease;
}

.entry-stack__card--active {
  top: 0;
  z-index: 4;
}

.entry-stack__card--peek-1 {
  top: 404rpx;
  z-index: 3;
}

.entry-stack__card--peek-2 {
  top: 472rpx;
  z-index: 2;
}

.entry-stack__card--peek-3 {
  top: 540rpx;
  z-index: 1;
}

.entry-stack__surface {
  position: relative;
  min-height: 96rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.08);
  box-shadow:
    0 24rpx 44rpx rgba(34, 25, 18, 0.16),
    0 8rpx 18rpx rgba(34, 25, 18, 0.08),
    inset 0 1rpx 0 rgba(255, 255, 255, 0.18);
}

.entry-stack__surface::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(125deg, rgba(255, 255, 255, 0.24) 8%, rgba(255, 255, 255, 0.04) 28%, rgba(255, 255, 255, 0) 52%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.entry-stack__surface::after {
  content: '';
  position: absolute;
  inset: 1rpx;
  border-radius: 36rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.1);
  pointer-events: none;
}

.entry-stack__card--active .entry-stack__surface {
  min-height: 388rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.entry-stack__active-top,
.entry-stack__peek,
.entry-stack__footer,
.entry-stack__tags {
  display: flex;
  align-items: center;
}

.entry-stack__active-top,
.entry-stack__peek,
.entry-stack__footer {
  justify-content: space-between;
}

.entry-stack__tags {
  gap: 10rpx;
  flex-wrap: wrap;
}

.entry-stack__eyebrow,
.entry-stack__peek-eyebrow {
  display: block;
  font-size: 18rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  opacity: 0.72;
}

.entry-stack__title {
  display: block;
  margin-top: 10rpx;
  font-size: 52rpx;
  line-height: 1.08;
  font-weight: 700;
}

.entry-stack__description {
  max-width: 520rpx;
  font-size: 26rpx;
  line-height: 1.6;
  opacity: 0.86;
}

.entry-stack__peek-title {
  display: block;
  margin-top: 6rpx;
  font-size: 34rpx;
  font-weight: 700;
}

.entry-stack__tag {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  font-size: 18rpx;
  background: rgba(255, 255, 255, 0.12);
  border: 1rpx solid rgba(255, 255, 255, 0.08);
}

.entry-stack__signature {
  font-size: 18rpx;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  opacity: 0.54;
}

.entry-stack__seal,
.entry-stack__peek-mark {
  flex-shrink: 0;
  border-radius: 50%;
  border: 1rpx solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
}

.entry-stack__seal {
  width: 72rpx;
  height: 72rpx;
}

.entry-stack__peek-mark {
  width: 56rpx;
  height: 56rpx;
}

.entry-stack__card--portrait .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 82% 20%, rgba(235, 212, 185, 0.2), transparent 18%),
    linear-gradient(160deg, #120f0d 0%, #322821 32%, #6b5a4e 100%);
}

.entry-stack__card--analysis .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 85% 18%, rgba(237, 219, 194, 0.16), transparent 18%),
    linear-gradient(160deg, #181310 0%, #2c231e 34%, #796656 100%);
}

.entry-stack__card--festival .entry-stack__surface {
  color: #fbf5ef;
  background:
    radial-gradient(circle at 84% 20%, rgba(255, 227, 196, 0.16), transparent 18%),
    linear-gradient(160deg, #40241a 0%, #7c4b33 34%, #c99668 100%);
}

.entry-stack__card--invitation .entry-stack__surface {
  color: #221b17;
  background:
    radial-gradient(circle at 84% 22%, rgba(255, 255, 255, 0.5), transparent 16%),
    linear-gradient(160deg, #d8cabd 0%, #e9ddd1 52%, #f7f0e8 100%);
}

.entry-stack__card--tshirt .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 84% 18%, rgba(206, 217, 229, 0.14), transparent 18%),
    linear-gradient(160deg, #111214 0%, #25303b 36%, #53616f 100%);
}

.entry-stack__card--poster .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 84% 18%, rgba(226, 214, 203, 0.14), transparent 18%),
    linear-gradient(160deg, #1d1b1a 0%, #433d38 36%, #7c7167 100%);
}

.entry-stack__card--invitation .entry-stack__tag,
.entry-stack__card--invitation .entry-stack__seal,
.entry-stack__card--invitation .entry-stack__peek-mark {
  border-color: rgba(50, 38, 28, 0.12);
  background: rgba(50, 38, 28, 0.06);
}
</style>
