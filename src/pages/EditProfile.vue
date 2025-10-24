<template>
  <div class="page card">
    <h2>Edit Profile</h2>
    <div v-if="loaded">
      <input v-model="form.nickname" placeholder="Nickname" />
      <input v-model="form.first_name" placeholder="First name" />
      <input v-model="form.last_name" placeholder="Last name" />
      <div style="display:flex; gap:8px; align-items:center">
        <input v-model="form.avatar" placeholder="Avatar URL" />
      </div>
      <textarea v-model="form.about" placeholder="About"></textarea>
      <label>Profile type:
        <select v-model="form.profile_type">
          <option value="public">Public</option>
          <option value="private">Private</option>
        </select>
      </label>
      <div style="margin-top:8px">
        <button @click="save">Save</button>
        <button @click="cancel">Cancel</button>
      </div>
      <div class="msg" v-if="msg">{{ msg }}</div>
    </div>
    <div v-else>Loading...</div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import * as api from '@/api/users'
import { useRouter } from 'vue-router'

export default {
  setup() {
    const form = ref({ nickname:'', first_name:'', last_name:'', avatar:'', about:'', profile_type:'public' })
    const loaded = ref(false)
    const msg = ref('')
    const router = useRouter()

    const load = async () => {
      try {
        const p = await api.getProfile()
        form.value.nickname = p.nickname || ''
        form.value.first_name = p.first_name || ''
        form.value.last_name = p.last_name || ''
        form.value.avatar = p.avatar || ''
        form.value.about = p.about || ''
        form.value.profile_type = p.profile_type || 'public'
      } catch (e) {
        msg.value = 'Failed to load profile'
      } finally {
        loaded.value = true
      }
    }

    const save = async () => {
      msg.value = ''
      try {
        const payload = {
          nickname: form.value.nickname,
          first_name: form.value.first_name,
          last_name: form.value.last_name,
          avatar: form.value.avatar,
          about: form.value.about,
          profile_type: form.value.profile_type
        }
        await api.updateProfile(payload)
        router.push(`/profile`)
      } catch (e) {
        msg.value = 'Failed to update profile'
      }
    }

    // avatar is provided as a URL

    const cancel = () => router.back()

    onMounted(load)
    return { form, loaded, msg, save, cancel }
  }
}
</script>

<style scoped>
.page { max-width:640px; margin:24px auto }
input, textarea, select { display:block; width:100%; margin:8px 0; padding:8px }
</style>
