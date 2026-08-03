"use strict";
const common_vendor = require("../../common/vendor.js");
const utils_request = require("../../utils/request.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "status",
  setup(__props) {
    const code = common_vendor.ref(0);
    const msg = common_vendor.ref("");
    common_vendor.onLoad((q) => {
      code.value = Number((q == null ? void 0 : q.code) || 0);
      msg.value = decodeURIComponent((q == null ? void 0 : q.msg) || "");
    });
    const view = () => {
      switch (code.value) {
        case utils_request.Code.ProfileAuditing:
          return {
            icon: "⏳",
            title: "资料审核中",
            desc: "我们正在核实你的资料，通常 24 小时内完成。期间可以先浏览其他人。",
            btn: "去逛逛",
            action: () => common_vendor.index.reLaunch({ url: "/pages/index/index" })
          };
        case utils_request.Code.ProfileRejected:
          return {
            icon: "⚠️",
            title: "资料未通过审核",
            desc: msg.value || "资料不符合规范，请修改后重新提交。",
            btn: "去修改资料",
            action: () => common_vendor.index.reLaunch({ url: "/pages/profile/setup?mode=edit" })
          };
        case utils_request.Code.Banned:
          return {
            icon: "🚫",
            title: "账号已被封禁",
            desc: msg.value || "你的账号因违反社区规范已被封禁。如有疑问请联系客服。",
            btn: "",
            action: () => {
            }
          };
        default:
          return {
            icon: "ℹ️",
            title: "提示",
            desc: msg.value || "当前操作无法完成",
            btn: "返回",
            action: () => common_vendor.index.navigateBack()
          };
      }
    };
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.t(view().icon),
        b: common_vendor.t(view().title),
        c: common_vendor.t(view().desc),
        d: view().btn
      }, view().btn ? {
        e: common_vendor.t(view().btn),
        f: common_vendor.o(($event) => view().action(), "f4")
      } : {});
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-2ba0d154"]]);
wx.createPage(MiniProgramPage);
