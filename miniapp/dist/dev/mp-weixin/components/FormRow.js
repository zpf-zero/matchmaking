"use strict";
const common_vendor = require("../common/vendor.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "FormRow",
  props: {
    label: {},
    required: { type: Boolean },
    alignRight: { type: Boolean }
  },
  setup(__props) {
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.t(_ctx.label),
        b: _ctx.required
      }, _ctx.required ? {} : {}, {
        c: _ctx.alignRight ? 1 : ""
      });
    };
  }
});
const Component = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-171d1c53"]]);
wx.createComponent(Component);
