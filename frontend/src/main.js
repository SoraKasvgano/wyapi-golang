import { createApp } from 'vue'
// 按需导入Element Plus组件
import {
  ElButton,
  ElInput,
  ElSelect,
  ElOption,
  ElForm,
  ElFormItem,
  ElCard,
  ElContainer,
  ElHeader,
  ElMain,
  ElFooter,
  ElConfigProvider,
  ElIcon,
  ElTag,
  ElLink,
  ElText,
  ElDivider,
  ElSpace,
  ElDialog,
  ElDrawer,
  ElAlert,
  ElProgress,
  ElRadio,
  ElRadioGroup,
  ElSwitch,
  ElMessage,
  ElTabs,
  ElTabPane,
  ElPagination
} from 'element-plus'

// 按需导入Element Plus样式，减少CSS体积
import 'element-plus/es/components/button/style/css'
import 'element-plus/es/components/input/style/css'
import 'element-plus/es/components/select/style/css'
import 'element-plus/es/components/option/style/css'
import 'element-plus/es/components/form/style/css'
import 'element-plus/es/components/form-item/style/css'
import 'element-plus/es/components/card/style/css'
import 'element-plus/es/components/container/style/css'
import 'element-plus/es/components/header/style/css'
import 'element-plus/es/components/main/style/css'
import 'element-plus/es/components/footer/style/css'
import 'element-plus/es/components/config-provider/style/css'
import 'element-plus/es/components/icon/style/css'
import 'element-plus/es/components/tag/style/css'
import 'element-plus/es/components/link/style/css'
import 'element-plus/es/components/text/style/css'
import 'element-plus/es/components/divider/style/css'
import 'element-plus/es/components/space/style/css'
import 'element-plus/es/components/dialog/style/css'
import 'element-plus/es/components/drawer/style/css'
import 'element-plus/es/components/loading/style/css'
import 'element-plus/es/components/alert/style/css'
import 'element-plus/es/components/progress/style/css'
import 'element-plus/es/components/radio/style/css'
import 'element-plus/es/components/radio-group/style/css'
import 'element-plus/es/components/switch/style/css'
import 'element-plus/es/components/notification/style/css'
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/tabs/style/css'
import 'element-plus/es/components/tab-pane/style/css'
import 'element-plus/es/components/pagination/style/css'
import 'element-plus/es/components/tour/style/css'

// 导入深色主题变量
import 'element-plus/theme-chalk/dark/css-vars.css'

// 全局样式与主题变量
import './styles/global.css'


import App from './App.vue'

const app = createApp(App)

// 注册Element Plus组件
app.component('ElButton', ElButton)
app.component('ElInput', ElInput)
app.component('ElSelect', ElSelect)
app.component('ElOption', ElOption)
app.component('ElForm', ElForm)
app.component('ElFormItem', ElFormItem)
app.component('ElCard', ElCard)
app.component('ElContainer', ElContainer)
app.component('ElHeader', ElHeader)
app.component('ElMain', ElMain)
app.component('ElFooter', ElFooter)
app.component('ElConfigProvider', ElConfigProvider)
app.component('ElIcon', ElIcon)
app.component('ElTag', ElTag)
app.component('ElLink', ElLink)
app.component('ElText', ElText)
app.component('ElDivider', ElDivider)
app.component('ElSpace', ElSpace)
app.component('ElDialog', ElDialog)
app.component('ElDrawer', ElDrawer)
app.component('ElAlert', ElAlert)
app.component('ElProgress', ElProgress)
app.component('ElRadio', ElRadio)
app.component('ElRadioGroup', ElRadioGroup)
app.component('ElSwitch', ElSwitch)
app.component('ElTabs', ElTabs)
app.component('ElTabPane', ElTabPane)
app.component('ElPagination', ElPagination)

// 取消全局注册图标，改为各组件内按需导入

// 全局配置ElMessage
app.config.globalProperties.$message = ElMessage

app.mount('#app')
