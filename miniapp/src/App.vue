<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/user'
import { ws } from '@/utils/ws'

/**
 * 全局实时通知。
 * 用户在任何页面被喜欢/配对成功都能立刻感知，
 * 而不是要等切到消息页才发现。
 */
function setupGlobalNotify(store: ReturnType<typeof useUserStore>) {
  ws.on('like', (data: any) => {
    store.fetchUnread()
    uni.showToast({
      title: `${data?.nickname || '有人'} 喜欢了你 ❤`,
      icon: 'none',
      duration: 2500,
    })
  })

  ws.on('match', (data: any) => {
    store.fetchUnread()
    const name = data?.peer?.nickname || '对方'
    uni.showModal({
      title: '配对成功',
      content: `你和 ${name} 互相喜欢，现在可以聊天了`,
      confirmText: '去聊天',
      cancelText: '稍后',
      success: (r) => {
        if (r.confirm && data?.conversation_id) {
          uni.navigateTo({
            url: `/pages/chat/index?id=${data.conversation_id}&name=${encodeURIComponent(name)}&peer=${data?.peer?.id || ''}`,
          })
        }
      },
    })
  })

  // 收到新消息刷新红点（聊天页会自己处理消息本身）
  ws.on('message', () => store.fetchUnread())
}

onLaunch(async () => {
  const store = useUserStore()
  store.fetchConfig()
  store.fetchDict()
  setupGlobalNotify(store)

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
