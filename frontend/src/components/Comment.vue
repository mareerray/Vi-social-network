<template>
  <div class="comment-box">
    <!-- existing comments -->
    <div v-if="comments.length > 0" class="mb-2">
      <div
        v-for="c in comments"
        :key="c.id"
        class="border-bottom pb-1 mb-1"
      >
        <strong>{{ c.nickname }}</strong>
        <span class="text-muted small ms-1">
          {{ new Date(c.created_at).toLocaleString() }}
        </span>
        <div>{{ c.content }}</div>
      </div>
    </div>

    <!-- new comment form -->
    <input
      v-model="text"
      placeholder="Write a comment..."
      class="form-control"
    />
    <input
      type="file"
      @change="onFile"
      accept="image/*"
      class="form-control mt-1"
      ref="fileInput"
    />
    <button class="btn btn-sm btn-primary mt-1" @click="submit">
      Comment
    </button>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { uploadFile } from '@/api'
import { addGroupComment, listGroupComments } from '@/api/groups'
import * as postApi from '@/api/post'

export default {
  props: {
    postId: { type: [Number, String], required: true },
    isGroup: { type: Boolean, default: false },
  },
  emits: ['comment-added'],
  setup(props, { emit }) {
    const text = ref('')
    const file = ref(null)
    const fileInput = ref(null)
    const comments = ref([])

    const onFile = (e) => {
      file.value = e.target.files[0]
    }

    const loadComments = async () => {
      if (!props.isGroup) return
      try {
        const res = await listGroupComments(props.postId)
        comments.value = Array.isArray(res.data) ? res.data : []
      } catch (e) {
        console.error('Failed to load comments:', e)
      }
    }

    const submit = async () => {
      if (!text.value && !file.value) return
      try {
        let imageUrl = ''

        if (file.value) {
          const uploaded = await uploadFile(file.value, 'post')
          imageUrl = uploaded.url
        }

        if (props.isGroup) {
          await addGroupComment({
            post_id: props.postId,
            content: text.value,
            image_url: imageUrl,
          })
          await loadComments()
        } else {
          await postApi.addComment(props.postId, text.value, imageUrl)
        }

        text.value = ''
        file.value = null
        if (fileInput.value) fileInput.value.value = ''
        emit('comment-added')
      } catch (e) {
        console.error('Failed to add comment:', e)
      }
    }

    onMounted(loadComments)
    return { text, file, comments, fileInput, onFile, submit }
  },
}
</script>

<style scoped>
.comment-box input {
  display: block;
  margin-bottom: 6px;
}
</style>
