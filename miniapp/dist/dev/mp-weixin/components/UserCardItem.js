"use strict";
const common_vendor = require("../common/vendor.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "UserCardItem",
  props: {
    user: {},
    showLike: { type: Boolean },
    liked: { type: Boolean }
  },
  emits: ["like"],
  setup(__props, { emit: __emit }) {
    const props = __props;
    const emit = __emit;
    function goDetail() {
      common_vendor.index.navigateTo({ url: `/pages/user/detail?id=${props.user.id}` });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: _ctx.user.avatar,
        b: common_vendor.t(_ctx.user.nickname),
        c: _ctx.user.match_rate > 0
      }, _ctx.user.match_rate > 0 ? {
        d: common_vendor.t(_ctx.user.match_rate)
      } : {}, {
        e: common_vendor.t(_ctx.user.age),
        f: common_vendor.t(_ctx.user.height),
        g: common_vendor.t(_ctx.user.education_text),
        h: common_vendor.t(_ctx.user.income_text),
        i: common_vendor.t(_ctx.user.city),
        j: common_vendor.t(_ctx.user.marital_text),
        k: _ctx.user.occupation
      }, _ctx.user.occupation ? {} : {}, {
        l: _ctx.user.occupation
      }, _ctx.user.occupation ? {
        m: common_vendor.t(_ctx.user.occupation)
      } : {}, {
        n: common_vendor.t(_ctx.user.active_text),
        o: _ctx.showLike
      }, _ctx.showLike ? {
        p: common_vendor.t(_ctx.liked ? "♥" : "♡"),
        q: _ctx.liked ? 1 : "",
        r: common_vendor.o(($event) => emit("like", _ctx.user), "3f")
      } : {}, {
        s: common_vendor.o(goDetail, "ef")
      });
    };
  }
});
const Component = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-c201bdda"]]);
wx.createComponent(Component);
