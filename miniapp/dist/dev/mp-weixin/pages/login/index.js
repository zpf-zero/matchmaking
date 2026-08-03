"use strict";
const common_vendor = require("../../common/vendor.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    const agreed = common_vendor.ref(false);
    const loading = common_vendor.ref(false);
    const mockCode = common_vendor.ref("test_user_1");
    common_vendor.onLoad(() => {
      store.fetchConfig();
    });
    async function handleLogin() {
      if (!agreed.value) {
        common_vendor.index.showToast({ title: "请先阅读并同意协议", icon: "none" });
        return;
      }
      if (loading.value)
        return;
      loading.value = true;
      try {
        const code = store.mockMode ? mockCode.value.trim() : await getWxCode();
        if (!code) {
          common_vendor.index.showToast({ title: "请填写测试账号", icon: "none" });
          return;
        }
        const res = await store.login(code);
        await store.fetchMe();
        if (res.need_profile || store.status === 0) {
          common_vendor.index.reLaunch({ url: "/pages/profile/setup" });
        } else {
          common_vendor.index.reLaunch({ url: "/pages/index/index" });
        }
      } catch {
      } finally {
        loading.value = false;
      }
    }
    function getWxCode() {
      return new Promise((resolve, reject) => {
        common_vendor.index.login({
          provider: "weixin",
          success: (res) => resolve(res.code),
          fail: reject
        });
      });
    }
    function openAgreement(type) {
      common_vendor.index.showModal({
        title: type === "user" ? "用户协议" : "隐私政策",
        content: "协议文本待补充。提审前必须提供正式文本。",
        showCancel: false
      });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.unref(store).mockMode
      }, common_vendor.unref(store).mockMode ? {
        b: mockCode.value,
        c: common_vendor.o(($event) => mockCode.value = $event.detail.value, "6e")
      } : {}, {
        d: common_vendor.t(loading.value ? "登录中..." : common_vendor.unref(store).mockMode ? "模拟登录" : "微信一键登录"),
        e: loading.value ? 1 : "",
        f: common_vendor.o(handleLogin, "0b"),
        g: agreed.value
      }, agreed.value ? {} : {}, {
        h: agreed.value ? 1 : "",
        i: common_vendor.o(($event) => openAgreement("user"), "22"),
        j: common_vendor.o(($event) => openAgreement("privacy"), "7b"),
        k: common_vendor.o(($event) => agreed.value = !agreed.value, "0a")
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-45258083"]]);
wx.createPage(MiniProgramPage);
