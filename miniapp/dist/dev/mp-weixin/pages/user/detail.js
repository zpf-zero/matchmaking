"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "detail",
  setup(__props) {
    const store = stores_user.useUserStore();
    const detail = common_vendor.ref(null);
    const loading = common_vendor.ref(true);
    const userId = common_vendor.ref(0);
    common_vendor.onLoad(async (q) => {
      userId.value = Number((q == null ? void 0 : q.id) || 0);
      if (!userId.value) {
        common_vendor.index.showToast({ title: "参数错误", icon: "none" });
        return;
      }
      await store.fetchDict();
      await load();
    });
    async function load() {
      loading.value = true;
      try {
        detail.value = await api_index.userApi.detail(userId.value);
      } catch {
        setTimeout(() => common_vendor.index.navigateBack(), 1200);
      } finally {
        loading.value = false;
      }
    }
    function previewPhoto(index) {
      var _a, _b;
      if (!((_b = (_a = detail.value) == null ? void 0 : _a.photos) == null ? void 0 : _b.length))
        return;
      common_vendor.index.previewImage({ urls: detail.value.photos, current: index });
    }
    async function handleLike() {
      if (!store.canInteract) {
        common_vendor.index.showToast({ title: "资料审核通过后才能表达心意", icon: "none" });
        return;
      }
      const d = detail.value;
      try {
        const res = await api_index.interactionApi.like(d.id);
        d.is_liked = true;
        if (res.is_match) {
          d.is_matched = true;
          d.conversation_id = res.conversation_id;
          common_vendor.index.showModal({
            title: "配对成功",
            content: `你和 ${d.nickname} 互相喜欢，现在可以聊天了`,
            confirmText: "去聊天",
            cancelText: "稍后",
            success: (r) => r.confirm && goChat()
          });
        } else {
          common_vendor.index.showToast({ title: "已表达心意，等待对方回应", icon: "none" });
        }
      } catch {
      }
    }
    async function handlePass() {
      try {
        await api_index.interactionApi.pass(detail.value.id);
        common_vendor.index.showToast({ title: "已跳过", icon: "none" });
        setTimeout(() => common_vendor.index.navigateBack(), 600);
      } catch {
      }
    }
    function goChat() {
      const d = detail.value;
      common_vendor.index.navigateTo({
        url: `/pages/chat/index?id=${d.conversation_id}&name=${encodeURIComponent(d.nickname)}`
      });
    }
    function moreActions() {
      common_vendor.index.showActionSheet({
        itemList: ["举报该用户", "拉黑该用户"],
        success: (res) => {
          if (res.tapIndex === 0)
            doReport();
          else
            doBlock();
        }
      });
    }
    function doReport() {
      const reasons = store.dict["report_reason"] || [];
      common_vendor.index.showActionSheet({
        itemList: reasons.map((r) => r.label),
        success: async (res) => {
          try {
            await api_index.commonApi.report({
              target_user_id: detail.value.id,
              reason: reasons[res.tapIndex].value
            });
            common_vendor.index.showToast({ title: "举报已提交，我们会尽快核实", icon: "none" });
          } catch {
          }
        }
      });
    }
    function doBlock() {
      common_vendor.index.showModal({
        title: "确认拉黑",
        content: "拉黑后双方将互相不可见，已有的配对也会解除",
        success: async (r) => {
          if (!r.confirm)
            return;
          try {
            await api_index.commonApi.addBlacklist(detail.value.id);
            common_vendor.index.showToast({ title: "已拉黑", icon: "none" });
            setTimeout(() => common_vendor.index.navigateBack(), 800);
          } catch {
          }
        }
      });
    }
    return (_ctx, _cache) => {
      var _a, _b;
      return common_vendor.e({
        a: detail.value
      }, detail.value ? common_vendor.e({
        b: detail.value.avatar,
        c: common_vendor.o(($event) => previewPhoto(0), "45"),
        d: common_vendor.t(detail.value.nickname),
        e: common_vendor.t(detail.value.age),
        f: common_vendor.t(detail.value.height),
        g: common_vendor.t(detail.value.city),
        h: common_vendor.o(moreActions, "e3"),
        i: detail.value.match_rate > 0
      }, detail.value.match_rate > 0 ? {
        j: common_vendor.t(detail.value.match_rate)
      } : {}, {
        k: detail.value.liked_me && !detail.value.is_matched
      }, detail.value.liked_me && !detail.value.is_matched ? {} : {}, {
        l: common_vendor.t(detail.value.education_text),
        m: common_vendor.t(detail.value.income_text),
        n: common_vendor.t(detail.value.marital_text),
        o: common_vendor.t(common_vendor.unref(store).label("has_children", detail.value.has_children)),
        p: common_vendor.t(common_vendor.unref(store).label("house_status", detail.value.house_status)),
        q: common_vendor.t(common_vendor.unref(store).label("car_status", detail.value.car_status)),
        r: common_vendor.t(detail.value.occupation || "—"),
        s: common_vendor.t(detail.value.school || "—"),
        t: common_vendor.t(detail.value.province),
        v: common_vendor.t(detail.value.city),
        w: common_vendor.t(detail.value.home_province),
        x: common_vendor.t(detail.value.home_city),
        y: common_vendor.t(detail.value.self_intro),
        z: (_a = detail.value.photos) == null ? void 0 : _a.length
      }, ((_b = detail.value.photos) == null ? void 0 : _b.length) ? {
        A: common_vendor.f(detail.value.photos, (p, i, i0) => {
          return {
            a: i,
            b: p,
            c: common_vendor.o(($event) => previewPhoto(i), i)
          };
        })
      } : {}, {
        B: detail.value.is_matched
      }, detail.value.is_matched ? {
        C: common_vendor.o(goChat, "5e")
      } : {
        D: common_vendor.o(handlePass, "5b"),
        E: common_vendor.t(detail.value.is_liked ? "已喜欢" : "喜欢 TA"),
        F: detail.value.is_liked ? 1 : "",
        G: common_vendor.o(handleLike, "0f")
      }) : !loading.value ? {} : {}, {
        H: !loading.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-0b0b35bb"]]);
wx.createPage(MiniProgramPage);
