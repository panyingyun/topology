import { createApp } from 'vue'
import naive from 'naive-ui'
import VXETable from 'vxe-table'
import 'vxe-table/lib/style.css'
import App from './App.vue'
import './style.css'

const app = createApp(App)
app.use(naive)
app.use(VXETable)
app.mount('#app')
