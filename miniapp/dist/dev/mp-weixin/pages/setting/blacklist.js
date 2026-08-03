"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "blacklist",
  setup(__props) {
    const list = common_vendor.ref([]);
    const loaded = common_vendor.ref(false);
    common_vendor.onShow(load);
    async function load() {
      try {
        list.value = await api_index.commonApi.blacklist();
      } catch {
      } finally {
        loaded.value = true;
      }
    }
    function remove(userId, nickname) {
      common_vendor.index.showModal({
        title: "移出黑名单",
        content: `将 ${nickname} 移出黑名单后，你们可以重新看到对方`,
        success: async (r) => {
          if (!r.confirm)
            return;
          try {
            await api_index.commonApi.removeBlacklist(userId);
            await load();
            common_vendor.index.showToast({ title: "已移出", icon: "none" });
          } catch {
          }
        }
      });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: list.value.length
      }, list.value.length ? {
        b: common_vendor.f(list.value, (b, k0, i0) => {
          return {
            a: b.avatar,
            b: common_vendor.t(b.nickname),
            c: common_vendor.o(($event) => remove(b.user_id, b.nickname), b.user_id),
            d: b.user_id
          };
        })
      } : loaded.value ? {} : {}, {
        c: loaded.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-0172ab08"]]);
wx.createPage(MiniProgramPage);
