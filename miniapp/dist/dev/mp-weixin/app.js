"use strict";
Object.defineProperty(exports, Symbol.toStringTag, { value: "Module" });
const common_vendor = require("./common/vendor.js");
const stores_user = require("./stores/user.js");
const utils_ws = require("./utils/ws.js");
if (!Math) {
  "./pages/index/index.js";
  "./pages/search/index.js";
  "./pages/message/index.js";
  "./pages/mine/index.js";
  "./pages/login/index.js";
  "./pages/profile/setup.js";
  "./pages/user/detail.js";
  "./pages/user/requirement.js";
  "./pages/user/photos.js";
  "./pages/chat/index.js";
  "./pages/interaction/list.js";
  "./pages/setting/index.js";
  "./pages/setting/blacklist.js";
  "./pages/common/status.js";
  "./pages/admin/index.js";
  "./pages/admin/audit.js";
  "./pages/admin/report.js";
  "./pages/admin/user.js";
  "./pages/admin/user-detail.js";
}
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "App",
  setup(__props) {
    common_vendor.onLaunch(async () => {
      const store = stores_user.useUserStore();
      store.fetchConfig();
      store.fetchDict();
      if (!store.isLogin) {
        common_vendor.index.reLaunch({ url: "/pages/login/index" });
        return;
      }
      try {
        await store.fetchMe();
        if (store.status === 0) {
          common_vendor.index.reLaunch({ url: "/pages/profile/setup" });
          return;
        }
        if (store.canInteract)
          utils_ws.ws.connect();
      } catch {
      }
    });
    common_vendor.onShow(() => {
      const store = stores_user.useUserStore();
      if (store.isLogin && store.canInteract) {
        utils_ws.ws.connect();
        store.fetchUnread();
      }
    });
    common_vendor.onHide(() => {
      utils_ws.ws.close();
    });
    return () => {
    };
  }
});
function createApp() {
  const app = common_vendor.createSSRApp(_sfc_main);
  app.use(common_vendor.createPinia());
  return { app, Pinia: common_vendor.Pinia };
}
createApp().app.mount("#app");
exports.createApp = createApp;
