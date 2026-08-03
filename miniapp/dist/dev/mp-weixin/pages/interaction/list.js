"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
if (!Math) {
  UserCardItem();
}
const UserCardItem = () => "../../components/UserCardItem.js";
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "list",
  setup(__props) {
    const store = stores_user.useUserStore();
    const tabs = ["谁喜欢我", "我喜欢的", "谁看过我"];
    const active = common_vendor.ref(0);
    const list = common_vendor.ref([]);
    const page = common_vendor.ref(1);
    const finished = common_vendor.ref(false);
    const loading = common_vendor.ref(false);
    const loaded = common_vendor.ref(false);
    common_vendor.onLoad(async (q) => {
      active.value = Number((q == null ? void 0 : q.tab) || 0);
      await store.fetchDict();
      switchTab(active.value);
    });
    common_vendor.onReachBottom(() => {
      if (!finished.value)
        load(false);
    });
    function switchTab(i) {
      active.value = i;
      list.value = [];
      loaded.value = false;
      load(true);
      if (i === 0)
        api_index.interactionApi.markRead(2).then(() => store.fetchUnread());
      if (i === 2)
        api_index.interactionApi.markRead(1).then(() => store.fetchUnread());
    }
    async function load(reset) {
      if (loading.value)
        return;
      loading.value = true;
      if (reset) {
        page.value = 1;
        finished.value = false;
      }
      const fn = [api_index.interactionApi.likedMe, api_index.interactionApi.myLikes, api_index.interactionApi.viewedMe][active.value];
      try {
        const res = await fn(page.value);
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
    const emptyText = () => ["还没有人喜欢你，完善资料能提高曝光", "你还没有喜欢过任何人", "还没有人看过你的资料"][active.value];
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.f(tabs, (t, i, i0) => {
          return {
            a: common_vendor.t(t),
            b: i,
            c: active.value === i ? 1 : "",
            d: common_vendor.o(($event) => switchTab(i), i)
          };
        }),
        b: common_vendor.f(list.value, (u, k0, i0) => {
          return {
            a: u.id,
            b: "65272a63-0-" + i0,
            c: common_vendor.p({
              user: u
            })
          };
        }),
        c: loaded.value && !list.value.length
      }, loaded.value && !list.value.length ? {
        d: common_vendor.t(emptyText())
      } : list.value.length && finished.value ? {} : {}, {
        e: list.value.length && finished.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-65272a63"]]);
wx.createPage(MiniProgramPage);
