<template>
  <div class="card create-post-card">
    <div class="card-header bg-gradient">
      <h5 class="mb-0 text-white">
        <i class="fas fa-edit me-2"></i>
        Create New Post
      </h5>
    </div>
    <div class="card-body">
      <form @submit.prevent="submit">
        <!-- Content Input -->
        <div class="mb-3">
          <div class="input-group">
            <span class="input-group-text bg-light border-0">
              <i class="fas fa-comment-dots text-primary"></i>
            </span>
            <textarea 
              v-model="content" 
              class="form-control border-0 shadow-sm" 
              placeholder="What's on your mind? Share something interesting..." 
              rows="4"
              required
            ></textarea>
          </div>
        </div>

        <!-- File Upload -->
        <div class="mb-3">
          <label class="form-label fw-semibold">
            <i class="fas fa-image text-primary me-2"></i>
            Add an Image
          </label>
          <input 
            type="file" 
            @change="onFile" 
            accept="image/*" 
            class="form-control" 
            ref="fileInput"
          />
          <div v-if="file" class="mt-2">
            <small class="text-success">
              <i class="fas fa-check-circle me-1"></i>
              {{ file.name }} selected
            </small>
          </div>
        </div>

        <div class="row">
          <!-- Privacy Setting -->
          <div class="col-md-6 mb-3">
            <label class="form-label fw-semibold">
              <i class="fas fa-shield-alt text-primary me-2"></i>
              Privacy
            </label>
            <select v-model="privacy" class="form-select">
              <option value="public">
                 Public
              </option>
              <option value="followers">
                 Followers Only
              </option>
              <option value="private">
                 Private
              </option>
            </select>
          </div>

          <!-- Allowed Users -->
          <div class="col-md-6 mb-3">
            <label class="form-label fw-semibold">
              <i class="fas fa-user-friends text-primary me-2"></i>
              Specific Users <small class="text-muted">(optional)</small>
            </label>
            <input 
              v-model="allowed" 
              placeholder="Enter user IDs separated by commas" 
              class="form-control"
            />
          </div>
        </div>

        <!-- Submit Button -->
        <div class="d-flex justify-content-between align-items-center">
          <div class="text-muted">
            <small>
              <i class="fas fa-info-circle me-1"></i>
              Your post will be visible based on privacy settings
            </small>
          </div>
          <button 
            class="btn btn-primary btn-lg px-4" 
            type="submit" 
            :disabled="!content.trim()"
          >
            <i class="fas fa-paper-plane me-2"></i>
            Share Post
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import * as postApi from '@/api/post'
import { uploadFile } from '@/api'

export default {
  emits: ['post-created'],
  setup(props, { emit }) {
    const content = ref('')
    const file = ref(null)
    const privacy = ref('public')
    const allowed = ref('')
    const fileInput = ref(null)

    function onFile(e) {
      file.value = e.target.files[0]
    }

    async function submit() {
      if (!content.value.trim()) return
      
      try {
        let imageUrl = '';
        if (file.value) {
          const uploadResponse = await uploadFile(file.value, 'post');
          imageUrl = uploadResponse.url;
        }

        const postData = {
          content: content.value,
          privacy: privacy.value,
          allowed: allowed.value,
          image_url: imageUrl,
        }
        
        await postApi.createPost(postData)
        
        // Reset form
        content.value = ''
        file.value = null
        privacy.value = 'public'
        allowed.value = ''
        if (fileInput.value) fileInput.value.value = ''
        
        // Emit event so parent can refresh the feed
        emit('post-created')
      } catch (error) {
        console.error('Error creating post:', error)
        // You could add error handling UI here
      }
    }

    return { content, file, privacy, allowed, fileInput, onFile, submit }
  }
}
</script>

<style scoped>
.create-post-card {
  border: none;
  border-radius: 1rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  overflow: hidden;
}

.create-post-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(0, 0, 0, 0.15);
}

.card-header.bg-gradient {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border: none;
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

.input-group-text {
  border-radius: 0.75rem 0 0 0.75rem;
  border: 2px solid #e2e8f0;
  border-right: none;
}

.input-group .form-control {
  border-radius: 0 0.75rem 0.75rem 0;
  border-left: none;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 0.75rem;
  font-weight: 600;
  transition: all 0.3s ease;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  background: #6c757d;
  opacity: 0.6;
  transform: none;
}

.text-primary {
  color: #667eea !important;
}

.text-success {
  color: #10b981 !important;
}

.form-label {
  color: #374151;
  margin-bottom: 0.5rem;
}
</style>
