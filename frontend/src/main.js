import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { useAuthStore } from './store/auth'
import { useChatStore } from './store/chat'

// Bootstrap (CSS + JS bundle)
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

// bootstrap: check for existing session and auto-connect chat
app.mount('#app')

// run post-mount tasks
const auth = useAuthStore()
const chat = useChatStore()
;(async () => {
	await auth.fetchUser()
	if (auth.user) {
		// auto-connect websocket when a user session is present
		chat.connect()
	}
})()

