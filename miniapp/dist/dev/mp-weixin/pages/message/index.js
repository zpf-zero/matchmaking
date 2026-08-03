"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const utils_ws = require("../../utils/ws.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    const list = common_vendor.ref([]);
    const loaded = common_vendor.ref(false);
    function onNewMessage() {
      load();
    }
    common_vendor.onShow(() => {
      if (!store.canInteract) {
        loaded.value = true;
        return;
      }
      load();
      store.fetchUnread();
      utils_ws.ws.on("message", onNewMessage);
      utils_ws.ws.on("match", onNewMessage);
    });
    common_vendor.onUnload(() => {
      utils_ws.ws.off("message", onNewMessage);
      utils_ws.ws.off("match", onNewMessage);
    });
    async function load() {
      try {
        list.value = await api_index.chatApi.conversations();
      } catch {
      } finally {
        loaded.value = true;
      }
    }
    function goChat(c) {
      common_vendor.index.navigateTo({
        url: `/pages/chat/index?id=${c.id}&name=${encodeURIComponent(c.peer_nickname)}&peer=${c.peer_id}`
      });
    }
    function goInteraction(tab) {
      common_vendor.index.navigateTo({ url: `/pages/interaction/list?tab=${tab}` });
    }
    function timeText(t) {
      if (!t)
        return "";
      const d = new Date(t.replace(/-/g, "/").replace(/\..*$/, ""));
      const now = /* @__PURE__ */ new Date();
      const sameDay = d.toDateString() === now.toDateString();
      if (sameDay) {
        return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
      }
      const diff = (now.getTime() - d.getTime()) / 864e5;
      if (diff < 2)
        return "昨天";
      if (diff < 7)
        return `${Math.floor(diff)}天前`;
      return `${d.getMonth() + 1}/${d.getDate()}`;
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.unref(store).unread.liked_me
      }, common_vendor.unref(store).unread.liked_me ? {
        b: common_vendor.t(common_vendor.unref(store).unread.liked_me)
      } : {}, {
        c: common_vendor.o(($event) => goInteraction(0), "fe"),
        d: common_vendor.o(($event) => goInteraction(1), "57"),
        e: common_vendor.unref(store).unread.viewed_me
      }, common_vendor.unref(store).unread.viewed_me ? {
        f: common_vendor.t(common_vendor.unref(store).unread.viewed_me)
      } : {}, {
        g: common_vendor.o(($event) => goInteraction(2), "b2"),
        h: !common_vendor.unref(store).canInteract
      }, !common_vendor.unref(store).canInteract ? {} : list.value.length ? {
        j: common_vendor.f(list.value, (c, k0, i0) => {
          return common_vendor.e({
            a: c.peer_avatar,
            b: c.peer_online
          }, c.peer_online ? {} : {}, {
            c: c.unread
          }, c.unread ? {
            d: common_vendor.t(c.unread > 99 ? "99+" : c.unread)
          } : {}, {
            e: common_vendor.t(c.peer_nickname),
            f: common_vendor.t(timeText(c.last_message_at)),
            g: common_vendor.t(c.last_message || "你们已配对，打个招呼吧"),
            h: c.id,
            i: common_vendor.o(($event) => goChat(c), c.id)
          });
        })
      } : loaded.value ? {} : {}, {
        i: list.value.length,
        k: loaded.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-bdafde18"]]);
wx.createPage(MiniProgramPage);
