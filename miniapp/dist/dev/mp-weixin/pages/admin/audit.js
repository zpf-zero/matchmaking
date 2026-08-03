"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "audit",
  setup(__props) {
    const store = stores_user.useUserStore();
    const list = common_vendor.ref([]);
    const page = common_vendor.ref(1);
    const total = common_vendor.ref(0);
    const finished = common_vendor.ref(false);
    const loading = common_vendor.ref(false);
    const loaded = common_vendor.ref(false);
    common_vendor.onShow(async () => {
      await store.fetchDict();
      load(true);
    });
    common_vendor.onReachBottom(() => {
      if (!finished.value)
        load(false);
    });
    async function load(reset) {
      if (loading.value)
        return;
      loading.value = true;
      if (reset) {
        page.value = 1;
        finished.value = false;
      }
      try {
        const res = await api_index.adminApi.auditList(page.value);
        total.value = res.total;
        list.value = reset ? res.list || [] : [...list.value, ...res.list || []];
        finished.value = list.value.length >= res.total;
        if (!finished.value)
          page.value++;
        loaded.value = true;
      } catch {
        loaded.value = true;
      } finally {
        loading.value = false;
      }
    }
    function previewAll(item, index) {
      const urls = [item.avatar, ...item.photos.map((p) => p.url)].filter(Boolean);
      common_vendor.index.previewImage({ urls, current: index });
    }
    async function pass(item) {
      try {
        await api_index.adminApi.audit(item.user_id, true);
        list.value = list.value.filter((i) => i.user_id !== item.user_id);
        total.value = Math.max(0, total.value - 1);
        common_vendor.index.showToast({ title: "已通过", icon: "none" });
      } catch {
      }
    }
    function reject(item) {
      const presets = ["头像非真人照片", "资料信息不真实", "自我介绍含联系方式", "照片含违规内容", "自定义原因"];
      common_vendor.index.showActionSheet({
        itemList: presets,
        success: (res) => {
          if (res.tapIndex === presets.length - 1) {
            customReject(item);
          } else {
            doReject(item, presets[res.tapIndex]);
          }
        }
      });
    }
    function customReject(item) {
      common_vendor.index.showModal({
        title: "拒绝原因",
        editable: true,
        placeholderText: "请填写具体原因，用户会看到",
        success: (r) => {
          var _a;
          if (r.confirm && ((_a = r.content) == null ? void 0 : _a.trim())) {
            doReject(item, r.content.trim());
          }
        }
      });
    }
    async function doReject(item, remark) {
      try {
        await api_index.adminApi.audit(item.user_id, false, remark);
        list.value = list.value.filter((i) => i.user_id !== item.user_id);
        total.value = Math.max(0, total.value - 1);
        common_vendor.index.showToast({ title: "已拒绝", icon: "none" });
      } catch {
      }
    }
    function calcAge(item) {
      var _a;
      return ((_a = item.profile) == null ? void 0 : _a.age) || 0;
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.t(total.value),
        b: common_vendor.f(list.value, (item, k0, i0) => {
          var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k;
          return common_vendor.e({
            a: item.avatar,
            b: common_vendor.o(($event) => previewAll(item, 0), item.user_id),
            c: common_vendor.t(item.nickname),
            d: common_vendor.t(common_vendor.unref(store).label("gender", item.gender)),
            e: common_vendor.t(calcAge(item)),
            f: common_vendor.t((_a = item.profile) == null ? void 0 : _a.height),
            g: common_vendor.t((_b = item.profile) == null ? void 0 : _b.city),
            h: common_vendor.t(item.phone || "未绑定"),
            i: common_vendor.t(common_vendor.unref(store).label("education", (_c = item.profile) == null ? void 0 : _c.education)),
            j: common_vendor.t(common_vendor.unref(store).label("income_level", (_d = item.profile) == null ? void 0 : _d.income_level)),
            k: common_vendor.t(common_vendor.unref(store).label("marital_status", (_e = item.profile) == null ? void 0 : _e.marital_status)),
            l: common_vendor.t(common_vendor.unref(store).label("has_children", (_f = item.profile) == null ? void 0 : _f.has_children)),
            m: common_vendor.t((_g = item.profile) == null ? void 0 : _g.occupation),
            n: common_vendor.t(((_h = item.profile) == null ? void 0 : _h.school) || "—"),
            o: common_vendor.t((_i = item.profile) == null ? void 0 : _i.self_intro),
            p: (_j = item.photos) == null ? void 0 : _j.length
          }, ((_k = item.photos) == null ? void 0 : _k.length) ? {
            q: common_vendor.f(item.photos, (p, i, i1) => {
              return {
                a: p.id,
                b: p.url,
                c: common_vendor.o(($event) => previewAll(item, i + 1), p.id)
              };
            })
          } : {}, {
            r: common_vendor.o(($event) => reject(item), item.user_id),
            s: common_vendor.o(($event) => pass(item), item.user_id),
            t: item.user_id
          });
        }),
        c: loaded.value && !list.value.length
      }, loaded.value && !list.value.length ? {} : list.value.length && finished.value ? {} : {}, {
        d: list.value.length && finished.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-dccb1e66"]]);
wx.createPage(MiniProgramPage);
