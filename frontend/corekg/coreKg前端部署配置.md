# 使用配置

从`src/utils/useDeployConfig`获取配置信息,从`window.__LANG`获取语言信息.

# 配置位置

`index.html`在其他 js 代码之前引入`public/config`和`public/lang.js`.

`public/config`已存在一些预设值,其类型应当始终与`src/utils/useDeployConfig/type`保持一致.后续这项配置可能由后端注入.

`lang.js`在`window`上设置语言信息,`window.__LANG`是一个预期不会变更的运行时变量,仅在镜像构建完成后由运维调整.它被用于`i18n`的初始化,并作为`Accept-Language`传输给后端以得到语言正确的报错信息.
