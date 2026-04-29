<script setup lang="ts">
import { computed } from 'vue'
import { parseMarkdown } from '../../utils/markdown'

const props = defineProps<{
  source: string
  dark?: boolean
}>()

const nodes = computed(() => parseMarkdown(props.source))
</script>

<template>
  <view class="markdown-renderer" :class="{ 'markdown-renderer--dark': dark }">
    <block v-for="(node, i) in nodes" :key="i">
      <view v-if="node.type === 'h2'" class="md-h2">
        <text
          v-for="(span, j) in node.spans"
          :key="j"
          :class="{ 'md-bold': span.bold }"
        >
          {{ span.text }}
        </text>
      </view>
      <view v-else-if="node.type === 'h3'" class="md-h3">
        <text
          v-for="(span, j) in node.spans"
          :key="j"
          :class="{ 'md-bold': span.bold }"
        >
          {{ span.text }}
        </text>
      </view>
      <view v-else-if="node.type === 'p'" class="md-p">
        <text
          v-for="(span, j) in node.spans"
          :key="j"
          :class="{ 'md-bold': span.bold }"
        >
          {{ span.text }}
        </text>
      </view>
      <view v-else-if="node.type === 'ul'" class="md-ul">
        <view v-for="(item, k) in node.items" :key="k" class="md-li">
          <text class="md-li-marker">&#x25CF;</text>
          <text
            v-for="(span, m) in item"
            :key="m"
            class="md-li-text"
            :class="{ 'md-bold': span.bold }"
          >
            {{ span.text }}
          </text>
        </view>
      </view>
      <view v-else-if="node.type === 'hr'" class="md-hr" />
    </block>
  </view>
</template>

<style scoped>
.markdown-renderer {
  display: flex;
  flex-direction: column;
}

.markdown-renderer > view:last-child {
  margin-bottom: 0;
}

.md-h2 {
  margin-top: 36rpx;
  margin-bottom: 16rpx;
  padding-bottom: 12rpx;
  border-bottom: 1rpx solid var(--gallery-border);
}

.md-h2 text {
  font-size: 32rpx;
  font-weight: 600;
  line-height: 1.4;
  color: var(--gallery-text);
}

.md-h3 {
  margin-top: 28rpx;
  margin-bottom: 12rpx;
}

.md-h3 text {
  font-size: 28rpx;
  font-weight: 600;
  line-height: 1.4;
  color: var(--gallery-text);
}

.md-p {
  margin-bottom: 16rpx;
  line-height: 1.8;
}

.md-p text {
  font-size: 28rpx;
  color: var(--gallery-text);
}

.md-bold {
  font-weight: 600;
}

.md-ul {
  margin-bottom: 16rpx;
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.md-li {
  display: flex;
  align-items: flex-start;
  gap: 14rpx;
  line-height: 1.7;
}

.md-li-marker {
  font-size: 16rpx;
  color: var(--gallery-muted);
  margin-top: 10rpx;
  flex-shrink: 0;
}

.md-li-text {
  font-size: 28rpx;
  color: var(--gallery-text);
}

.md-hr {
  height: 1rpx;
  background: var(--gallery-border);
  margin: 24rpx 0;
}

/* Dark theme overrides */
.markdown-renderer--dark .md-h2 {
  border-bottom-color: rgba(0, 240, 255, 0.2);
}

.markdown-renderer--dark .md-h2 text,
.markdown-renderer--dark .md-h3 text,
.markdown-renderer--dark .md-p text,
.markdown-renderer--dark .md-li-text {
  color: #e0e0f0;
}

.markdown-renderer--dark .md-li-marker {
  color: #00f0ff;
}

.markdown-renderer--dark .md-hr {
  background: rgba(0, 240, 255, 0.12);
}

.markdown-renderer--dark .md-p .md-bold {
  color: #00f0ff;
}
</style>
