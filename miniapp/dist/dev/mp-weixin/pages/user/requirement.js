"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
if (!Math) {
  FormRow();
}
const FormRow = () => "../../components/FormRow.js";
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "requirement",
  setup(__props) {
    const store = stores_user.useUserStore();
    const saving = common_vendor.ref(false);
    const heightRange = Array.from({ length: 81 }, (_, i) => 140 + i);
    const ageRange = Array.from({ length: 63 }, (_, i) => 18 + i);
    const req = common_vendor.reactive({
      age_min: 22,
      age_max: 40,
      height_min: 150,
      height_max: 195,
      education_min: 0,
      income_min: 0,
      marital_status: "1",
      accept_children: 1,
      city_codes: "",
      other: ""
    });
    const acceptMarital = common_vendor.reactive({ 1: true, 2: false, 3: false });
    const cityName = common_vendor.ref("");
    common_vendor.onLoad(async () => {
      await store.fetchDict();
      if (!store.requirement)
        await store.fetchMe();
      const r = store.requirement;
      if (r) {
        Object.assign(req, r);
        Object.keys(acceptMarital).forEach((k) => acceptMarital[+k] = false);
        (r.marital_status || "").split(",").forEach((v) => {
          if (v)
            acceptMarital[+v] = true;
        });
      }
    });
    function pickNum(range, field) {
      common_vendor.index.showActionSheet({
        itemList: range.map(String),
        success: (res) => req[field] = range[res.tapIndex]
      });
    }
    function pickDict(type, field) {
      const list = [{ value: 0, label: "不限" }, ...store.dict[type] || []];
      common_vendor.index.showActionSheet({
        itemList: list.map((i) => i.label),
        success: (res) => req[field] = list[res.tapIndex].value
      });
    }
    function onRegion(e) {
      const [prov, city] = e.detail.value;
      cityName.value = `${prov} ${city}`;
      req.city_codes = (e.detail.code || [])[1] || "";
    }
    function clearCity() {
      cityName.value = "";
      req.city_codes = "";
    }
    async function save() {
      if (req.age_min > req.age_max) {
        common_vendor.index.showToast({ title: "年龄范围填反了", icon: "none" });
        return;
      }
      if (req.height_min > req.height_max) {
        common_vendor.index.showToast({ title: "身高范围填反了", icon: "none" });
        return;
      }
      saving.value = true;
      req.marital_status = Object.keys(acceptMarital).filter((k) => acceptMarital[+k]).join(",");
      try {
        await api_index.userApi.saveRequirement(req);
        await store.fetchMe();
        common_vendor.index.showToast({ title: "已保存" });
        setTimeout(() => common_vendor.index.navigateBack(), 700);
      } catch {
      } finally {
        saving.value = false;
      }
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.t(req.age_min),
        b: common_vendor.o(($event) => pickNum(common_vendor.unref(ageRange), "age_min"), "ed"),
        c: common_vendor.t(req.age_max),
        d: common_vendor.o(($event) => pickNum(common_vendor.unref(ageRange), "age_max"), "da"),
        e: common_vendor.p({
          label: "年龄范围",
          ["align-right"]: true
        }),
        f: common_vendor.t(req.height_min),
        g: common_vendor.o(($event) => pickNum(common_vendor.unref(heightRange), "height_min"), "01"),
        h: common_vendor.t(req.height_max),
        i: common_vendor.o(($event) => pickNum(common_vendor.unref(heightRange), "height_max"), "04"),
        j: common_vendor.p({
          label: "身高范围",
          ["align-right"]: true
        }),
        k: common_vendor.t(req.education_min ? common_vendor.unref(store).label("education", req.education_min) + "及以上" : "不限"),
        l: common_vendor.o(($event) => pickDict("education", "education_min"), "97"),
        m: common_vendor.p({
          label: "学历要求",
          ["align-right"]: true
        }),
        n: common_vendor.t(req.income_min ? common_vendor.unref(store).label("income_level", req.income_min) + "及以上" : "不限"),
        o: common_vendor.o(($event) => pickDict("income_level", "income_min"), "02"),
        p: common_vendor.p({
          label: "收入要求",
          ["align-right"]: true
        }),
        q: common_vendor.f(common_vendor.unref(store).dict["marital_status"] || [], (m, k0, i0) => {
          return {
            a: common_vendor.t(m.label),
            b: m.value,
            c: acceptMarital[m.value] ? 1 : "",
            d: common_vendor.o(($event) => acceptMarital[m.value] = !acceptMarital[m.value], m.value)
          };
        }),
        r: common_vendor.p({
          label: "可接受婚史",
          ["align-right"]: true
        }),
        s: req.accept_children === 1,
        t: common_vendor.o(($event) => req.accept_children = $event.detail.value ? 1 : 0, "f5"),
        v: common_vendor.p({
          label: "接受有子女",
          ["align-right"]: true
        }),
        w: common_vendor.t(cityName.value || (req.city_codes ? "已设置" : "不限")),
        x: common_vendor.o(onRegion, "1a"),
        y: req.city_codes
      }, req.city_codes ? {
        z: common_vendor.o(clearCity, "c9")
      } : {}, {
        A: common_vendor.p({
          label: "期望城市",
          ["align-right"]: true
        }),
        B: req.other,
        C: common_vendor.o(($event) => req.other = $event.detail.value, "47"),
        D: saving.value ? 1 : "",
        E: common_vendor.o(save, "cd")
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-7427c472"]]);
wx.createPage(MiniProgramPage);
