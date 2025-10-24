<template>
  <div class="container mt-4">
    <h2>Groups</h2>
    <div class="mb-3">
      <input v-model="name" class="form-control" placeholder="Group name" />
      <textarea v-model="description" class="form-control mt-2" placeholder="Description"></textarea>
      <button class="btn btn-primary mt-2" @click="create">Create</button>
    </div>
    <div class="list-group">
      <a v-for="g in groups" :key="g.id" :href="`/group/${g.id}`" class="list-group-item list-group-item-action">
        <h5>{{ g.name }}</h5>
        <p class="mb-0">{{ g.description }}</p>
      </a>
    </div>
  </div>
</template>

<script>
import { listGroups, createGroup } from '../api/groups'

export default {
  data() {
    return { groups: [], name: '', description: '' }
  },
  created() {
    listGroups().then(r => { this.groups = r.data })
  },
  methods: {
    create() {
      createGroup({ name: this.name, description: this.description }).then(() => { this.name=''; this.description=''; return listGroups() }).then(r=> this.groups = r.data)
    }
  }
}
</script>

<style scoped>
</style>
