"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    const stat = common_vendor.ref({
      total_user: 0,
      today_new: 0,
      pending_audit: 0,
      pending_report: 0,
      matched: 0,
      active_week: 0
    });
    common_vendor.onShow(async () => {
      try {
        stat.value = await api_index.adminApi.overview();
      } catch {
      }
    });
    function go(url) {
      common_vendor.index.navigateTo({ url });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.t(["", "普通用户", "审核员", "管理员", "超级管理员"][common_vendor.unref(store).role]),
        b: common_vendor.t(stat.value.total_user),
        c: common_vendor.t(stat.value.today_new),
        d: common_vendor.t(stat.value.active_week),
        e: common_vendor.t(stat.value.matched),
        f: stat.value.pending_audit
      }, stat.value.pending_audit ? {
        g: common_vendor.t(stat.value.pending_audit)
      } : {}, {
        h: common_vendor.o(($event) => go("/pages/admin/audit"), "cc"),
        i: stat.value.pending_report
      }, stat.value.pending_report ? {
        j: common_vendor.t(stat.value.pending_report)
      } : {}, {
        k: common_vendor.o(($event) => go("/pages/admin/report"), "e1"),
        l: common_vendor.unref(store).role < common_vendor.unref(stores_user.Role).Admin
      }, common_vendor.unref(store).role < common_vendor.unref(stores_user.Role).Admin ? {} : {}, {
        m: common_vendor.o(($event) => go("/pages/admin/user"), "b0")
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-f94c733c"]]);
wx.createPage(MiniProgramPage);
