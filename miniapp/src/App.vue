<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { ws } from '@/utils/ws'

onLaunch(async () => {
  const store = useUserStore()
  store.fetchConfig()
  store.fetchDict()

  if (!store.isLogin) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }

  try {
    await store.fetchMe()
    // 资料没填完直接引导去填，否则首页是空的，用户不知道该干什么
    if (store.status === 0) {
      uni.reLaunch({ url: '/pages/profile/setup' })
      return
    }
    if (store.canInteract) ws.connect()
  } catch {
    // fetchMe 失败时 request 层已按状态码做了跳转
  }
})

onShow(() => {
  const store = useUserStore()
  if (store.isLogin && store.canInteract) {
    // 小程序切后台会被断连，回前台重新接上
    ws.connect()
    store.fetchUnread()
  }
})

onHide(() => {
  ws.close()
})
</script>

<style lang="scss">
page {
  background-color: #f5f6f8;
  color: #1f2329;
  font-size: 28rpx;
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, sans-serif;
}

/* ---------- 全局通用类 ---------- */
.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.row {
  display: flex;
  align-items: center;
}

.row-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.btn-primary {
  background: #e8506d;
  color: #fff;
  border-radius: 44rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: 30rpx;

  &.disabled {
    background: #f0c2cc;
  }
}

.btn-plain {
  background: #fff;
  color: #646a73;
  border: 1rpx solid #ebedf0;
  border-radius: 44rpx;
  height: 88rpx;
  line-height: 88rpx;
  text-align: center;
  font-size: 30rpx;
}

.empty {
  padding: 120rpx 40rpx;
  text-align: center;
  color: #8a8f99;
  font-size: 28rpx;
}

.tag {
  display: inline-block;
  padding: 4rpx 14rpx;
  border-radius: 6rpx;
  font-size: 22rpx;
  background: #f5f6f8;
  color: #646a73;
  margin-right: 10rpx;
}

.tag-brand {
  background: #fff0f3;
  color: #e8506d;
}
</style>
