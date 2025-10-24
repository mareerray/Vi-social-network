<template>
	<div class="container">
		<div class="row justify-content-center">
			<div class="col-md-6 col-lg-5">
				<div class="card shadow-lg">
					<div class="card-body p-4">
						<div class="text-center mb-4">
							<i class="fas fa-user-plus fa-3x text-primary mb-3"></i>
							<h2 class="card-title text-primary fw-bold">Create Account</h2>
							<p class="text-muted">Join our social network today!</p>
						</div>

						<form @submit.prevent="onRegister">
							<!-- Required Fields -->
							<div class="row mb-3">
								<div class="col-sm-6">
									<label class="form-label fw-semibold">First Name *</label>
									<input v-model="first_name" 
										   type="text" 
										   class="form-control" 
										   placeholder="Enter first name"
										   required />
								</div>
								<div class="col-sm-6">
									<label class="form-label fw-semibold">Last Name *</label>
									<input v-model="last_name" 
										   type="text" 
										   class="form-control" 
										   placeholder="Enter last name"
										   required />
								</div>
							</div>

							<div class="mb-3">
								<label class="form-label fw-semibold">Email *</label>
								<input v-model="email" 
									   type="email" 
									   class="form-control" 
									   placeholder="Enter your email"
									   required />
							</div>

							<div class="mb-3">
								<label class="form-label fw-semibold">Password *</label>
								<input v-model="password" 
									   type="password" 
									   class="form-control" 
									   placeholder="Create a password"
									   required />
							</div>

							<!-- Optional Fields -->
							<div class="mb-3">
								<label class="form-label fw-semibold">Nickname</label>
								<input v-model="nickname" 
									   type="text" 
									   class="form-control" 
									   placeholder="Choose a nickname (optional)" />
								<small class="form-text text-muted">Leave blank to auto-generate from email</small>
							</div>

							<div class="mb-3">
								<label class="form-label fw-semibold">Date of Birth</label>
								<input v-model="date_of_birth" 
									   type="date" 
									   class="form-control" />
							</div>

							<div class="mb-3">
								<label class="form-label fw-semibold">Avatar</label>
								<input v-model="avatar"
									   type="url"
									   class="form-control"
									   placeholder="https://example.com/avatar.jpg" />
								<small class="form-text text-muted">Paste a public image URL for your avatar</small>
							</div>

							<div class="mb-3">
								<label class="form-label fw-semibold">About Me</label>
								<textarea v-model="about_me" 
										  class="form-control" 
										  rows="3"
										  placeholder="Tell us about yourself..."></textarea>
							</div>

							<div class="mb-4">
								<label class="form-label fw-semibold">Profile Type</label>
								<select v-model="profile_type" class="form-select">
									<option value="public">🌍 Public - Anyone can see my profile</option>
									<option value="private">🔒 Private - Only followers can see my profile</option>
								</select>
							</div>

							<div class="d-grid gap-2">
								<button type="submit" 
										class="btn btn-primary btn-lg">
									<i class="fas fa-user-plus me-2"></i>Create Account
								</button>
							</div>

							<div class="text-center mt-3">
								<p class="text-muted">Already have an account? 
									<router-link to="/login" class="text-primary text-decoration-none fw-semibold">
										Sign in here
									</router-link>
								</p>
							</div>
						</form>

						<div v-if="msg" class="alert alert-danger mt-3" role="alert">
							<i class="fas fa-exclamation-triangle me-2"></i>{{ msg }}
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import { ref } from 'vue'
import { useAuthStore } from '@/store/auth'
import { useRouter } from 'vue-router'
// avatar uploads/resize removed — users paste avatar URL only

export default {
	setup() {
		const email = ref('')
		const password = ref('')
		const first_name = ref('')
		const last_name = ref('')
		const msg = ref('')
		const auth = useAuthStore()
		const router = useRouter()

		const onRegister = async () => {
			try {
				// send keys matching backend json tags (about_me instead of about)
				await auth.register({
					nickname: nickname.value,
					email: email.value,
					password: password.value,
					first_name: first_name.value,
					last_name: last_name.value,
					date_of_birth: date_of_birth.value,
					avatar: avatar.value,
					about_me: about_me.value,
					profile_type: profile_type.value,
				})
				// avatar is provided as a URL by the user (no upload)
				// fetch user and go home
				await auth.fetchUser()
				router.push('/')
			} catch (e) {
				msg.value = e?.response?.data?.error || 'Registration failed'
			}
		}

		// Initialize refs for all form fields
		const nickname = ref('')
		const date_of_birth = ref('')
		const avatar = ref('')
		const about_me = ref('')
		const profile_type = ref('public')

		return { email, password, first_name, last_name, nickname, date_of_birth, avatar, about_me, profile_type, onRegister, msg }
	}
}
</script>

<style scoped>
.card {
  border: none;
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

.form-control, .form-select {
  border-radius: 0.75rem;
  border: 2px solid #e2e8f0;
  transition: all 0.3s ease;
}

.form-control:focus, .form-select:focus {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  transform: translateY(-1px);
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.75rem;
  font-weight: 600;
  padding: 12px 24px;
  transition: all 0.3s ease;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(102, 126, 234, 0.3);
}

.text-primary {
  color: #667eea !important;
}

.alert {
  border-radius: 0.75rem;
  border: none;
}
</style>

