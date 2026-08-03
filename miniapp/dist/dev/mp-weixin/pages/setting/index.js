"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    function goBlacklist() {
      common_vendor.index.navigateTo({ url: "/pages/setting/blacklist" });
    }
    function maskPhone(p) {
      if (!p || p.length < 11)
        return p || "未绑定";
      return p.slice(0, 3) + "****" + p.slice(7);
    }
    function showAgreement(type) {
      common_vendor.index.showModal({
        title: type === "user" ? "用户协议" : "隐私政策",
        content: "协议文本待补充。提审前必须提供正式文本。",
        showCancel: false
      });
    }
    function logout() {
      common_vendor.index.showModal({
        title: "退出登录",
        content: "确定退出当前账号吗？",
        success: async (r) => {
          if (!r.confirm)
            return;
          try {
            await api_index.authApi.logout();
          } catch {
          }
          store.logout();
        }
      });
    }
    function deleteAccount() {
      common_vendor.index.showModal({
        title: "注销账号",
        content: "注销后资料、配对和聊天记录将无法恢复，确定继续吗？",
        confirmText: "确认注销",
        confirmColor: "#f53f3f",
        success: async (r) => {
          if (!r.confirm)
            return;
          common_vendor.index.showModal({
            title: "再次确认",
            content: "这是不可逆操作，请再次确认",
            confirmText: "仍要注销",
            confirmColor: "#f53f3f",
            success: async (r2) => {
              if (!r2.confirm)
                return;
              try {
                await api_index.authApi.deleteAccount();
                common_vendor.index.showToast({ title: "账号已注销", icon: "none" });
                setTimeout(() => store.logout(), 1e3);
              } catch {
              }
            }
          });
        }
      });
    }
    return (_ctx, _cache) => {
      return {
        a: common_vendor.t(maskPhone(common_vendor.unref(store).phone)),
        b: common_vendor.o(goBlacklist, "d3"),
        c: common_vendor.o(($event) => showAgreement("user"), "78"),
        d: common_vendor.o(($event) => showAgreement("privacy"), "30"),
        e: common_vendor.o(logout, "dc"),
        f: common_vendor.o(deleteAccount, "10")
      };
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-3b30043e"]]);
wx.createPage(MiniProgramPage);
