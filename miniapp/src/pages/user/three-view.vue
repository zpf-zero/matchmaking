<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { userApi } from '@/api'
import type { ThreeViewQuestion } from '@/api/types'
import { useUserStore } from '@/stores/user'

const store = useUserStore()
const questions = ref<ThreeViewQuestion[]>([])
/** 与 questions 同序，-1 表示未答 */
const answers = ref<number[]>([])
const loading = ref(true)
const saving = ref(false)

/**
 * 三观问答。答题免费、看答案才需权益。
 *
 * 为什么答题免费（见 docs/08-权益体系重构方案.md 4.2②）：
 * 题库的全部价值在于「让所有人都填」—— 它不给平台增加任何数据采集成本，
 * 却能在无成本的前提下提供差异化。若答题也要权益，就没人填了，
 * 权益内容随之变成空的。
 */
onLoad(async () => {
  try {
    const res = await userApi.myThreeView()
    questions.value = res.questions || []
    answers.value = res.answers || []
  } catch {
    // request 层已提示
  } finally {
    loading.value = false
  }
})

const answered = computed(() => answers.value.filter((a) => a >= 0).length)
const allDone = computed(
  () => questions.value.length > 0 && answered.value === questions.value.length
)

/** 已答题数用于顶部进度条，让「还差几题」一目了然 */
const progress = computed(() =>
  questions.value.length ? Math.round((answered.value / questions.value.length) * 100) : 0
)

function choose(qi: number, oi: number) {
  // 再点一次取消选择：允许用户「手滑后撤」，而不是被迫二选一
  answers.value[qi] = answers.value[qi] === oi ? -1 : oi
}

async function save() {
  if (saving.value) return
  if (answered.value === 0) {
    uni.showToast({ title: '至少回答一道题', icon: 'none' })
    return
  }
  saving.value = true
  try {
    await userApi.saveThreeView(answers.value)
    uni.showToast({ title: '已保存', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 700)
  } catch {
    // request 层已提示
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <view class="page">
    <!-- 顶部说明：答题免费，看答案需权益。先把规则讲清，避免用户答完才发现 -->
    <view class="hero">
      <view class="hero-title">三观问答</view>
      <view class="hero-desc">
        回答越完整，越容易遇到观念相合的人。答题完全免费，邀请好友通过审核后，可查看彼此的选择与一致率。
      </view>
      <view v-if="questions.length" class="progress">
        <view class="progress-bar">
          <view class="progress-fill" :style="{ width: progress + '%' }"></view>
        </view>
        <text class="progress-text">已答 {{ answered }}/{{ questions.length }}</text>
      </view>
    </view>

    <view v-if="loading" class="empty">正在加载…</view>

    <template v-else>
      <view v-for="(q, qi) in questions" :key="q.id" class="card">
        <view class="q-title">
          <text class="q-no">{{ qi + 1 }}</text>
          <text class="q-text">{{ q.title }}</text>
        </view>
        <view class="options">
          <view
            v-for="(opt, oi) in q.options"
            :key="oi"
            class="option"
            :class="{ on: answers[qi] === oi }"
            @click="choose(qi, oi)"
          >
            <view class="dot"></view>
            <text class="opt-text">{{ opt }}</text>
          </view>
        </view>
      </view>

      <view class="foot-tip">
        没有标准答案，选更接近自己真实想法的即可。可以只答一部分，之后再来补。
      </view>

      <view class="actions">
        <view class="save-btn" :class="{ done: allDone }" @click="save">
          {{ saving ? '保存中…' : allDone ? '完成' : '保存' }}
        </view>
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped>
.page {
  padding: 28rpx 24rpx calc(180rpx + env(safe-area-inset-bottom));
}

.hero {
  padding: 8rpx 8rpx 26rpx;
}

.hero-title {
  font-size: $font-display;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -1rpx;
}

.hero-desc {
  margin-top: 14rpx;
  font-size: $font-mid;
  color: $text-weak;
  line-height: 1.7;
}

.progress {
  margin-top: 24rpx;
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.progress-bar {
  flex: 1;
  height: 10rpx;
  border-radius: 6rpx;
  background: $bg-subtle;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 6rpx;
  background: $brand;
  transition: width 220ms ease;
}

.progress-text {
  font-size: $font-micro;
  color: $text-placeholder;
  font-variant-numeric: tabular-nums;
}

.card {
  background: $bg-card;
  border: 1rpx solid $border;
  border-radius: $radius-lg;
  padding: 28rpx;
  margin-bottom: 22rpx;
}

.q-title {
  display: flex;
  align-items: flex-start;
  gap: 14rpx;
  margin-bottom: 22rpx;
}

.q-no {
  flex-shrink: 0;
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  border-radius: 50%;
  background: $brand-light;
  color: $brand;
  font-size: $font-micro;
  font-weight: 600;
}

.q-text {
  flex: 1;
  font-size: $font-body;
  font-weight: 550;
  color: $text-main;
  line-height: 1.55;
}

.options {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}

.option {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid $border;
  border-radius: $radius;
  background: $bg-raised;
  transition: border-color 160ms ease, background-color 160ms ease;

  &.on {
    border-color: $brand;
    background: $brand-light;

    .dot {
      border-color: $brand;
      background: $brand;
      box-shadow: inset 0 0 0 5rpx $bg-raised;
    }

    .opt-text {
      color: $brand;
      font-weight: 550;
    }
  }
}

.dot {
  flex-shrink: 0;
  width: 30rpx;
  height: 30rpx;
  border-radius: 50%;
  border: 2rpx solid $border;
  background: transparent;
  transition: all 160ms ease;
}

.opt-text {
  flex: 1;
  font-size: $font-mid;
  color: $text-sub;
  line-height: 1.5;
}

.foot-tip {
  margin: 6rpx 8rpx 0;
  font-size: $font-caption;
  color: $text-placeholder;
  line-height: 1.7;
}

.empty {
  padding: 120rpx 0;
  text-align: center;
  font-size: $font-caption;
  color: $text-placeholder;
}

.actions {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 20rpx 24rpx calc(20rpx + env(safe-area-inset-bottom));
  background: $bg-card;
  border-top: 1rpx solid $border;
}

.save-btn {
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  border-radius: $radius;
  background: $brand;
  color: $text-on-accent;
  font-size: $font-body;
  font-weight: 650;

  &.done {
    background: $success;
  }

  &:active {
    opacity: 0.88;
  }
}
</style>
