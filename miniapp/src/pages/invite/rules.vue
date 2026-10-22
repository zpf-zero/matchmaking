<script setup lang="ts">
/**
 * 活动规则页。
 *
 * 《微信小程序平台运营规范》5.1.5 要求分享活动的完整信息（整体规则、活动步骤、
 * 互动利益结果）必须清晰、显著地告知用户，不得用小微字体或灰色字体规避阅读。
 * 所以本页刻意用与正文同级的字号和对比度，不做视觉降级。
 *
 * 文案上同样回避一切诱导性表述：不写「不分享就…」、不写「必得」、不写抽奖和
 * 概率性收益（5.1.1 / 5.1.6），也不承诺「长期参与才可得」（5.1.7）。
 */
</script>

<template>
  <view class="agreement-page">
    <view class="hd">
      <view class="hd-title">邀请有礼 · 活动规则</view>
      <view class="hd-sub">最后更新：2026-10-21</view>
    </view>

    <view class="sec">
      <view class="sec-title">一、活动内容</view>
      <view class="p">
        把你的专属邀请码告诉朋友，朋友在填写资料时填写该邀请码。当朋友完成注册、
        完善资料并通过平台人工审核后，你和朋友双方各解锁
        <text class="em">「缘来婚恋」完整资料查看权益（永久有效）</text>。
      </view>
      <view class="p">
        <text class="em">权益一次解锁即完整拥有。</text>邀请一位好友通过审核即可获得全部权益，
        邀请更多好友不会带来额外权益，平台也不设「邀请越多、权益越多」的阶梯规则。
      </view>
    </view>

    <view class="sec">
      <view class="sec-title">二、参与步骤</view>
      <view class="step">1. 在「邀请有礼」页面获取你的专属邀请码；</view>
      <view class="step">2. 将邀请码发给朋友，或通过页面的分享按钮把小程序分享给朋友；</view>
      <view class="step">3. 朋友通过小程序完成注册，填写资料时填入你的邀请码；</view>
      <view class="step">4. 朋友提交资料并通过平台人工审核；</view>
      <view class="step">5. 审核通过的当刻，双方权益自动到账，无需额外操作。</view>
    </view>

    <view class="sec">
      <view class="sec-title">三、权益内容与发放时间</view>
      <view class="p">
        <text class="em">权益内容：</text>邀请人与被邀请人各解锁以下内容，固定额度，
        不涉及抽奖或随机结果。
      </view>
      <view class="li">· 更多照片：免费用户可看前 3 张，解锁后可查看对方全部照片；</view>
      <view class="li">· 三观问答答案：答题本身免费，解锁后可查看彼此的选择与一致率；</view>
      <view class="li">· 对方的原生家庭：仅展示对方主动授权公开的部分；</view>
      <view class="li">· 完整「谁看过我」列表（免费版可查看最近 3 位）；</view>
      <view class="li">· 推荐页曝光加权，资料优先展示给更多合适的用户。</view>
      <view class="p">
        <text class="em">发放时间：</text>朋友通过人工审核后即时发放，永久有效。
        你可以在「邀请有礼」页面查看每一条邀请记录的状态。
      </view>
    </view>

    <view class="sec">
      <view class="sec-title">四、参与限制</view>
      <view class="li">· 每位用户仅可使用一次他人的邀请码，不可重复使用；</view>
      <view class="li">· 邀请码仅限在首次提交资料审核前填写，已通过审核的账号无法再使用邀请码；</view>
      <view class="li">· 不能使用自己的邀请码；</view>
      <view class="li">· 每个账号最多累计结算 20 位邀请好友。该数字为防范批量注册刷量的
        风控上限，并非权益目标 —— 如前所述，一位好友通过审核即可获得全部权益；</view>
      <view class="li">· 邀请人的账号须已通过平台审核且状态正常，未过审或违规被封禁的账号不参与结算；</view>
      <view class="li">· 权益为一次性解锁，解锁后长期有效，无需也不能重复领取。</view>
    </view>

    <view class="sec">
      <view class="sec-title">五、关于「分享」的说明</view>
      <view class="p">
        分享小程序<text class="em">不是获得奖励的前提条件</text>，平台也不会检测你是否完成分享。
        奖励的发放唯一取决于被邀请的好友是否真实完成注册并通过资料审核。
        你可以仅通过发送邀请码的方式参与本活动。
      </view>
      <view class="p note">
        平台所有基础功能（浏览、筛选、喜欢、聊天）均可免费使用，
        未获得权益不影响任何功能的使用。
      </view>
    </view>

    <view class="sec">
      <view class="sec-title">六、隐私说明</view>
      <view class="p">
        「原生家庭」属于敏感个人信息，平台默认不对其他用户展示。
        只有在你主动开启「对已解锁用户可见」后，才会向已解锁的用户展示你填写的部分；
        随时可在此关闭。三观问答的答案是公开的，但查看他人的答案需要拥有权益。
      </view>
    </view>

    <view class="sec">
      <view class="sec-title">七、异常处理</view>
      <view class="p">
        如发现通过批量注册、虚假资料、机器操作等方式骗取权益，平台有权撤销相关权益
        并对涉事账号做相应处理。如对权益发放有疑问，可通过「设置 - 意见反馈」联系我们。
      </view>
    </view>

    <view class="ft">缘来婚恋 · 认真认识，再决定心意</view>
  </view>
</template>

<style lang="scss" scoped>
.agreement-page {
  min-height: 100dvh;
  padding: 36rpx 32rpx calc(60rpx + env(safe-area-inset-bottom));
  background: $bg-page;
}

.hd {
  margin-bottom: 40rpx;
}

.hd-title {
  font-size: $font-display;
  font-weight: 650;
  color: $text-main;
  letter-spacing: -1rpx;
}

.hd-sub {
  margin-top: 12rpx;
  font-size: $font-micro;
  color: $text-placeholder;
}

.sec {
  margin-bottom: 44rpx;
}

.sec-title {
  font-size: $font-lead;
  font-weight: 600;
  color: $text-main;
  margin-bottom: 18rpx;
}

.p {
  font-size: $font-body;
  color: $text-sub;
  line-height: 1.8;
  margin-bottom: 14rpx;

  &.note {
    color: $text-weak;
    font-size: $font-caption;
  }
}

/* 与正文同级字号，不做视觉降级 —— 规范明确禁止用小字体/灰字让用户难以阅读 */
.li,
.step {
  font-size: $font-body;
  color: $text-sub;
  line-height: 1.8;
  margin-bottom: 8rpx;
}

.step {
  margin-bottom: 12rpx;
}

.em {
  color: $text-main;
  font-weight: 600;
}

.ft {
  margin-top: 60rpx;
  padding-top: 30rpx;
  border-top: 1rpx solid $border;
  text-align: center;
  font-size: $font-caption;
  color: $text-placeholder;
}
</style>
