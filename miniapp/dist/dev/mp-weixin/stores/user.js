"use strict";
const common_vendor = require("../common/vendor.js");
const api_index = require("../api/index.js");
const utils_ws = require("../utils/ws.js");
const Status = {
  Incomplete: 0,
  Auditing: 1,
  Normal: 2,
  Rejected: 3,
  Banned: 4
};
const Role = {
  User: 1,
  Auditor: 2,
  Admin: 3,
  SuperAdmin: 4
};
const useUserStore = common_vendor.defineStore("user", {
  state: () => ({
    token: common_vendor.index.getStorageSync("token") || "",
    user: null,
    phone: "",
    profile: null,
    requirement: null,
    photos: [],
    dict: {},
    mockMode: true,
    unread: { liked_me: 0, viewed_me: 0, message: 0, total: 0 }
  }),
  getters: {
    isLogin: (s) => !!s.token,
    status: (s) => {
      var _a;
      return ((_a = s.user) == null ? void 0 : _a.status) ?? Status.Incomplete;
    },
    role: (s) => {
      var _a;
      return ((_a = s.user) == null ? void 0 : _a.role) ?? Role.User;
    },
    /** role > 1 时「我的」页才渲染管理入口。真正的拦截在后端 */
    isAdminRole: (s) => {
      var _a;
      return (((_a = s.user) == null ? void 0 : _a.role) ?? 1) > Role.User;
    },
    canManageUser: (s) => {
      var _a;
      return (((_a = s.user) == null ? void 0 : _a.role) ?? 1) >= Role.Admin;
    },
    /** 资料已提交，可以浏览他人 */
    canBrowse: (s) => {
      var _a;
      const st = ((_a = s.user) == null ? void 0 : _a.status) ?? 0;
      return st === Status.Auditing || st === Status.Normal || st === Status.Rejected;
    },
    /** 审核通过，可以发起互动 */
    canInteract: (s) => {
      var _a;
      return ((_a = s.user) == null ? void 0 : _a.status) === Status.Normal;
    }
  },
  actions: {
    async login(code) {
      const res = await api_index.authApi.wxLogin(code);
      this.token = res.token;
      common_vendor.index.setStorageSync("token", res.token);
      this.user = res.user;
      return res;
    },
    async fetchMe() {
      const me = await api_index.userApi.me();
      this.user = me.user;
      this.phone = me.phone;
      this.profile = me.profile;
      this.requirement = me.requirement;
      this.photos = me.photos || [];
      return me;
    },
    async fetchDict() {
      if (Object.keys(this.dict).length)
        return this.dict;
      this.dict = await api_index.commonApi.dict();
      return this.dict;
    },
    async fetchConfig() {
      try {
        const cfg = await api_index.commonApi.config();
        this.mockMode = cfg.mock_mode;
      } catch {
      }
    },
    async fetchUnread() {
      if (!this.canBrowse)
        return;
      try {
        this.unread = await (await "../api/index.js").interactionApi.unreadCount();
      } catch {
      }
    },
    /** 字典值转文字 */
    label(type, value) {
      var _a, _b;
      return ((_b = (_a = this.dict[type]) == null ? void 0 : _a.find((i) => i.value === value)) == null ? void 0 : _b.label) || "";
    },
    logout() {
      this.token = "";
      this.user = null;
      this.profile = null;
      this.requirement = null;
      this.photos = [];
      common_vendor.index.removeStorageSync("token");
      utils_ws.ws.close();
      common_vendor.index.reLaunch({ url: "/pages/login/index" });
    }
  }
});
exports.Role = Role;
exports.Status = Status;
exports.useUserStore = useUserStore;
