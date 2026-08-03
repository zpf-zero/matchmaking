"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
if (!Math) {
  UserCardItem();
}
const UserCardItem = () => "../../components/UserCardItem.js";
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    const list = common_vendor.ref([]);
    const page = common_vendor.ref(1);
    const total = common_vendor.ref(0);
    const loading = common_vendor.ref(false);
    const finished = common_vendor.ref(false);
    const firstLoaded = common_vendor.ref(false);
    common_vendor.onShow(() => {
      if (!store.canBrowse)
        return;
      if (!firstLoaded.value)
        load(true);
      store.fetchUnread();
    });
    common_vendor.onPullDownRefresh(async () => {
      await load(true);
      common_vendor.index.stopPullDownRefresh();
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
        const res = await api_index.discoverApi.recommend(page.value);
        total.value = res.total;
        list.value = reset ? res.list || [] : [...list.value, ...res.list || []];
        finished.value = list.value.length >= res.total;
        if (!finished.value)
          page.value++;
        firstLoaded.value = true;
      } catch {
      } finally {
        loading.value = false;
      }
    }
    async function handleLike(user) {
      if (!store.canInteract) {
        common_vendor.index.showToast({ title: "资料审核通过后才能表达心意", icon: "none" });
        return;
      }
      try {
        const res = await api_index.interactionApi.like(user.id);
        list.value = list.value.filter((u) => u.id !== user.id);
        total.value = Math.max(0, total.value - 1);
        if (res.is_match) {
          common_vendor.index.showModal({
            title: "配对成功",
            content: `你和 ${user.nickname} 互相喜欢，现在可以聊天了`,
            confirmText: "去聊天",
            cancelText: "稍后",
            success: (r) => {
              if (r.confirm) {
                common_vendor.index.navigateTo({
                  url: `/pages/chat/index?id=${res.conversation_id}&name=${encodeURIComponent(user.nickname)}`
                });
              }
            }
          });
        } else {
          common_vendor.index.showToast({ title: "已表达心意", icon: "none" });
        }
      } catch {
      }
    }
    function goSearch() {
      common_vendor.index.switchTab({ url: "/pages/search/index" });
    }
    function goEditProfile() {
      common_vendor.index.navigateTo({ url: "/pages/profile/setup?mode=edit" });
    }
    function goRequirement() {
      common_vendor.index.navigateTo({ url: "/pages/user/requirement" });
    }
    return (_ctx, _cache) => {
      var _a;
      return common_vendor.e({
        a: common_vendor.unref(store).status === 1
      }, common_vendor.unref(store).status === 1 ? {} : common_vendor.unref(store).status === 3 ? {
        c: common_vendor.t((_a = common_vendor.unref(store).profile) == null ? void 0 : _a.audit_remark),
        d: common_vendor.o(goEditProfile, "96")
      } : {}, {
        b: common_vendor.unref(store).status === 3,
        e: common_vendor.o(goSearch, "fe"),
        f: list.value.length
      }, list.value.length ? common_vendor.e({
        g: common_vendor.f(list.value, (u, k0, i0) => {
          return {
            a: u.id,
            b: common_vendor.o(handleLike, u.id),
            c: "83a5a03c-0-" + i0,
            d: common_vendor.p({
              user: u,
              ["show-like"]: true
            })
          };
        }),
        h: finished.value
      }, finished.value ? {} : {}) : firstLoaded.value ? {
        j: common_vendor.o(goRequirement, "6f")
      } : {}, {
        i: firstLoaded.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-83a5a03c"]]);
wx.createPage(MiniProgramPage);
