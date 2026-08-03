"use strict";
const common_vendor = require("../../common/vendor.js");
const api_index = require("../../api/index.js");
const stores_user = require("../../stores/user.js");
const utils_request = require("../../utils/request.js");
const MAX = 9;
const _sfc_main = /* @__PURE__ */ common_vendor.defineComponent({
  __name: "photos",
  setup(__props) {
    const store = stores_user.useUserStore();
    const uploading = common_vendor.ref(false);
    common_vendor.onShow(() => {
      store.fetchMe();
    });
    function auditText(s) {
      return ["审核中", "", "未通过"][s] || "";
    }
    async function add() {
      if (store.photos.length >= MAX) {
        common_vendor.index.showToast({ title: `最多 ${MAX} 张`, icon: "none" });
        return;
      }
      common_vendor.index.chooseImage({
        count: MAX - store.photos.length,
        sizeType: ["compressed"],
        success: async (res) => {
          const paths = res.tempFilePaths;
          uploading.value = true;
          common_vendor.index.showLoading({ title: "上传中" });
          try {
            for (const p of paths) {
              const url = await utils_request.uploadImage(p, "photo");
              await api_index.userApi.addPhoto(url);
            }
            await store.fetchMe();
          } catch {
          } finally {
            common_vendor.index.hideLoading();
            uploading.value = false;
          }
        }
      });
    }
    function remove(id) {
      common_vendor.index.showModal({
        title: "删除照片",
        content: "确定删除这张照片吗？",
        success: async (r) => {
          if (!r.confirm)
            return;
          try {
            await api_index.userApi.deletePhoto(id);
            await store.fetchMe();
          } catch {
          }
        }
      });
    }
    function preview(i) {
      common_vendor.index.previewImage({
        urls: store.photos.map((p) => p.url),
        current: i
      });
    }
    return (_ctx, _cache) => {
      return common_vendor.e({
        a: common_vendor.f(common_vendor.unref(store).photos, (p, i, i0) => {
          return common_vendor.e({
            a: p.url,
            b: common_vendor.o(($event) => preview(i), p.id),
            c: p.audit_status !== 1
          }, p.audit_status !== 1 ? {
            d: common_vendor.t(auditText(p.audit_status))
          } : {}, {
            e: common_vendor.o(($event) => remove(p.id), p.id),
            f: p.id
          });
        }),
        b: common_vendor.unref(store).photos.length < MAX
      }, common_vendor.unref(store).photos.length < MAX ? {
        c: common_vendor.o(add, "70")
      } : {}, {
        d: common_vendor.t(common_vendor.unref(store).photos.length),
        e: common_vendor.t(MAX)
      });
    };
  }
});
const MiniProgramPage = /* @__PURE__ */ common_vendor._export_sfc(_sfc_main, [["__scopeId", "data-v-02a64b7e"]]);
wx.createPage(MiniProgramPage);
