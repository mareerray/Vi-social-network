<template>
  <div class="container py-4">
    <div class="row g-4">
      <!-- Main Content -->
      <div class="col-lg-8">
        <div class="mb-4">
          <CreatePost @post-created="loadPosts" />
        </div>
        
        <!-- Posts Feed -->
        <div class="posts-feed">
          <div v-if="posts.length === 0" class="text-center py-5">
            <i class="fas fa-newspaper fa-3x text-muted mb-3"></i>
            <h5 class="text-muted">No posts yet</h5>
            <p class="text-muted">Be the first to share something!</p>
          </div>
          
          <div v-for="p in posts" :key="p.id" class="mb-4">
            <PostCard :post="p" @comment-added="loadPosts" />
          </div>
        </div>
      </div>
      
      <!-- Sidebar -->
      <div class="col-lg-4">
        <div class="sidebar-content">
          <!-- Quick Stats -->
          <div class="card mb-4">
            <div class="card-header bg-gradient">
              <h6 class="mb-0 text-white">
                <i class="fas fa-chart-line me-2"></i>
                Activity Overview
              </h6>
            </div>
            <div class="card-body">
              <div class="d-flex justify-content-between align-items-center mb-2">
                <span class="text-muted">Your Posts</span>
                <span class="badge bg-primary">{{ userPostsCount }}</span>
              </div>
              <div class="d-flex justify-content-between align-items-center mb-2">
                <span class="text-muted">Total Posts</span>
                <span class="badge bg-success">{{ posts.length }}</span>
              </div>
            </div>
          </div>

          <!-- Trending Topics -->
          <div class="card">
            <div class="card-header bg-gradient">
              <h6 class="mb-0 text-white">
                <i class="fas fa-fire me-2"></i>
                What's Happening
              </h6>
            </div>
            <div class="card-body">
              <div class="trending-item mb-3">
                <small class="text-muted d-block">Trending in Social</small>
                <strong>#SocialNetwork</strong>
                <small class="text-muted d-block">{{ posts.length }} posts</small>
              </div>
              <div class="trending-item mb-3">
                <small class="text-muted d-block">Technology</small>
                <strong>#TechTalk</strong>
                <small class="text-muted d-block">{{ Math.floor(posts.length / 2) }} posts</small>
              </div>
              <div class="trending-item">
                <small class="text-muted d-block">Community</small>
                <strong>#Community</strong>
                <small class="text-muted d-block">{{ Math.floor(posts.length / 3) }} posts</small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import * as postApi from '@/api/post'
import CreatePost from '@/components/CreatePost.vue'
import PostCard from '@/components/PostCard.vue'

export default {
  components: { CreatePost, PostCard },
  setup() {
    const posts = ref([])
    const userPostsCount = ref(0)
    
    const loadPosts = async () => { 
      try {
        const postData = await postApi.listPosts()
        posts.value = postData || []
        // Count user's posts (this would ideally come from an API)
        userPostsCount.value = Math.floor((posts.value || []).length * 0.3) // Mock calculation
      } catch (error) {
        console.error("Failed to load posts:", error)
        posts.value = []
        userPostsCount.value = 0
      }
    }
    
    onMounted(loadPosts)
    return { posts, loadPosts, userPostsCount }
  }
}
</script>

<style scoped>
.container { 
  max-width: 1200px; 
}

.posts-feed {
  min-height: 400px;
}

.sidebar-content {
  position: sticky;
  top: 100px;
}

.card {
  border: none;
  border-radius: 1rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 15px rgba(0, 0, 0, 0.15);
}

.card-header.bg-gradient {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  border-radius: 1rem 1rem 0 0 !important;
  border: none;
}

.trending-item {
  padding: 8px 0;
  border-bottom: 1px solid #f1f3f4;
  cursor: pointer;
  transition: all 0.2s ease;
}

.trending-item:hover {
  background-color: #f8f9fa;
  border-radius: 0.5rem;
  padding: 8px 12px;
  margin: 0 -12px;
}

.trending-item:last-child {
  border-bottom: none;
}

.badge {
  font-weight: 500;
  padding: 6px 10px;
}
</style>
