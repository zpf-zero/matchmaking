"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const utils_request = require("../../utils/request.js");
if (!Math) {
  FormRow();
}
const FormRow = () => "../../components/FormRow.js";
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "setup",
  setup(__props) {
    const store = stores_user.useUserStore();
    const isEdit = common_vendor.ref(false);
    const step = common_vendor.ref(1);
    const submitting = common_vendor.ref(false);
    const form = common_vendor.reactive({
      phone: "",
      nickname: "",
      avatar: "",
      gender: 0,
      birthday: "",
      height: 0,
      weight: 0,
      marital_status: 0,
      has_children: 0,
      education: 0,
      school: "",
      occupation: "",
      income_level: 0,
      house_status: 0,
      car_status: 0,
      province: "",
      city: "",
      city_code: "",
      home_province: "",
      home_city: "",
      self_intro: ""
    });
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
    const heightRange = Array.from({ length: 81 }, (_, i) => 140 + i);
    const ageRange = Array.from({ length: 63 }, (_, i) => 18 + i);
    const maxBirthday = common_vendor.computed(() => {
      const d = /* @__PURE__ */ new Date();
      d.setFullYear(d.getFullYear() - 18);
      return d.toISOString().slice(0, 10);
    });
    common_vendor.onLoad(async (query) => {
      isEdit.value = (query == null ? void 0 : query.mode) === "edit";
      await store.fetchDict();
      if (!store.user)
        await store.fetchMe();
      const u = store.user;
      if (u) {
        form.nickname = u.nickname;
        form.avatar = u.avatar;
        form.gender = u.gender;
      }
      form.phone = store.phone;
      const p = store.profile;
      if (p) {
        Object.assign(form, {
          birthday: p.birthday ? String(p.birthday).slice(0, 10) : "",
          height: p.height,
          weight: p.weight,
          marital_status: p.marital_status,
          has_children: p.has_children,
          education: p.education,
          school: p.school,
          occupation: p.occupation,
          income_level: p.income_level,
          house_status: p.house_status,
          car_status: p.car_status,
          province: p.province,
          city: p.city,
          city_code: p.city_code,
          home_province: p.home_province,
          home_city: p.home_city,
          self_intro: p.self_intro
        });
      }
      const r = store.requirement;
      if (r) {
        Object.assign(req, r);
        Object.keys(acceptMarital).forEach((k) => acceptMarital[+k] = false);
        (r.marital_status || "").split(",").forEach((v) => {
          if (v)
            acceptMarital[+v] = true;
        });
      }
      if (isEdit.value)
        step.value = 1;
    });
    function pickDict(type, field) {
      const list = store.dict[type] || [];
      common_vendor.index.showActionSheet({
        itemList: list.map((i) => i.label),
        success: (res) => {
          form[field] = list[res.tapIndex].value;
        }
      });
    }
    function pickReqDict(type, field) {
      const list = [{ value: 0, label: "不限" }, ...store.dict[type] || []];
      common_vendor.index.showActionSheet({
        itemList: list.map((i) => i.label),
        success: (res) => req[field] = list[res.tapIndex].value
      });
    }
    function onRegionChange(e, target) {
      const [prov, city] = e.detail.value;
      const codes = e.detail.code || [];
      if (target === "live") {
        form.province = prov;
        form.city = city;
        form.city_code = codes[1] || "";
      } else {
        form.home_province = prov;
        form.home_city = city;
      }
    }
    async function chooseAvatar() {
      common_vendor.index.chooseImage({
        count: 1,
        sizeType: ["compressed"],
        success: async (res) => {
          const path = res.tempFilePaths[0];
          common_vendor.index.showLoading({ title: "上传中" });
          try {
            form.avatar = await utils_request.uploadImage(path, "avatar");
          } finally {
            common_vendor.index.hideLoading();
          }
        }
      });
    }
    function validateStep(n) {
      if (n === 1) {
        if (!form.phone)
          return "请填写手机号";
        if (!/^1[3-9]\d{9}$/.test(form.phone))
          return "手机号格式不正确";
        if (!form.avatar)
          return "请上传头像";
        if (form.nickname.length < 2)
          return "昵称至少 2 个字";
        if (!form.gender)
          return "请选择性别";
        if (!form.birthday)
          return "请选择出生日期";
      }
      if (n === 2) {
        if (!form.height)
          return "请选择身高";
        if (!form.marital_status)
          return "请选择婚姻状况";
        if (!form.province)
          return "请选择现居城市";
        if (!form.home_province)
          return "请选择户籍地";
      }
      if (n === 3) {
        if (!form.education)
          return "请选择学历";
        if (!form.occupation)
          return "请填写职业";
        if (!form.income_level)
          return "请选择月收入";
        if (!form.house_status)
          return "请选择购房情况";
        if (!form.car_status)
          return "请选择购车情况";
      }
      if (n === 4) {
        if (form.self_intro.trim().length < 30)
          return "自我介绍至少 30 字";
      }
      return "";
    }
    async function next() {
      const err = validateStep(step.value);
      if (err) {
        common_vendor.index.showToast({ title: err, icon: "none" });
        return;
      }
      if (step.value === 1 && form.phone !== store.phone) {
        try {
          await api_index.authApi.bindPhone({ phone: form.phone });
          store.phone = form.phone;
        } catch {
          return;
        }
      }
      if (step.value < 4) {
        step.value++;
        common_vendor.index.pageScrollTo({ scrollTop: 0, duration: 0 });
        return;
      }
      await submit();
    }
    async function submit() {
      if (submitting.value)
        return;
      submitting.value = true;
      common_vendor.index.showLoading({ title: "提交中" });
      try {
        req.marital_status = Object.keys(acceptMarital).filter((k) => acceptMarital[+k]).join(",");
        await api_index.userApi.saveProfile(form);
        await api_index.userApi.saveRequirement(req);
        if (!isEdit.value) {
          await api_index.userApi.submitAudit();
        }
        await store.fetchMe();
        common_vendor.index.hideLoading();
        common_vendor.index.showModal({
          title: isEdit.value ? "已保存" : "提交成功",
          content: isEdit.value ? "资料修改后需要重新审核，期间不影响正常使用" : "资料已提交审核，通常 24 小时内完成",
          showCancel: false,
          success: () => common_vendor.index.reLaunch({ url: "/pages/index/index" })
        });
      } catch {
        common_vendor.index.hideLoading();
      } finally {
        submitting.value = false;
      }
    }
    function prev() {
      if (step.value > 1)
        step.value--;
    }
    return (_ctx, _cache) => {
      var _a, _b;
      return common_vendor.e({
        a: common_vendor.f(4, (n, k0, i0) => {
          return common_vendor.e({
            a: common_vendor.t(n),
            b: n < 4
          }, n < 4 ? {} : {}, {
            c: n,
            d: n <= step.value ? 1 : ""
          });
        }),
        b: common_vendor.t(["基础信息", "个人情况", "硬性条件", "介绍与期望"][step.value - 1]),
        c: common_vendor.t(step.value),
        d: form.phone,
        e: common_vendor.o(($event) => form.phone = $event.detail.value, "50"),
        f: common_vendor.p({
          label: "手机号",
          required: true
        }),
        g: form.avatar
      }, form.avatar ? {
        h: form.avatar
      } : {}, {
        i: common_vendor.o(chooseAvatar, "63"),
        j: common_vendor.p({
          label: "头像",
          required: true,
          ["align-right"]: true
        }),
        k: form.nickname,
        l: common_vendor.o(($event) => form.nickname = $event.detail.value, "10"),
        m: common_vendor.p({
          label: "昵称",
          required: true
        }),
        n: (_a = common_vendor.unref(store).user) == null ? void 0 : _a.gender
      }, ((_b = common_vendor.unref(store).user) == null ? void 0 : _b.gender) ? {
        o: common_vendor.t(common_vendor.unref(store).label("gender", form.gender))
      } : {
        p: common_vendor.f([{
          v: 1,
          t: "男"
        }, {
          v: 2,
          t: "女"
        }], (g, k0, i0) => {
          return {
            a: common_vendor.t(g.t),
            b: g.v,
            c: form.gender === g.v ? 1 : "",
            d: common_vendor.o(($event) => form.gender = g.v, g.v)
          };
        })
      }, {
        q: common_vendor.p({
          label: "性别",
          required: true,
          ["align-right"]: true
        }),
        r: common_vendor.t(form.birthday || "请选择"),
        s: !form.birthday ? 1 : "",
        t: maxBirthday.value,
        v: form.birthday,
        w: common_vendor.o(($event) => form.birthday = $event.detail.value, "b2"),
        x: common_vendor.p({
          label: "出生日期",
          required: true,
          ["align-right"]: true
        }),
        y: step.value === 1,
        z: common_vendor.t(form.height ? form.height + " cm" : "请选择"),
        A: !form.height ? 1 : "",
        B: common_vendor.unref(heightRange),
        C: common_vendor.o(($event) => form.height = common_vendor.unref(heightRange)[$event.detail.value], "7f"),
        D: common_vendor.p({
          label: "身高",
          required: true,
          ["align-right"]: true
        }),
        E: form.weight,
        F: common_vendor.o(common_vendor.m(($event) => form.weight = $event.detail.value, {
          number: true
        }), "22"),
        G: common_vendor.p({
          label: "体重",
          ["align-right"]: true
        }),
        H: common_vendor.t(common_vendor.unref(store).label("marital_status", form.marital_status) || "请选择"),
        I: !form.marital_status ? 1 : "",
        J: common_vendor.o(($event) => pickDict("marital_status", "marital_status"), "9e"),
        K: common_vendor.p({
          label: "婚姻状况",
          required: true,
          ["align-right"]: true
        }),
        L: common_vendor.t(common_vendor.unref(store).label("has_children", form.has_children) || "无"),
        M: common_vendor.o(($event) => pickDict("has_children", "has_children"), "42"),
        N: common_vendor.p({
          label: "子女情况",
          required: true,
          ["align-right"]: true
        }),
        O: common_vendor.t(form.province ? form.province + " " + form.city : "请选择"),
        P: !form.province ? 1 : "",
        Q: common_vendor.o(($event) => onRegionChange($event, "live"), "28"),
        R: common_vendor.p({
          label: "现居城市",
          required: true,
          ["align-right"]: true
        }),
        S: common_vendor.t(form.home_province ? form.home_province + " " + form.home_city : "请选择"),
        T: !form.home_province ? 1 : "",
        U: common_vendor.o(($event) => onRegionChange($event, "home"), "20"),
        V: common_vendor.p({
          label: "户籍地",
          required: true,
          ["align-right"]: true
        }),
        W: step.value === 2,
        X: common_vendor.t(common_vendor.unref(store).label("education", form.education) || "请选择"),
        Y: !form.education ? 1 : "",
        Z: common_vendor.o(($event) => pickDict("education", "education"), "ec"),
        aa: common_vendor.p({
          label: "学历",
          required: true,
          ["align-right"]: true
        }),
        ab: form.school,
        ac: common_vendor.o(($event) => form.school = $event.detail.value, "81"),
        ad: common_vendor.p({
          label: "毕业院校",
          ["align-right"]: true
        }),
        ae: form.occupation,
        af: common_vendor.o(($event) => form.occupation = $event.detail.value, "d3"),
        ag: common_vendor.p({
          label: "职业",
          required: true,
          ["align-right"]: true
        }),
        ah: common_vendor.t(common_vendor.unref(store).label("income_level", form.income_level) || "请选择"),
        ai: !form.income_level ? 1 : "",
        aj: common_vendor.o(($event) => pickDict("income_level", "income_level"), "b4"),
        ak: common_vendor.p({
          label: "月收入",
          required: true,
          ["align-right"]: true
        }),
        al: common_vendor.t(common_vendor.unref(store).label("house_status", form.house_status) || "请选择"),
        am: !form.house_status ? 1 : "",
        an: common_vendor.o(($event) => pickDict("house_status", "house_status"), "79"),
        ao: common_vendor.p({
          label: "购房情况",
          required: true,
          ["align-right"]: true
        }),
        ap: common_vendor.t(common_vendor.unref(store).label("car_status", form.car_status) || "请选择"),
        aq: !form.car_status ? 1 : "",
        ar: common_vendor.o(($event) => pickDict("car_status", "car_status"), "32"),
        as: common_vendor.p({
          label: "购车情况",
          required: true,
          ["align-right"]: true
        }),
        at: step.value === 3,
        av: form.self_intro,
        aw: common_vendor.o(($event) => form.self_intro = $event.detail.value, "c1"),
        ax: common_vendor.t(form.self_intro.length),
        ay: form.self_intro.length < 30 ? 1 : "",
        az: common_vendor.t(req.age_min),
        aA: common_vendor.unref(ageRange),
        aB: common_vendor.o(($event) => req.age_min = common_vendor.unref(ageRange)[$event.detail.value], "49"),
        aC: common_vendor.t(req.age_max),
        aD: common_vendor.unref(ageRange),
        aE: common_vendor.o(($event) => req.age_max = common_vendor.unref(ageRange)[$event.detail.value], "f6"),
        aF: common_vendor.p({
          label: "年龄范围",
          ["align-right"]: true
        }),
        aG: common_vendor.t(req.height_min),
        aH: common_vendor.unref(heightRange),
        aI: common_vendor.o(($event) => req.height_min = common_vendor.unref(heightRange)[$event.detail.value], "a2"),
        aJ: common_vendor.t(req.height_max),
        aK: common_vendor.unref(heightRange),
        aL: common_vendor.o(($event) => req.height_max = common_vendor.unref(heightRange)[$event.detail.value], "0d"),
        aM: common_vendor.p({
          label: "身高范围",
          ["align-right"]: true
        }),
        aN: common_vendor.t(req.education_min ? common_vendor.unref(store).label("education", req.education_min) + "及以上" : "不限"),
        aO: common_vendor.o(($event) => pickReqDict("education", "education_min"), "ec"),
        aP: common_vendor.p({
          label: "学历要求",
          ["align-right"]: true
        }),
        aQ: common_vendor.t(req.income_min ? common_vendor.unref(store).label("income_level", req.income_min) + "及以上" : "不限"),
        aR: common_vendor.o(($event) => pickReqDict("income_level", "income_min"), "16"),
        aS: common_vendor.p({
          label: "收入要求",
          ["align-right"]: true
        }),
        aT: common_vendor.f(common_vendor.unref(store).dict["marital_status"] || [], (m, k0, i0) => {
          return {
            a: common_vendor.t(m.label),
            b: m.value,
            c: acceptMarital[m.value] ? 1 : "",
            d: common_vendor.o(($event) => acceptMarital[m.value] = !acceptMarital[m.value], m.value)
          };
        }),
        aU: common_vendor.p({
          label: "可接受婚史",
          ["align-right"]: true
        }),
        aV: req.accept_children === 1,
        aW: common_vendor.o(($event) => req.accept_children = $event.detail.value ? 1 : 0, "6e"),
        aX: common_vendor.p({
          label: "接受有子女",
          ["align-right"]: true
        }),
        aY: step.value === 4,
        aZ: step.value > 1
      }, step.value > 1 ? {
        ba: common_vendor.o(prev, "59")
      } : {}, {
        bb: common_vendor.t(step.value < 4 ? "下一步" : isEdit.value ? "保存" : "提交审核"),
        bc: step.value > 1 ? 1 : "",
        bd: submitting.value ? 1 : "",
        be: common_vendor.o(next, "a3")
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-15137599"]]);
wx.createPage(MiniProgramPage);
