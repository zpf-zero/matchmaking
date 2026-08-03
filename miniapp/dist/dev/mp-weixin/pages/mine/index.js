"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    const todo = common_vendor.ref(0);
    common_vendor.onShow(async () => {
      if (!store.isLogin)
        return;
      try {
        await store.fetchMe();
        store.fetchUnread();
        if (store.isAdminRole) {
          const ov = await api_index.adminApi.overview();
          todo.value = ov.pending_audit + ov.pending_report;
        }
      } catch {
      }
    });
    const statusText = common_vendor.computed(() => {
      switch (store.status) {
        case stores_user.Status.Incomplete:
          return { text: "资料未完善", cls: "warn" };
        case stores_user.Status.Auditing:
          return { text: "审核中", cls: "warn" };
        case stores_user.Status.Normal:
          return { text: "已通过审核", cls: "ok" };
        case stores_user.Status.Rejected:
          return { text: "审核未通过", cls: "danger" };
        case stores_user.Status.Banned:
          return { text: "账号已封禁", cls: "danger" };
        default:
          return { text: "", cls: "" };
      }
    });
    const roleText = common_vendor.computed(() => {
      return ["", "", "审核员", "管理员", "超级管理员"][store.role] || "";
    });
    function go(url) {
      common_vendor.index.navigateTo({ url });
    }
    return (_ctx, _cache) => {
      var _a, _b, _c;
      return common_vendor.e({
        a: (_a = common_vendor.unref(store).user) == null ? void 0 : _a.avatar,
        b: common_vendor.t(((_b = common_vendor.unref(store).user) == null ? void 0 : _b.nickname) || "未填写"),
        c: roleText.value
      }, roleText.value ? {
        d: common_vendor.t(roleText.value)
      } : {}, {
        e: common_vendor.t(statusText.value.text),
        f: common_vendor.n(statusText.value.cls),
        g: common_vendor.unref(store).profile
      }, common_vendor.unref(store).profile ? {
        h: common_vendor.t(common_vendor.unref(store).profile.complete_rate)
      } : {}, {
        i: common_vendor.o(($event) => go("/pages/profile/setup?mode=edit"), "a7"),
        j: common_vendor.unref(store).status === common_vendor.unref(stores_user.Status).Rejected
      }, common_vendor.unref(store).status === common_vendor.unref(stores_user.Status).Rejected ? {
        k: common_vendor.t(((_c = common_vendor.unref(store).profile) == null ? void 0 : _c.audit_remark) || "资料不符合规范")
      } : {}, {
        l: common_vendor.t(common_vendor.unref(store).unread.liked_me),
        m: common_vendor.o(($event) => go("/pages/interaction/list?tab=0"), "60"),
        n: common_vendor.o(($event) => go("/pages/interaction/list?tab=1"), "74"),
        o: common_vendor.t(common_vendor.unref(store).unread.viewed_me),
        p: common_vendor.o(($event) => go("/pages/interaction/list?tab=2"), "15"),
        q: common_vendor.o(($event) => go("/pages/profile/setup?mode=edit"), "6e"),
        r: common_vendor.o(($event) => go("/pages/user/requirement"), "1a"),
        s: common_vendor.t(common_vendor.unref(store).photos.length),
        t: common_vendor.o(($event) => go("/pages/user/photos"), "00"),
        v: common_vendor.o(($event) => go("/pages/setting/index"), "30"),
        w: common_vendor.unref(store).isAdminRole
      }, common_vendor.unref(store).isAdminRole ? common_vendor.e({
        x: todo.value
      }, todo.value ? {
        y: common_vendor.t(todo.value)
      } : {}, {
        z: common_vendor.o(($event) => go("/pages/admin/index"), "07")
      }) : {});
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-9023ef44"]]);
wx.createPage(MiniProgramPage);
