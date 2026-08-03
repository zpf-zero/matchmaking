"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "user",
  setup(__props) {
    const store = stores_user.useUserStore();
    const list = common_vendor.ref([]);
    const keyword = common_vendor.ref("");
    const status = common_vendor.ref(-1);
    const page = common_vendor.ref(1);
    const total = common_vendor.ref(0);
    const finished = common_vendor.ref(false);
    const loading = common_vendor.ref(false);
    const loaded = common_vendor.ref(false);
    const statusOptions = [
      { v: -1, t: "全部" },
      { v: 1, t: "待审核" },
      { v: 2, t: "正常" },
      { v: 3, t: "未通过" },
      { v: 4, t: "已封禁" }
    ];
    const statusText = ["待完善", "待审核", "正常", "未通过", "已封禁"];
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
        const res = await api_index.adminApi.userList({
          keyword: keyword.value.trim(),
          status: status.value,
          page: page.value
        });
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
    function setStatus(v) {
      status.value = v;
      load(true);
    }
    function goDetail(id) {
      common_vendor.index.navigateTo({ url: `/pages/admin/user-detail?id=${id}` });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.o(($event) => load(true), "d2"),
        b: keyword.value,
        c: common_vendor.o(($event) => keyword.value = $event.detail.value, "90"),
        d: common_vendor.o(($event) => load(true), "14"),
        e: common_vendor.f(statusOptions, (o, k0, i0) => {
          return {
            a: common_vendor.t(o.t),
            b: o.v,
            c: status.value === o.v ? 1 : "",
            d: common_vendor.o(($event) => setStatus(o.v), o.v)
          };
        }),
        f: common_vendor.t(total.value),
        g: common_vendor.f(list.value, (u, k0, i0) => {
          return common_vendor.e({
            a: u.avatar,
            b: common_vendor.t(u.nickname || "未填写"),
            c: common_vendor.t(u.id),
            d: u.role > 1
          }, u.role > 1 ? {
            e: common_vendor.t(["", "", "审核员", "管理员", "超管"][u.role])
          } : {}, {
            f: common_vendor.t(u.age ? u.age + "岁 · " : ""),
            g: common_vendor.t(common_vendor.unref(store).label("gender", u.gender)),
            h: common_vendor.t(u.city ? " · " + u.city : ""),
            i: common_vendor.t(u.phone || "未绑定手机号"),
            j: common_vendor.t(statusText[u.status]),
            k: common_vendor.n("s" + u.status),
            l: u.id,
            m: common_vendor.o(($event) => goDetail(u.id), u.id)
          });
        }),
        h: loaded.value && !list.value.length
      }, loaded.value && !list.value.length ? {} : list.value.length && finished.value ? {} : {}, {
        i: list.value.length && finished.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-e781e4d5"]]);
wx.createPage(MiniProgramPage);
