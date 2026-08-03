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
    const total = common_vendor.ref(0);
    const page = common_vendor.ref(1);
    const loading = common_vendor.ref(false);
    const finished = common_vendor.ref(false);
    const searched = common_vendor.ref(false);
    const panelOpen = common_vendor.ref(true);
    const heightRange = [0, ...Array.from({ length: 81 }, (_, i) => 140 + i)];
    const ageRange = [0, ...Array.from({ length: 63 }, (_, i) => 18 + i)];
    const filters = common_vendor.reactive({
      age_min: 0,
      age_max: 0,
      height_min: 0,
      height_max: 0,
      education_min: 0,
      income_min: 0,
      marital_status: "",
      has_children: "",
      city_code: "",
      city_name: "",
      house_status: 0,
      car_status: 0
    });
    common_vendor.onShow(() => {
      store.fetchDict();
    });
    function pickNumber(range, field) {
      common_vendor.index.showActionSheet({
        itemList: range.map((v) => v === 0 ? "不限" : String(v)),
        success: (res) => filters[field] = range[res.tapIndex]
      });
    }
    function pickDict(type, field) {
      const list2 = [{ value: 0, label: "不限" }, ...store.dict[type] || []];
      common_vendor.index.showActionSheet({
        itemList: list2.map((i) => i.label),
        success: (res) => filters[field] = list2[res.tapIndex].value
      });
    }
    function pickMarital() {
      const items = [{ value: "", label: "不限" }, ...(store.dict["marital_status"] || []).map((i) => ({ value: String(i.value), label: i.label }))];
      common_vendor.index.showActionSheet({
        itemList: items.map((i) => i.label),
        success: (res) => filters.marital_status = items[res.tapIndex].value
      });
    }
    function pickChildren() {
      const items = [{ value: "", label: "不限" }, ...(store.dict["has_children"] || []).map((i) => ({ value: String(i.value), label: i.label }))];
      common_vendor.index.showActionSheet({
        itemList: items.map((i) => i.label),
        success: (res) => filters.has_children = items[res.tapIndex].value
      });
    }
    function onRegion(e) {
      const [prov, city] = e.detail.value;
      filters.city_name = `${prov} ${city}`;
      filters.city_code = (e.detail.code || [])[1] || "";
    }
    function reset() {
      Object.assign(filters, {
        age_min: 0,
        age_max: 0,
        height_min: 0,
        height_max: 0,
        education_min: 0,
        income_min: 0,
        marital_status: "",
        has_children: "",
        city_code: "",
        city_name: "",
        house_status: 0,
        car_status: 0
      });
    }
    async function doSearch(resetPage = true) {
      if (loading.value)
        return;
      loading.value = true;
      if (resetPage) {
        page.value = 1;
        finished.value = false;
        panelOpen.value = false;
      }
      const params = { page: page.value, page_size: 20 };
      Object.entries(filters).forEach(([k, v]) => {
        if (k === "city_name")
          return;
        if (v !== 0 && v !== "")
          params[k] = v;
      });
      try {
        const res = await api_index.discoverApi.search(params);
        total.value = res.total;
        list.value = resetPage ? res.list || [] : [...list.value, ...res.list || []];
        finished.value = list.value.length >= res.total;
        if (!finished.value)
          page.value++;
        searched.value = true;
      } catch {
      } finally {
        loading.value = false;
      }
    }
    common_vendor.onReachBottom(() => {
      if (searched.value && !finished.value)
        doSearch(false);
    });
    function labelOr(type, value, fallback = "不限") {
      return value ? store.label(type, value) : fallback;
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.t(panelOpen.value ? "收起 ∧" : "展开 ∨"),
        b: common_vendor.o(($event) => panelOpen.value = !panelOpen.value, "d8"),
        c: common_vendor.t(filters.age_min || "不限"),
        d: common_vendor.o(($event) => pickNumber(ageRange, "age_min"), "a6"),
        e: common_vendor.t(filters.age_max || "不限"),
        f: common_vendor.o(($event) => pickNumber(ageRange, "age_max"), "b8"),
        g: common_vendor.t(filters.height_min || "不限"),
        h: common_vendor.o(($event) => pickNumber(heightRange, "height_min"), "d9"),
        i: common_vendor.t(filters.height_max || "不限"),
        j: common_vendor.o(($event) => pickNumber(heightRange, "height_max"), "cd"),
        k: common_vendor.t(labelOr("education", filters.education_min)),
        l: common_vendor.o(($event) => pickDict("education", "education_min"), "4e"),
        m: common_vendor.t(labelOr("income_level", filters.income_min)),
        n: common_vendor.o(($event) => pickDict("income_level", "income_min"), "dd"),
        o: common_vendor.t(filters.marital_status ? common_vendor.unref(store).label("marital_status", +filters.marital_status) : "不限"),
        p: common_vendor.o(pickMarital, "b2"),
        q: common_vendor.t(filters.has_children !== "" ? common_vendor.unref(store).label("has_children", +filters.has_children) : "不限"),
        r: common_vendor.o(pickChildren, "e6"),
        s: common_vendor.t(filters.city_name || "不限"),
        t: common_vendor.o(onRegion, "2b"),
        v: common_vendor.t(labelOr("house_status", filters.house_status)),
        w: common_vendor.o(($event) => pickDict("house_status", "house_status"), "a8"),
        x: common_vendor.t(labelOr("car_status", filters.car_status)),
        y: common_vendor.o(($event) => pickDict("car_status", "car_status"), "ed"),
        z: common_vendor.o(reset, "cf"),
        A: common_vendor.o(($event) => doSearch(true), "2f"),
        B: panelOpen.value,
        C: searched.value
      }, searched.value ? {
        D: common_vendor.t(total.value)
      } : {}, {
        E: common_vendor.f(list.value, (u, k0, i0) => {
          return {
            a: u.id,
            b: "308a4d57-0-" + i0,
            c: common_vendor.p({
              user: u
            })
          };
        }),
        F: searched.value && !list.value.length
      }, searched.value && !list.value.length ? {} : list.value.length && finished.value ? {} : {}, {
        G: list.value.length && finished.value
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-308a4d57"]]);
wx.createPage(MiniProgramPage);
