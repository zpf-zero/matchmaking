"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const utils_request = require("../../utils/request.js");
const utils_ws = require("../../utils/ws.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "index",
  setup(__props) {
    const store = stores_user.useUserStore();
    const convId = common_vendor.ref(0);
    const peerName = common_vendor.ref("");
    const messages = common_vendor.ref([]);
    const input = common_vendor.ref("");
    const scrollTop = common_vendor.ref(0);
    const loadingMore = common_vendor.ref(false);
    const noMore = common_vendor.ref(false);
    const sending = common_vendor.ref(false);
    const myId = () => {
      var _a;
      return ((_a = store.user) == null ? void 0 : _a.id) || 0;
    };
    common_vendor.onLoad(async (q) => {
      convId.value = Number((q == null ? void 0 : q.id) || 0);
      peerName.value = decodeURIComponent((q == null ? void 0 : q.name) || "聊天");
      common_vendor.index.setNavigationBarTitle({ title: peerName.value });
      await loadHistory();
      await markRead();
      utils_ws.ws.on("message", onWsMessage);
      utils_ws.ws.on("ack", onWsAck);
      utils_ws.ws.on("error", onWsError);
      utils_ws.ws.connect();
    });
    common_vendor.onUnload(() => {
      utils_ws.ws.off("message", onWsMessage);
      utils_ws.ws.off("ack", onWsAck);
      utils_ws.ws.off("error", onWsError);
    });
    async function loadHistory(before = 0) {
      if (loadingMore.value || noMore.value)
        return;
      loadingMore.value = true;
      try {
        const list = await api_index.chatApi.messages(convId.value, before);
        if (!list.length) {
          noMore.value = true;
          return;
        }
        messages.value = before ? [...list, ...messages.value] : list;
        if (!before)
          await scrollToBottom();
      } catch {
      } finally {
        loadingMore.value = false;
      }
    }
    function loadMore() {
      if (messages.value.length)
        loadHistory(messages.value[0].id);
    }
    async function markRead() {
      try {
        await api_index.chatApi.read(convId.value);
        store.fetchUnread();
      } catch {
      }
    }
    function onWsMessage(msg) {
      if (msg.conversation_id !== convId.value)
        return;
      messages.value.push(msg);
      scrollToBottom();
      markRead();
    }
    function onWsAck(data) {
      const idx = messages.value.findIndex((m) => m._temp_id === data.temp_id);
      if (idx >= 0) {
        messages.value[idx] = data.message;
      } else {
        messages.value.push(data.message);
      }
      scrollToBottom();
    }
    function onWsError(data) {
      if (data.temp_id) {
        const m = messages.value.find((x) => x._temp_id === data.temp_id);
        if (m) {
          m._sending = false;
          m._failed = true;
        }
      }
      common_vendor.index.showToast({ title: data.message, icon: "none" });
    }
    async function send() {
      const content = input.value.trim();
      if (!content || sending.value)
        return;
      input.value = "";
      await doSend(content, 1);
    }
    function chooseImage() {
      common_vendor.index.chooseImage({
        count: 1,
        sizeType: ["compressed"],
        success: async (res) => {
          const path = res.tempFilePaths[0];
          common_vendor.index.showLoading({ title: "上传中" });
          try {
            const url = await utils_request.uploadImage(path, "chat");
            await doSend(url, 2);
          } catch {
          } finally {
            common_vendor.index.hideLoading();
          }
        }
      });
    }
    async function doSend(content, msgType) {
      if (sending.value)
        return;
      sending.value = true;
      const tempId = `t${Date.now()}`;
      const optimistic = {
        id: 0,
        conversation_id: convId.value,
        from_user_id: myId(),
        type: msgType,
        content,
        is_read: 0,
        created_at: (/* @__PURE__ */ new Date()).toISOString(),
        _temp_id: tempId,
        _sending: true
      };
      messages.value.push(optimistic);
      await scrollToBottom();
      const ok = utils_ws.ws.connected && utils_ws.ws.send("send", {
        conversation_id: convId.value,
        msg_type: msgType,
        content,
        temp_id: tempId
      });
      if (!ok) {
        try {
          const msg = await api_index.chatApi.send(convId.value, content, msgType);
          const idx = messages.value.findIndex((m) => m._temp_id === tempId);
          if (idx >= 0)
            messages.value[idx] = msg;
        } catch {
          const m = messages.value.find((x) => x._temp_id === tempId);
          if (m) {
            m._sending = false;
            m._failed = true;
          }
        }
      }
      sending.value = false;
    }
    function previewImage(url) {
      common_vendor.index.previewImage({ urls: [url] });
    }
    async function scrollToBottom() {
      await common_vendor.nextTick$1();
      scrollTop.value = 999999 + Math.random();
    }
    function isMine(m) {
      return m.from_user_id === myId();
    }
    function unmatch() {
      common_vendor.index.showModal({
        title: "解除配对",
        content: "解除后将无法继续聊天，且需要重新互相喜欢才能配对",
        success: async (r) => {
          if (!r.confirm)
            return;
          try {
            await api_index.chatApi.closeConversation(convId.value);
            common_vendor.index.showToast({ title: "已解除配对", icon: "none" });
            setTimeout(() => common_vendor.index.navigateBack(), 800);
          } catch {
          }
        }
      });
    }
    function timeText(t) {
      const d = new Date(t.replace(/-/g, "/").replace(/\..*$/, ""));
      return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: noMore.value
      }, noMore.value ? {} : {}, {
        b: common_vendor.f(messages.value, (m, k0, i0) => {
          return common_vendor.e({
            a: m.type === 2
          }, m.type === 2 ? {
            b: m.content,
            c: common_vendor.o(($event) => previewImage(m.content), m.id || m._temp_id)
          } : {
            d: common_vendor.t(m.content)
          }, {
            e: m._sending
          }, m._sending ? {} : m._failed ? {} : {
            g: common_vendor.t(timeText(m.created_at))
          }, {
            f: m._failed,
            h: m.id || m._temp_id,
            i: isMine(m) ? 1 : ""
          });
        }),
        c: scrollTop.value,
        d: common_vendor.o(loadMore, "b9"),
        e: common_vendor.o(send, "80"),
        f: input.value,
        g: common_vendor.o(($event) => input.value = $event.detail.value, "0c"),
        h: common_vendor.o(chooseImage, "a5"),
        i: input.value.trim() ? 1 : "",
        j: common_vendor.o(send, "90"),
        k: common_vendor.o(unmatch, "1d")
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-da04a0a0"]]);
wx.createPage(MiniProgramPage);
