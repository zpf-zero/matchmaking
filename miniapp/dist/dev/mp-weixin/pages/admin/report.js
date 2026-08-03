"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "report",
  setup(__props) {
    const store = stores_user.useUserStore();
    const list = common_vendor.ref([]);
    const page = common_vendor.ref(1);
    const finished = common_vendor.ref(false);
    const loading = common_vendor.ref(false);
    const loaded = common_vendor.ref(false);
    const filterStatus = common_vendor.ref(0);
    const statusText = ["待处理", "已处理", "已驳回"];
    common_vendor.onShow(() => load(true));
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
        const res = await api_index.adminApi.reportList(filterStatus.value, page.value);
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
    function switchFilter(s) {
      filterStatus.value = s;
      load(true);
    }
    function handle(item) {
      const canBan = store.role >= stores_user.Role.Admin;
      const items = canBan ? ["忽略（无问题）", "封禁被举报用户"] : ["忽略（无问题）"];
      common_vendor.index.showActionSheet({
        itemList: items,
        success: (res) => {
          if (res.tapIndex === 0)
            doHandle(item, "ignore", "核实后未发现问题");
          else
            confirmBan(item);
        }
      });
    }
    function confirmBan(item) {
      common_vendor.index.showModal({
        title: "封禁原因",
        editable: true,
        placeholderText: "将展示给被封禁用户",
        success: (r) => {
          var _a;
          if (r.confirm)
            doHandle(item, "ban", ((_a = r.content) == null ? void 0 : _a.trim()) || "违反社区规范");
        }
      });
    }
    async function doHandle(item, action, remark) {
      try {
        await api_index.adminApi.handleReport(item.report.id, action, remark);
        common_vendor.index.showToast({ title: "已处理", icon: "none" });
        load(true);
      } catch {
      }
    }
    function goUser(id) {
      common_vendor.index.navigateTo({ url: `/pages/admin/user-detail?id=${id}` });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: filterStatus.value === 0 ? 1 : "",
        b: common_vendor.o(($event) => switchFilter(0), "e5"),
        c: filterStatus.value === -1 ? 1 : "",
        d: common_vendor.o(($event) => switchFilter(-1), "72"),
        e: common_vendor.f(list.value, (item, k0, i0) => {
          return common_vendor.e({
            a: common_vendor.t(item.reason_text),
            b: common_vendor.t(statusText[item.report.status]),
            c: common_vendor.n("s" + item.report.status),
            d: common_vendor.t(item.from_nickname),
            e: common_vendor.t(item.report.from_user_id),
            f: common_vendor.t(item.target_nickname),
            g: common_vendor.t(item.report.target_user_id),
            h: common_vendor.o(($event) => goUser(item.report.target_user_id), item.report.id),
            i: item.report.description
          }, item.report.description ? {
            j: common_vendor.t(item.report.description)
          } : {}, {
            k: item.report.handle_remark
          }, item.report.handle_remark ? {
            l: common_vendor.t(item.report.handle_remark)
          } : {}, {
            m: item.report.status === 0
          }, item.report.status === 0 ? {
            n: common_vendor.o(($event) => handle(item), item.report.id)
          } : {}, {
            o: item.report.id
          });
        }),
        f: loaded.value && !list.value.length
      }, loaded.value && !list.value.length ? {} : list.value.length && finished.value ? {} : {}, {
        g: list.value.length && finished.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-55be1cc4"]]);
wx.createPage(MiniProgramPage);
