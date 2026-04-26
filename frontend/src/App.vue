<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { ensureSession } from './services/session'

onLaunch(() => {
  void ensureSession()
})
</script>

<style>
page {
  --gallery-bg: #fdf8f8;
  --gallery-surface: #ffffff;
  --gallery-surface-soft: #f4efed;
  --gallery-border: rgba(28, 27, 27, 0.08);
  --gallery-text: #1c1b1b;
  --gallery-muted: #6d6865;
  --gallery-accent: #111111;
  background: var(--gallery-bg);
  color: var(--gallery-text);
}

view,
text,
button,
image,
scroll-view,
input,
textarea {
  box-sizing: border-box;
}

button {
  border: none;
}

button::after {
  border: none;
}
</style>
