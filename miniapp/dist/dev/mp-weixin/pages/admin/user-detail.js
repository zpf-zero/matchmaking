"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "user-detail",
  setup(__props) {
    const store = stores_user.useUserStore();
    const uid = common_vendor.ref(0);
    const data = common_vendor.ref(null);
    const loaded = common_vendor.ref(false);
    const statusText = ["待完善", "待审核", "正常", "未通过", "已封禁"];
    const reasonText = ["", "虚假资料", "骚扰", "广告", "违法信息", "其他"];
    common_vendor.onLoad(async (q) => {
      uid.value = Number((q == null ? void 0 : q.id) || 0);
      await store.fetchDict();
      await load();
    });
    async function load() {
      try {
        data.value = await api_index.adminApi.userDetail(uid.value);
      } catch {
      } finally {
        loaded.value = true;
      }
    }
    function ban() {
      common_vendor.index.showModal({
        title: "封禁用户",
        editable: true,
        placeholderText: "封禁原因，将展示给该用户",
        success: async (r) => {
          var _a;
          if (!r.confirm)
            return;
          const reason = (_a = r.content) == null ? void 0 : _a.trim();
          if (!reason) {
            common_vendor.index.showToast({ title: "请填写封禁原因", icon: "none" });
            return;
          }
          try {
            await api_index.adminApi.ban(uid.value, reason);
            common_vendor.index.showToast({ title: "已封禁", icon: "none" });
            load();
          } catch {
          }
        }
      });
    }
    function unban() {
      common_vendor.index.showModal({
        title: "解除封禁",
        content: "解封后该用户资料需重新通过审核才能正常使用",
        success: async (r) => {
          if (!r.confirm)
            return;
          try {
            await api_index.adminApi.unban(uid.value);
            common_vendor.index.showToast({ title: "已解封", icon: "none" });
            load();
          } catch {
          }
        }
      });
    }
    function preview(urls, i) {
      common_vendor.index.previewImage({ urls, current: i });
    }
    return (_ctx, _cache) => {
      var _a, _b, _c, _d, _e, _f;
      return common_vendor.e({
        a: data.value
      }, data.value ? common_vendor.e({
        b: data.value.user.avatar,
        c: common_vendor.o(($event) => preview([data.value.user.avatar], 0), "8b"),
        d: common_vendor.t(data.value.user.nickname || "未填写"),
        e: common_vendor.t(data.value.user.id),
        f: common_vendor.t(common_vendor.unref(store).label("gender", data.value.user.gender)),
        g: common_vendor.t(data.value.phone || "未绑定"),
        h: common_vendor.t(statusText[data.value.user.status]),
        i: common_vendor.n("s" + data.value.user.status),
        j: data.value.user.ban_reason
      }, data.value.user.ban_reason ? {
        k: common_vendor.t(data.value.user.ban_reason)
      } : {}, {
        l: common_vendor.t(data.value.stats.like_count),
        m: common_vendor.t(data.value.stats.liked_count),
        n: common_vendor.t(data.value.stats.match_count),
        o: (_a = data.value.profile) == null ? void 0 : _a.id
      }, ((_b = data.value.profile) == null ? void 0 : _b.id) ? {
        p: common_vendor.t(data.value.profile.age),
        q: common_vendor.t(data.value.profile.height),
        r: common_vendor.t(common_vendor.unref(store).label("education", data.value.profile.education)),
        s: common_vendor.t(common_vendor.unref(store).label("income_level", data.value.profile.income_level)),
        t: common_vendor.t(common_vendor.unref(store).label("marital_status", data.value.profile.marital_status)),
        v: common_vendor.t(common_vendor.unref(store).label("has_children", data.value.profile.has_children)),
        w: common_vendor.t(data.value.profile.occupation),
        x: common_vendor.t(data.value.profile.city),
        y: common_vendor.t(data.value.profile.self_intro)
      } : {}, {
        z: (_c = data.value.photos) == null ? void 0 : _c.length
      }, ((_d = data.value.photos) == null ? void 0 : _d.length) ? {
        A: common_vendor.t(data.value.photos.length),
        B: common_vendor.f(data.value.photos, (p, i, i0) => {
          return {
            a: p.id,
            b: p.url,
            c: common_vendor.o(($event) => preview(data.value.photos.map((x) => x.url), i), p.id)
          };
        })
      } : {}, {
        C: (_e = data.value.reports) == null ? void 0 : _e.length
      }, ((_f = data.value.reports) == null ? void 0 : _f.length) ? {
        D: common_vendor.t(data.value.reports.length),
        E: common_vendor.f(data.value.reports, (r, k0, i0) => {
          return {
            a: common_vendor.t(reasonText[r.reason]),
            b: common_vendor.t(r.description || "无描述"),
            c: common_vendor.t(["待处理", "已处理", "已驳回"][r.status]),
            d: r.id
          };
        })
      } : {}, {
        F: common_vendor.unref(store).role >= common_vendor.unref(stores_user.Role).Admin
      }, common_vendor.unref(store).role >= common_vendor.unref(stores_user.Role).Admin ? common_vendor.e({
        G: data.value.user.status !== 4
      }, data.value.user.status !== 4 ? {
        H: common_vendor.o(ban, "d9")
      } : {
        I: common_vendor.o(unban, "6f")
      }) : {}) : loaded.value ? {} : {}, {
        J: loaded.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-e7bd75d2"]]);
wx.createPage(MiniProgramPage);
