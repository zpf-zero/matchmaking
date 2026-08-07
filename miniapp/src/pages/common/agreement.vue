<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

type AgreementType = 'user' | 'privacy'

const type = ref<AgreementType>('user')

const userSections = [
  { title: '一、服务说明', content: '缘来婚恋为年满 18 周岁的单身用户提供资料展示、条件筛选、互动配对和站内聊天服务。平台不承诺用户一定能够完成配对或建立婚姻关系。' },
  { title: '二、账号使用', content: '用户应通过本人微信账号登录并妥善保管登录状态，不得转让、出租或出借账号。因用户主动泄露或保管不当造成的损失，由用户自行承担。' },
  { title: '三、资料真实性', content: '用户应提交真实、准确、完整的个人资料和照片，不得冒用他人身份，不得发布虚假婚恋状态、收入、学历、职业或资产信息。资料修改后可能需要重新审核。' },
  { title: '四、行为规范', content: '禁止发布违法违规、色情低俗、侮辱骚扰、诈骗引流、广告营销、赌博借贷等内容；禁止诱导他人脱离平台进行不安全交易或提供财物。' },
  { title: '五、交友安全', content: '用户应独立判断其他用户信息的真实性。线下见面建议选择公共场所并告知亲友，不轻信投资、借款、转账等要求。发现异常时可使用举报或拉黑功能。' },
  { title: '六、审核与处置', content: '平台可对资料、照片、举报和公开内容进行审核。对于违反本协议或影响其他用户安全的行为，平台可采取提示修改、限制功能、下架内容、封禁账号等措施。' },
  { title: '七、账号注销', content: '用户可在“设置”页面申请注销账号。注销后账号资料将被清除或匿名化，配对关系将关闭，相关操作不可恢复；依法需要保留的信息除外。' },
  { title: '八、协议更新', content: '平台可根据业务和法律法规变化更新本协议。涉及用户重要权益的变更将通过页面提示等合理方式告知。继续使用服务视为接受更新后的协议。' },
]

const privacySections = [
  { title: '一、信息收集范围', content: '为提供婚恋服务，我们可能收集微信账号标识、手机号、昵称、头像、性别、生日、所在地、婚姻状况、学历、职业、收入区间、择偶要求、个人介绍、相册，以及喜欢、浏览、配对、聊天、举报和拉黑记录。' },
  { title: '二、使用目的', content: '上述信息用于账号登录与安全验证、资料审核、推荐与筛选、匹配度计算、互动和聊天、消息提醒、投诉处理、风险控制以及改善产品体验。' },
  { title: '三、权限调用', content: '选择照片或发送图片时可能使用相册、相机相关能力；绑定手机号时将使用微信提供的手机号授权能力。我们仅在对应功能需要时发起授权，拒绝授权不会影响其他无关功能。' },
  { title: '四、信息展示', content: '审核通过的昵称、头像、性别、年龄、所在地、职业、个人介绍和相册等资料会向符合条件的其他用户展示。手机号不会在他人资料页公开展示。' },
  { title: '五、信息共享', content: '我们不会出售用户个人信息。为完成微信登录、手机号授权、内容安全、服务器托管和图片存储，可能在必要范围内向微信及相关技术服务提供方传输信息，并要求其采取安全保护措施。' },
  { title: '六、存储与安全', content: '我们将在实现服务目的所需的合理期限内保存信息，并采取访问控制、身份鉴权、日志审计等措施保护信息安全。具体运营主体、存储地点和保存期限将在小程序正式提审前补充确认。' },
  { title: '七、用户权利', content: '用户可在小程序内查看和修改个人资料、管理相册和择偶要求、删除聊天关系、拉黑用户，并可通过“设置—注销账号”申请删除账号信息。' },
  { title: '八、未成年人保护', content: '本服务仅面向年满 18 周岁的用户。我们不主动向未成年人提供婚恋交友服务，发现不符合年龄要求的账号时可停止提供服务。' },
  { title: '九、联系我们', content: '运营主体名称、联系邮箱及投诉渠道尚待主体资质确认，正式提审前将在本页面补充。用户当前可通过小程序内举报功能反馈违规内容。' },
]

onLoad((query) => {
  type.value = query?.type === 'privacy' ? 'privacy' : 'user'
  uni.setNavigationBarTitle({ title: type.value === 'privacy' ? '隐私政策' : '用户协议' })
})

const title = computed(() => (type.value === 'privacy' ? '隐私政策' : '用户协议'))
const sections = computed(() => (type.value === 'privacy' ? privacySections : userSections))
</script>

<template>
  <view class="agreement-page">
    <view class="title">{{ title }}</view>
    <view class="meta">更新日期：2026年8月7日</view>
    <view class="draft-tip">当前为开发阶段协议草案，运营主体和联系方式将在正式提审前补充。</view>
    <view v-for="section in sections" :key="section.title" class="section">
      <view class="section-title">{{ section.title }}</view>
      <view class="section-content">{{ section.content }}</view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.agreement-page { padding: 32rpx; background: #fff; min-height: 100vh; }
.title { text-align: center; font-size: 40rpx; font-weight: 600; color: #1f2329; }
.meta { text-align: center; margin-top: 12rpx; font-size: 22rpx; color: #8a8f99; }
.draft-tip { margin: 32rpx 0; padding: 22rpx 24rpx; border-radius: 12rpx; background: #fff7e8; color: #b76e00; font-size: 24rpx; line-height: 1.7; }
.section { margin-bottom: 34rpx; }
.section-title { margin-bottom: 14rpx; font-size: 30rpx; font-weight: 600; color: #1f2329; }
.section-content { font-size: 27rpx; line-height: 1.9; color: #4e5969; text-align: justify; }
</style>
