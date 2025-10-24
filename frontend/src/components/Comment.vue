<template>
	<div class="comment-box">
		<input v-model="text" placeholder="Write a comment..." class="form-control" />
		<input type="file" @change="onFile" accept="image/*" class="form-control mt-1" ref="fileInput" />
		<button class="btn btn-sm btn-primary mt-1" @click="submit">Comment</button>
	</div>
</template>

<script>
import { ref } from 'vue'
import * as postApi from '@/api/post'
import { uploadFile } from '@/api'

export default {
	props: { postId: { type: [Number, String], required: true } },
	emits: ['comment-added'],
	setup(props, { emit }) {
		const text = ref('')
		const file = ref(null)
		const fileInput = ref(null)

		const onFile = (e) => {
			file.value = e.target.files[0]
		}

		const submit = async () => {
			if (!text.value && !file.value) return
			try {
				let imageUrl = ''
				if (file.value) {
					const uploadResponse = await uploadFile(file.value, 'post')
					imageUrl = uploadResponse.url
				}
				await postApi.addComment(props.postId, text.value, imageUrl)
				text.value = ''
				file.value = null
				if (fileInput.value) fileInput.value.value = ''
				emit('comment-added')
			} catch (e) {
				// swallow for now
				console.error('Failed to add comment', e)
			}
		}
		return { text, file, fileInput, onFile, submit }
	}
}
</script>

<style scoped>
.comment-box input { display:block; margin-bottom:6px }
</style>
