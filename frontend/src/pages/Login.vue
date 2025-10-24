<template>
	<div class="container py-5">
		<div class="row justify-content-center">
			<div class="col-md-6 col-lg-5">
				<div class="card shadow-lg">
					<div class="card-body p-5">
						<!-- Header -->
						<div class="text-center mb-4">
							<i class="fas fa-user-circle fa-3x text-primary mb-3"></i>
							<h2 class="h4 text-dark mb-2">Welcome Back!</h2>
							<p class="text-muted">Sign in to your account</p>
						</div>

						<!-- Error Alert -->
						<div v-if="msg" class="alert alert-danger d-flex align-items-center mb-4" role="alert">
							<i class="fas fa-exclamation-triangle me-2"></i>
							{{ msg }}
						</div>

						<!-- Login Form -->
						<form @submit.prevent="onLogin">
							<div class="mb-4">
								<label for="identifier" class="form-label fw-semibold">
									<i class="fas fa-envelope text-primary me-2"></i>
									Email or Username
								</label>
								<input 
									id="identifier"
									v-model="identifier" 
									type="text"
									class="form-control form-control-lg"
									placeholder="Enter your email or username"
									required
								/>
							</div>

							<div class="mb-4">
								<label for="password" class="form-label fw-semibold">
									<i class="fas fa-lock text-primary me-2"></i>
									Password
								</label>
								<input 
									id="password"
									v-model="password" 
									type="password"
									class="form-control form-control-lg"
									placeholder="Enter your password"
									required
								/>
							</div>

							<!-- Login Button -->
							<div class="d-grid mb-4">
								<button 
									type="submit" 
									class="btn btn-primary btn-lg"
									:disabled="!identifier || !password"
								>
									<i class="fas fa-sign-in-alt me-2"></i>
									Sign In
								</button>
							</div>
						</form>

						<!-- Register Link -->
						<div class="text-center">
							<p class="text-muted mb-0">
								Don't have an account? 
								<router-link to="/register" class="text-primary fw-semibold text-decoration-none">
									Sign up here
								</router-link>
							</p>
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

export default {
	setup() {
		const identifier = ref('')
		const password = ref('')
		const msg = ref('')
		const auth = useAuthStore()
		const router = useRouter()

		const onLogin = async () => {
			try {
				await auth.login(identifier.value, password.value)
				router.push('/')
			} catch (e) {
				msg.value = e?.response?.data || 'Login failed'
			}
		}

		return { identifier, password, onLogin, msg }
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

.form-control {
  border-radius: 0.75rem;
  border: 2px solid #e2e8f0;
  transition: all 0.3s ease;
}

.form-control:focus {
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

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  background: #6c757d;
  opacity: 0.6;
}

.text-primary {
  color: #667eea !important;
}

.alert {
  border-radius: 0.75rem;
  border: none;
}

router-link {
  transition: all 0.3s ease;
}

router-link:hover {
  text-decoration: underline !important;
}
</style>

