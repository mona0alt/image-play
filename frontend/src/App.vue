<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { ensureMockSession } from './services/session'

onLaunch(() => {
  void ensureMockSession()
})
</script>

<template>
  <view class="content">
    <text class="title">Image Play</text>
  </view>
</template>

<style>
.content {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}
.title {
  font-size: 32px;
  font-weight: bold;
}
</style>
